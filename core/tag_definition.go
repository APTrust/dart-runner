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
var TagsSetBySystem = []string{
	"Bagging-Date",
	"Bagging-Software",
	"Payload-Oxum",
	"Bag-Size",
	"BagIt-Profile-Identifier",
}

// TagDefinition describes a tag in a BagItProfile, whether it's
// required, what values are allowed, etc.
type TagDefinition struct {
	DefaultValue    string            `json:"defaultValue"`
	EmptyOK         bool              `json:"emptyOK"`
	Errors          map[string]string `json:"-"`
	Help            string            `json:"help"`
	ID              string            `json:"id"`
	IsBuiltIn       bool              `json:"isBuiltIn"`
	IsUserAddedFile bool              `json:"isUserAddedFile"`
	IsUserAddedTag  bool              `json:"isUserAddedTag"`
	Required        bool              `json:"required"`
	TagFile         string            `json:"tagFile"`
	TagName         string            `json:"tagName"`
	UserValue       string            `json:"userValue"`
	Values          []string          `json:"values"`
	WasAddedForJob  bool              `json:"wasAddedForJob"`
}

// SystemMustSet returns true if the value for this tag must be
// set by the bagger at runtime. This applies to tags like Bag-Size,
// Payload-Oxum, Bagging-Software and a few others in the bag-info.txt
// file. See TagsSetBySystem for a full list.
func (t *TagDefinition) SystemMustSet() bool {
	return t.TagFile == "bag-info.txt" && util.StringListContains(TagsSetBySystem, t.TagName)
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
		DefaultValue:    t.DefaultValue,
		EmptyOK:         t.EmptyOK,
		Help:            t.Help,
		ID:              t.ID,
		IsBuiltIn:       t.IsBuiltIn,
		IsUserAddedFile: t.IsUserAddedFile,
		IsUserAddedTag:  t.IsUserAddedTag,
		Required:        t.Required,
		TagFile:         t.TagFile,
		TagName:         t.TagName,
		UserValue:       t.UserValue,
		Values:          make([]string, len(t.Values)),
	}
	if t.Errors != nil {
		copyOfTagDef.Errors = util.CopyMap[string, string](t.Errors)
	}
	copy(copyOfTagDef.Values, t.Values)
	return copyOfTagDef
}

func (t *TagDefinition) Validate() bool {
	t.Errors = make(map[string]string)
	if util.IsEmpty(t.TagFile) {
		t.Errors["TagFile"] = "You must specify a tag file."
	}
	if util.IsEmpty(t.TagName) {
		t.Errors["TagName"] = "You must specify a tag name."
	}
	if !util.IsEmptyStringList(t.Values) {
		if !util.IsEmpty(t.DefaultValue) && !util.StringListContains(t.Values, t.DefaultValue) {
			t.Errors["DefaultValue"] = "The default value must be one of the allowed values."
		}
		if !util.IsEmpty(t.UserValue) && !util.StringListContains(t.Values, t.UserValue) {
			t.Errors["UserValue"] = "The value must be one of the allowed values."
		}
	}
	return len(t.Errors) == 0
}

func (t *TagDefinition) ToForm() *Form {
	form := NewForm(constants.TypeTagDefinition, t.ID, nil)
	form.UserCanDelete = !t.IsBuiltIn

	form.AddField("ID", "ID", t.ID, true)
	form.AddField("IsBuiltIn", "IsBuiltIn", strconv.FormatBool(t.IsBuiltIn), true)
	form.AddField("IsUserAddedFile", "IsUserAddedFile", strconv.FormatBool(t.IsUserAddedFile), true)
	form.AddField("IsUserAddedTag", "IsUserAddedTag", strconv.FormatBool(t.IsUserAddedTag), true)

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

	for field, errMsg := range t.Errors {
		form.Fields[field].Error = errMsg
	}

	return form
}

// // UnmarshalJson will unmarshal both current and malformed legacy (DART 2.x)
// // TagDefinition structs into a current TagDefinition object.
// // In some cases in DART 2.x, if a used defined a custom TagDefinition,
// // empty and single-value Values property from the input form were saved as
// // a string instead of an array of strings. This unmarshaler handles that case.
// func (t *TagDefinition) UnmarshalJSON(data []byte) error {
// 	var currentStruct struct {
// 		DefaultValue    string            `json:"defaultValue"`
// 		EmptyOK         bool              `json:"emptyOK"`
// 		Errors          map[string]string `json:"-"`
// 		Help            string            `json:"help"`
// 		ID              string            `json:"id"`
// 		IsBuiltIn       bool              `json:"isBuiltIn"`
// 		IsUserAddedFile bool              `json:"isUserAddedFile"`
// 		IsUserAddedTag  bool              `json:"isUserAddedTag"`
// 		Required        bool              `json:"required"`
// 		SystemMustSet   bool              `json:"systemMustSet"`
// 		TagFile         string            `json:"tagFile"`
// 		TagName         string            `json:"tagName"`
// 		UserValue       string            `json:"userValue"`
// 		Values          []string          `json:"values"`
// 		WasAddedForJob  bool              `json:"wasAddedForJob"`
// 	}

// 	var malformedLegacyStruct struct {
// 		DefaultValue    string            `json:"defaultValue"`
// 		EmptyOK         bool              `json:"emptyOK"`
// 		Errors          map[string]string `json:"-"`
// 		Help            string            `json:"help"`
// 		ID              string            `json:"id"`
// 		IsBuiltIn       bool              `json:"isBuiltIn"`
// 		IsUserAddedFile bool              `json:"isUserAddedFile"`
// 		IsUserAddedTag  bool              `json:"isUserAddedTag"`
// 		Required        bool              `json:"required"`
// 		SystemMustSet   bool              `json:"systemMustSet"`
// 		TagFile         string            `json:"tagFile"`
// 		TagName         string            `json:"tagName"`
// 		UserValue       string            `json:"userValue"`
// 		Values          string            `json:"values"`
// 		WasAddedForJob  bool              `json:"wasAddedForJob"`
// 	}

// 	// If this works, go with it.
// 	err := json.Unmarshal(data, &currentStruct)
// 	if err == nil {
// 		t.DefaultValue = currentStruct.DefaultValue
// 		t.EmptyOK = currentStruct.EmptyOK
// 		t.Errors = currentStruct.Errors
// 		t.Help = currentStruct.Help
// 		t.ID = currentStruct.ID
// 		t.IsBuiltIn = currentStruct.IsBuiltIn
// 		t.IsUserAddedFile = currentStruct.IsUserAddedFile
// 		t.IsUserAddedTag = currentStruct.IsUserAddedTag
// 		t.Required = currentStruct.Required
// 		t.SystemMustSet = currentStruct.SystemMustSet
// 		t.TagFile = currentStruct.TagFile
// 		t.TagName = currentStruct.TagName
// 		t.UserValue = currentStruct.UserValue
// 		t.Values = currentStruct.Values
// 		t.WasAddedForJob = currentStruct.WasAddedForJob
// 		return nil
// 	}

// 	// Otherwise, we're dealing with a malformed struct.
// 	err = json.Unmarshal(data, &malformedLegacyStruct)
// 	if err == nil {
// 		t.DefaultValue = malformedLegacyStruct.DefaultValue
// 		t.EmptyOK = malformedLegacyStruct.EmptyOK
// 		t.Errors = malformedLegacyStruct.Errors
// 		t.Help = malformedLegacyStruct.Help
// 		t.ID = malformedLegacyStruct.ID
// 		t.IsBuiltIn = malformedLegacyStruct.IsBuiltIn
// 		t.IsUserAddedFile = malformedLegacyStruct.IsUserAddedFile
// 		t.IsUserAddedTag = malformedLegacyStruct.IsUserAddedTag
// 		t.Required = malformedLegacyStruct.Required
// 		t.SystemMustSet = malformedLegacyStruct.SystemMustSet
// 		t.TagFile = malformedLegacyStruct.TagFile
// 		t.TagName = malformedLegacyStruct.TagName
// 		t.UserValue = malformedLegacyStruct.UserValue
// 		t.Values = []string{malformedLegacyStruct.Values}
// 		t.WasAddedForJob = malformedLegacyStruct.WasAddedForJob
// 		return err
// 	}

// 	return nil
// }
