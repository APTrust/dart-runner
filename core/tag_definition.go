package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

var reWhiteSpace = regexp.MustCompile(`[\s\r\n]+`)

// TagDefinition describes a tag in a BagItProfile, whether it's
// required, what values are allowed, etc.
type TagDefinition struct {
	DefaultValue  string   `json:"defaultValue"`
	EmptyOK       bool     `json:"emptyOK"`
	Help          string   `json:"help"`
	ID            string   `json:"id"`
	IsBuiltIn     bool     `json:"isBuiltIn"`
	Required      bool     `json:"required"`
	SystemMustSet bool     `json:"systemMustSet"`
	TagFile       string   `json:"tagFile"`
	TagName       string   `json:"tagName"`
	UserValue     string   `json:"userValue"`
	Values        []string `json:"values"`
}

// IsLegalValue returns true if val is a legal value for this tag definition.
// If TagDefinition.Values is empty, all values are legal.
func (t *TagDefinition) IsLegalValue(val string) bool {
	if t.Values == nil || len(t.Values) == 0 {
		return true
	}
	return util.StringListContains(t.Values, val)
}

// GetValue returns this tag's UserValue, if that's non-empty,
// or its DefaultValue.
func (t *TagDefinition) GetValue() string {
	val := t.UserValue
	if val == "" {
		val = t.DefaultValue
	}
	return val
}

// ToFormattedString returns the tag as string in a format suitable
// for writing to a tag file. Following LOC's bagit.py, this function
// does not break lines into 79 character chunks. It prints the whole
// tag on a single line, replacing newlines with spaces.
func (t *TagDefinition) ToFormattedString() string {
	cleanValue := reWhiteSpace.ReplaceAllString(t.GetValue(), " ")
	return fmt.Sprintf("%s: %s", t.TagName, strings.TrimSpace(cleanValue))
}

// Copy returns a pointer to a new TagDefinition whose values are the
// same as this TagDefinition.
func (t *TagDefinition) Copy() *TagDefinition {
	copyOfTagDef := &TagDefinition{
		DefaultValue: t.DefaultValue,
		EmptyOK:      t.EmptyOK,
		Help:         t.Help,
		ID:           t.ID,
		Required:     t.Required,
		TagFile:      t.TagFile,
		TagName:      t.TagName,
		UserValue:    t.UserValue,
		Values:       make([]string, len(t.Values)),
	}
	copy(copyOfTagDef.Values, t.Values)
	return copyOfTagDef
}

func (t *TagDefinition) ToForm() *Form {
	form := NewForm(constants.TypeTagDefinition, t.ID, nil)
	form.UserCanDelete = !t.IsBuiltIn

	form.AddField("ID", "ID", t.ID, true)
	form.AddField("IsBuiltIn", "IsBuiltIn", strconv.FormatBool(t.IsBuiltIn), true)

	helpField := form.AddField("Help", "Help Text", t.Help, false)
	helpField.Help = "(Optional) Describe the significance of this tag so users know what data to enter."

	tagNameField := form.AddField("TagName", "Tag Name", t.TagName, true)
	tagFileField := form.AddField("TagFile", "Tag File", t.TagFile, true)
	if t.IsBuiltIn {
		tagNameField.Attrs["readonly"] = "readonly"
		tagFileField.Attrs["readonly"] = "readonly"
	}

	// Allowed values will be displayed in a textarea, with one value per line.
	trimmedValues := make([]string, len(t.Values))
	for i, value := range t.Values {
		trimmedValues[i] = strings.TrimSpace(value)
	}
	valuesStr := strings.Join(trimmedValues, "\n")
	valuesField := form.AddField("Values", "Allowed Values", valuesStr, false)
	valuesField.Help = "If you want to restrict the values allowed for this tag, enter the allowed values here, one item per line."

	defaultValueField := form.AddField("DefaultValue", "Default Value", t.DefaultValue, false)
	if len(trimmedValues) > 0 {
		defaultValueField.Choices = MakeChoiceList(trimmedValues, strings.TrimSpace(t.DefaultValue))
	}

	requiredField := form.AddField("Required", "Required", strconv.FormatBool(t.Required), false)
	requiredField.Help = "Does this tag require a value?"
	requiredField.Choices = YesNoChoices(t.Required)

	return form
}
