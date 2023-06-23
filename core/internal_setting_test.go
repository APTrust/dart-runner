package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInternalSettingPersistence(t *testing.T) {

	// Clean up when test completes.
	defer core.ClearDartTable()

	// Insert three records for testing.
	s1 := core.NewInternalSetting("Setting 1", "Value 1")
	s2 := core.NewInternalSetting("Setting 2", "Value 2")
	s3 := core.NewInternalSetting("Setting 3", "Value 3")
	assert.Nil(t, core.ObjSave(s1))
	assert.Nil(t, core.ObjSave(s2))
	assert.Nil(t, core.ObjSave(s3))

	// Make sure S1 was saved as expected.
	result := core.ObjFind(s1.ID)
	require.Nil(t, result.Error)
	s1Reload := result.InternalSetting()
	require.NotNil(t, s1Reload)
	assert.Equal(t, s1.ID, s1Reload.ID)
	assert.Equal(t, s1.Name, s1Reload.Name)
	assert.Equal(t, s1.Value, s1Reload.Value)

	// Make sure order, offset and limit work on list query.
	result = core.ObjList(constants.TypeInternalSetting, "obj_name", 1, 0)
	require.Nil(t, result.Error)
	require.Equal(t, 1, len(result.InternalSettings))
	assert.Equal(t, s1.ID, result.InternalSettings[0].ID)

	// Make sure we can get all results.
	result = core.ObjList(constants.TypeInternalSetting, "obj_name", 100, 0)
	require.Nil(t, result.Error)
	require.Equal(t, 3, len(result.InternalSettings))
	assert.Equal(t, s1.ID, result.InternalSettings[0].ID)
	assert.Equal(t, s2.ID, result.InternalSettings[1].ID)
	assert.Equal(t, s3.ID, result.InternalSettings[2].ID)

	// We cannot delete internal settings. This should return an error.
	assert.NotNil(t, core.ObjDelete(s1))
}

func TestInternalSettingValidation(t *testing.T) {
	s1 := core.NewInternalSetting("", "")
	assert.False(t, s1.Validate())
	assert.Equal(t, "Name cannot be empty.", s1.Errors["Name"])
	assert.Equal(t, "Value cannot be empty.", s1.Errors["Value"])
	assert.Equal(t, constants.ErrObjecValidation, core.ObjSave(s1))

	s1.Name = "Setting 1 Name"
	assert.False(t, s1.Validate())
	assert.Equal(t, "", s1.Errors["Name"])
	assert.Equal(t, "Value cannot be empty.", s1.Errors["Value"])
	assert.Equal(t, constants.ErrObjecValidation, core.ObjSave(s1))

	s1.Value = "Setting 1 Value"
	assert.True(t, s1.Validate())
	assert.Equal(t, "", s1.Errors["Name"])
	assert.Equal(t, "", s1.Errors["Value"])
	assert.Nil(t, core.ObjSave(s1))

	result := core.ObjFind(s1.ID)
	assert.Nil(t, result.Error)
	require.NotNil(t, result.InternalSetting())
	assert.Equal(t, s1.Name, result.InternalSetting().Name)
}

func TestInternalSettingToForm(t *testing.T) {
	setting := core.NewInternalSetting("Setting 1", "Value 1")
	form := setting.ToForm()
	assert.Equal(t, 3, len(form.Fields))
	assert.False(t, form.UserCanDelete)
	assert.Equal(t, setting.ID, form.Fields["ID"].Value)
	assert.Equal(t, setting.Name, form.Fields["Name"].Value)
	assert.Equal(t, setting.Value, form.Fields["Value"].Value)

	assert.Equal(t, "readonly", form.Fields["Name"].Attrs["readonly"])
	assert.Equal(t, "readonly", form.Fields["Value"].Attrs["readonly"])
}

func TestInternalSettingPersistentObject(t *testing.T) {
	setting := core.NewInternalSetting("Setting 1", "Value 1")
	assert.Equal(t, constants.TypeInternalSetting, setting.ObjType())
	assert.Equal(t, "InternalSetting", setting.ObjType())
	assert.Equal(t, setting.ID, setting.ObjID())
	assert.True(t, util.LooksLikeUUID(setting.ObjID()))
	assert.False(t, setting.IsDeletable())
	assert.Equal(t, "Setting 1", setting.ObjName())
	assert.Equal(t, "InternalSetting: 'Setting 1' = 'Value 1'", setting.String())
	assert.Empty(t, setting.GetErrors())

	setting.Errors = map[string]string{
		"Error 1": "Message 1",
		"Error 2": "Message 2",
	}

	assert.Equal(t, 2, len(setting.GetErrors()))
	assert.Equal(t, "Message 1", setting.GetErrors()["Error 1"])
	assert.Equal(t, "Message 2", setting.GetErrors()["Error 2"])
}
