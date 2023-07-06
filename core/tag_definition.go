package core

import (
	"fmt"
	"regexp"
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

	// TODO: Finish implementing this & test

	return form
}
