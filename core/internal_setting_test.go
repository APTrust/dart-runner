package core_test

import (
	"database/sql"
	"testing"

	"github.com/APTrust/dart-runner/core"
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
	assert.Nil(t, s1.Save())
	assert.Nil(t, s2.Save())
	assert.Nil(t, s3.Save())

	// Make sure S1 was saved as expected.
	s1Reload, err := core.InternalSettingFind(s1.ID)
	require.Nil(t, err)
	require.NotNil(t, s1Reload)
	assert.Equal(t, s1.ID, s1Reload.ID)
	assert.Equal(t, s1.Name, s1Reload.Name)
	assert.Equal(t, s1.Value, s1Reload.Value)

	// Make sure order, offset and limit work on list query.
	settings, err := core.InternalSettingList("obj_name", 1, 0)
	require.Nil(t, err)
	require.Equal(t, 1, len(settings))
	assert.Equal(t, s1.ID, settings[0].ID)

	// Make sure we can get all results.
	settings, err = core.InternalSettingList("obj_name", 100, 0)
	require.Nil(t, err)
	require.Equal(t, 3, len(settings))
	assert.Equal(t, s1.ID, settings[0].ID)
	assert.Equal(t, s2.ID, settings[1].ID)
	assert.Equal(t, s3.ID, settings[2].ID)

	// Make sure delete works. Should return no error.
	assert.Nil(t, s1.Delete())

	// Make sure the record was truly deleted.
	deletedRecord, err := core.InternalSettingFind(s1.ID)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.Nil(t, deletedRecord)
}

func TestInternalSettingValidation(t *testing.T) {
	s1 := core.NewInternalSetting("", "")
	assert.False(t, s1.Validate())
	assert.Equal(t, "Name cannot be empty.", s1.Errors["Name"])
	assert.Equal(t, "Value cannot be empty.", s1.Errors["Value"])
	assert.Equal(t, core.ErrObjecValidation, s1.Save())

	s1.Name = "Setting 1 Name"
	assert.False(t, s1.Validate())
	assert.Equal(t, "", s1.Errors["Name"])
	assert.Equal(t, "Value cannot be empty.", s1.Errors["Value"])
	assert.Equal(t, core.ErrObjecValidation, s1.Save())

	s1.Value = "Setting 1 Value"
	assert.True(t, s1.Validate())
	assert.Equal(t, "", s1.Errors["Name"])
	assert.Equal(t, "", s1.Errors["Value"])
	assert.Nil(t, s1.Save())

	s1Reload, err := core.InternalSettingFind(s1.ID)
	assert.Nil(t, err)
	require.NotNil(t, s1Reload)
	assert.Equal(t, s1.Name, s1Reload.Name)
}