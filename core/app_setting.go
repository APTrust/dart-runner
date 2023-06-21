package core

import (
	"fmt"
	"strconv"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
	"github.com/google/uuid"
)

// AppSetting represents an application-wide setting that can be
// configured by the user. For example, the bagging directory
// into which DART writes new bags.
//
// Field names for JSON serialization match the old DART 2 names,
// so we don't break legacy installations.
type AppSetting struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Value         string            `json:"value"`
	Help          string            `json:"help"`
	Errors        map[string]string `json:"errors"`
	UserCanDelete bool              `json:"userCanDelete"`
}

// NewAppSetting creates a new AppSetting with the specified name
// and value. UserCanDelete will be true by default. If a setting
// is required for DART to function properly (such as the Bagging
// Directory setting), set UserCanDelete to false.
func NewAppSetting(name, value string) *AppSetting {
	return &AppSetting{
		ID:            uuid.NewString(),
		Name:          name,
		Value:         value,
		UserCanDelete: true,
		Errors:        make(map[string]string),
	}
}

// ObjID returns this setting's object id (uuid).
func (setting *AppSetting) ObjID() string {
	return setting.ID
}

// ObjName returns this object's name, so names will be
// searchable and sortable in the DB.
func (setting *AppSetting) ObjName() string {
	return setting.Name
}

// ObjType returns this object's type name.
func (setting *AppSetting) ObjType() string {
	return constants.TypeAppSetting
}

// ToForm returns a form so the user can edit this AppSetting.
// The form can be rendered by the app_setting/form.html template.
func (setting *AppSetting) ToForm() *Form {
	form := NewForm(constants.TypeAppSetting, setting.ID, setting.Errors)
	form.UserCanDelete = setting.UserCanDelete

	form.AddField("ID", "ID", setting.ID, true)

	_ = form.AddField("UserCanDelete", "UserCanDelete", strconv.FormatBool(setting.UserCanDelete), true)

	nameField := form.AddField("Name", "Name", setting.Name, true)
	// If user cannot delete this field, they can't rename it either.
	// Renaming the setting would prevent the app from finding it,
	// an in the case of a required setting like "Bagging Directory,"
	// that would cause lots of problems.
	if !setting.UserCanDelete {
		nameField.Attrs["readonly"] = "readonly"
	}

	valueField := form.AddField("Value", "Value", setting.Value, true)
	valueField.Help = "If the setting has help text, it will be displayed here." // setting.Help

	return form
}

// Validate validates this setting, returning true if it's valid,
// false if not. If false, this sets specific error messages in the
// Errors map, which are suitable for display on the form.
func (setting *AppSetting) Validate() bool {
	setting.Errors = make(map[string]string)
	isValid := true
	if !util.LooksLikeUUID(setting.ID) {
		setting.Errors["ID"] = "ID must be a valid uuid."
		isValid = false
	}
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

func (setting *AppSetting) String() string {
	return fmt.Sprintf("AppSetting: '%s' = '%s'", setting.Name, setting.Value)
}

func (setting *AppSetting) GetErrors() map[string]string {
	return setting.Errors
}

func (setting *AppSetting) IsDeletable() bool {
	return setting.UserCanDelete
}
