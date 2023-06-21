package core

import (
	"fmt"

	"github.com/APTrust/dart-runner/constants"
	"github.com/google/uuid"
)

// InternalSetting is set by DART and cannot be edited by user.
// These settings may record when migrations were run, or other
// internal info. These settings cannot be created or edited by
// users.
type InternalSetting struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Value  string            `json:"value"`
	Errors map[string]string `json:"errors"`
}

func NewInternalSetting(name, value string) *InternalSetting {
	return &InternalSetting{
		ID:     uuid.NewString(),
		Name:   name,
		Value:  value,
		Errors: make(map[string]string),
	}
}

func (setting *InternalSetting) ObjID() string {
	return setting.ID
}

func (setting *InternalSetting) ObjName() string {
	return setting.Name
}

func (setting *InternalSetting) ObjType() string {
	return constants.TypeInternalSetting
}

func (setting *InternalSetting) String() string {
	return fmt.Sprintf("InternalSetting: '%s' = '%s'", setting.Name, setting.Value)
}

// Validate validates this setting, returning true if it's valid,
// false if not. If false, this sets specific error messages in the
// Errors map, which are suitable for display on the form.
func (setting *InternalSetting) Validate() bool {
	setting.Errors = make(map[string]string)
	isValid := true
	if setting.Name == "" {
		setting.Errors["Name"] = "Name cannot be empty."
		isValid = false
	}
	if setting.Value == "" {
		setting.Errors["Value"] = "Value cannot be empty."
		isValid = false
	}
	return isValid
}

// ToForm returns a form to comply with the PersistentObject
// interface, but internal settings are not editable, so the
// app never displays this form.
func (setting *InternalSetting) ToForm() *Form {
	form := NewForm(constants.TypeAppSetting, setting.ID, setting.Errors)
	form.UserCanDelete = false

	form.AddField("ID", "ID", setting.ID, true)
	nameField := form.AddField("Name", "Name", setting.Name, true)
	nameField.Attrs["readonly"] = "readonly"

	valueField := form.AddField("Value", "Value", setting.Value, true)
	valueField.Attrs["readonly"] = "readonly"
	return form
}

func (setting *InternalSetting) GetErrors() map[string]string {
	return setting.Errors
}

func (setting *InternalSetting) IsDeletable() bool {
	return false
}
