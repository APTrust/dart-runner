package core_test

import (
	"database/sql"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppSettingPersistence(t *testing.T) {

	// Clean up when test completes.
	defer core.ClearDartTable()

	// Insert three records for testing.
	s1 := core.NewAppSetting("Setting 1", "Value 1")
	s2 := core.NewAppSetting("Setting 2", "Value 2")
	s3 := core.NewAppSetting("Setting 3", "Value 3")
	s3.UserCanDelete = false
	assert.Nil(t, core.ObjSave(s1))
	assert.Nil(t, core.ObjSave(s2))
	assert.Nil(t, core.ObjSave(s3))

	// Make sure S1 was saved as expected.
	result := core.ObjFind(s1.ID)
	require.Nil(t, result.Error)
	s1Reload := result.AppSetting()
	require.NotNil(t, s1Reload)
	assert.Equal(t, s1.ID, s1Reload.ID)
	assert.Equal(t, s1.Name, s1Reload.Name)
	assert.Equal(t, s1.Value, s1Reload.Value)

	// Make sure order, offset and limit work on list query.
	result = core.ObjList(constants.TypeAppSetting, "obj_name", 1, 0)
	require.Nil(t, result.Error)
	require.Equal(t, 1, len(result.AppSettings))
	assert.Equal(t, s1.ID, result.AppSettings[0].ID)

	// Make sure we can get all results.
	result = core.ObjList(constants.TypeAppSetting, "obj_name", 100, 0)
	require.Nil(t, result.Error)
	require.Equal(t, 3, len(result.AppSettings))
	assert.Equal(t, s1.ID, result.AppSettings[0].ID)
	assert.Equal(t, s2.ID, result.AppSettings[1].ID)
	assert.Equal(t, s3.ID, result.AppSettings[2].ID)

	// Make sure delete works. Should return no error.
	assert.Nil(t, core.ObjDelete(s1))

	// Make sure the record was truly deleted.
	result = core.ObjFind(s1.ID)
	assert.Equal(t, sql.ErrNoRows, result.Error)
	assert.Nil(t, result.AppSetting())

	// User should not be able to delete s3 because
	// s3.UserCanDelete = false.
	assert.Equal(t, constants.ErrNotDeletable, core.ObjDelete(s3))
}

func TestAppSettingValidation(t *testing.T) {
	// Post-test cleanup
	defer core.ClearDartTable()

	s1 := core.NewAppSetting("", "")
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
	require.NotNil(t, result.AppSetting())
	assert.Equal(t, s1.Name, result.AppSetting().Name)
}

func TestAppSettingToForm(t *testing.T) {
	setting := core.NewAppSetting("Setting 1", "Value 1")
	form := setting.ToForm()
	assert.Equal(t, 4, len(form.Fields))
	assert.True(t, form.UserCanDelete)
	assert.Equal(t, setting.ID, form.Fields["ID"].Value)
	assert.Equal(t, setting.Name, form.Fields["Name"].Value)
	assert.Equal(t, setting.Value, form.Fields["Value"].Value)

	assert.Empty(t, form.Fields["Name"].Attrs["readonly"])
	assert.True(t, form.Fields["Name"].Required)
	assert.True(t, form.Fields["Value"].Required)

	setting.UserCanDelete = false
	form = setting.ToForm()
	assert.False(t, form.UserCanDelete)
	assert.Equal(t, "readonly", form.Fields["Name"].Attrs["readonly"])
}

func TestAppSettingPersistentObject(t *testing.T) {
	setting := core.NewAppSetting("Setting 1", "Value 1")
	assert.Equal(t, constants.TypeAppSetting, setting.ObjType())
	assert.Equal(t, "AppSetting", setting.ObjType())
	assert.Equal(t, setting.ID, setting.ObjID())
	assert.True(t, util.LooksLikeUUID(setting.ObjID()))
	assert.True(t, setting.IsDeletable())
	assert.Equal(t, "Setting 1", setting.ObjName())
	assert.Equal(t, "AppSetting: 'Setting 1' = 'Value 1'", setting.String())
	assert.Empty(t, setting.GetErrors())

	setting.UserCanDelete = false
	setting.Errors = map[string]string{
		"Error 1": "Message 1",
		"Error 2": "Message 2",
	}

	assert.False(t, setting.IsDeletable())
	assert.Equal(t, 2, len(setting.GetErrors()))
	assert.Equal(t, "Message 1", setting.GetErrors()["Error 1"])
	assert.Equal(t, "Message 2", setting.GetErrors()["Error 2"])
}
