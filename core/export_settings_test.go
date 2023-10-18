package core_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExportSettings(t *testing.T) {
	s := core.NewExportSettings()
	assert.True(t, util.LooksLikeUUID(s.ID))
	assert.NotNil(t, s.AppSettings)
	assert.NotNil(t, s.BagItProfiles)
	assert.NotNil(t, s.Errors)
	assert.NotNil(t, s.Questions)
	assert.NotNil(t, s.RemoteRepositories)
	assert.NotNil(t, s.StorageServices)
}

func TestExportSettingsPersistentObjectInterface(t *testing.T) {
	defer core.ClearDartTable()
	s := core.NewExportSettings()
	s.Name = "Test Settings"

	assert.Equal(t, s.ID, s.ObjID())
	assert.Equal(t, s.Name, s.ObjName())
	assert.Equal(t, constants.TypeExportSettings, s.ObjType())
	assert.True(t, s.IsDeletable())
	assert.True(t, s.Validate())

	assert.NoError(t, core.ObjSave(s))
	result := core.ObjFind(s.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, s.ID, result.ExportSetting().ID)

	assert.NoError(t, core.ObjDelete(s))
	result = core.ObjFind(s.ID)
	assert.Equal(t, sql.ErrNoRows, result.Error)

	for i := 0; i < 3; i++ {
		settings := core.NewExportSettings()
		settings.Name = fmt.Sprintf("Settings %d", i)
		assert.NoError(t, core.ObjSave(settings))
	}

	result = core.ObjList(constants.TypeExportSettings, "obj_name", 10, 0)
	require.NoError(t, result.Error)
	items := result.ExportSettings
	require.Equal(t, 3, len(items))
	assert.Equal(t, "Settings 0", items[0].Name)
	assert.Equal(t, "Settings 1", items[1].Name)
	assert.Equal(t, "Settings 2", items[2].Name)
}

func TestExportSettingsToForm(t *testing.T) {

}

func TestExportSettingsObjectIds(t *testing.T) {
	s := core.NewExportSettings()
	appSettingIds := make([]string, 3)
	profileIds := make([]string, 3)
	repoIds := make([]string, 3)
	ssIds := make([]string, 3)
	for i := 0; i < 3; i++ {
		appSetting := core.NewAppSetting(fmt.Sprintf("Name %d", i), fmt.Sprintf("Value %d", i))
		appSettingIds[i] = appSetting.ID
		s.AppSettings = append(s.AppSettings, appSetting)

		profile := core.NewBagItProfile()
		profileIds[i] = profile.ID
		s.BagItProfiles = append(s.BagItProfiles, profile)

		repo := core.NewRemoteRepository()
		repoIds[i] = repo.ID
		s.RemoteRepositories = append(s.RemoteRepositories, repo)

		ss := core.NewStorageService()
		ssIds[i] = ss.ID
		s.StorageServices = append(s.StorageServices, ss)
	}

	ids, err := s.ObjectIds(constants.TypeAppSetting)
	require.NoError(t, err)
	assert.EqualValues(t, appSettingIds, ids)

	ids, err = s.ObjectIds(constants.TypeBagItProfile)
	require.NoError(t, err)
	assert.EqualValues(t, profileIds, ids)

	ids, err = s.ObjectIds(constants.TypeRemoteRepository)
	require.NoError(t, err)
	assert.EqualValues(t, repoIds, ids)

	ids, err = s.ObjectIds(constants.TypeStorageService)
	require.NoError(t, err)
	assert.EqualValues(t, ssIds, ids)

	ids, err = s.ObjectIds("type does not exist")
	assert.Equal(t, constants.ErrUnknownType, err)
	assert.Empty(t, ids)
}
