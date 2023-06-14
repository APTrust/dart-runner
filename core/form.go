package core

import "fmt"

// Choice represents a choice on an HTML form. A list of
// choices can be rendered as an HTML select list or a
// checkbox group or radio group.
type Choice struct {
	Label    string
	Value    string
	Selected bool
}

// Field represents a field in an HTML form.
type Field struct {
	ID             string
	Name           string
	Label          string
	Value          string
	Required       bool
	Help           string
	Hide           bool
	Error          string
	Choices        []Choice
	CssClasses     []string
	Attrs          map[string]string
	Changed        bool
	FormGroupClass string // set to "form-group-hidden" if SystemMustSet or tag has default value - https://github.com/APTrust/dart/blob/213e0a9acde2c73fcc8430ca436263b806228f8f/ui/forms/job_tags_form.js#L80
}

// NewField creates a new field for a form.
func NewField(id, name, label, value string, required bool) *Field {
	return &Field{
		ID:         id,
		Name:       name,
		Label:      label,
		Value:      value,
		Required:   required,
		Choices:    make([]Choice, 0),
		CssClasses: make([]string, 0),
		Attrs:      make(map[string]string),
	}
}

// AddChoice adds a choice to a field, and marks it as selected
// if it matches the field's value.
func (field *Field) AddChoice(label, value string) {
	choice := Choice{
		Label: label,
		Value: value,
	}
	choice.Selected = value == field.Value
	field.Choices = append(field.Choices, choice)
}

// Form represents an HTML form.
type Form struct {
	ObjType        string
	ObjectID       string
	Fields         map[string]*Field
	SubmitFunction string
	CancelFunction string
	DeleteFunction string
	UserCanDelete  bool
	Errors         map[string]string
}

// NewForm creates a new form for the specified object type and id.
// The errors map is used to set error messages on the form's fields.
func NewForm(objType, objectId string, errors map[string]string) *Form {
	if errors == nil {
		errors = make(map[string]string)
	}
	return &Form{
		ObjType:        objType,
		ObjectID:       objectId,
		Fields:         make(map[string]*Field),
		CancelFunction: fmt.Sprintf("%sList", objType),
		SubmitFunction: fmt.Sprintf("%sSave", objType),
		DeleteFunction: fmt.Sprintf("%sDelete", objType),
		Errors:         errors,
	}
}

// AddField adds a new field to the form.
func (f *Form) AddField(name, label, value string, required bool) *Field {
	id := fmt.Sprintf("%s_%s", f.ObjType, name)
	f.Fields[name] = NewField(id, name, label, value, required)
	f.Fields[name].Error = f.Errors[name]
	return f.Fields[name]
}

// String converts the form to a string for logging purposes.
func (f *Form) String() string {
	return fmt.Sprintf("Form: %s id=%s", f.ObjType, f.ObjectID)
}
