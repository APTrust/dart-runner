package core_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

type ExportSettingsTestData struct {
	ExportSettings    *core.ExportSettings
	AppSettingIds     []string
	ProfileIds        []string
	RepoIds           []string
	StorageServiceIds []string
}

func getExportSettingsTestData() ExportSettingsTestData {
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
		profile.Name = fmt.Sprintf("Profile %d", i)
		profileIds[i] = profile.ID
		s.BagItProfiles = append(s.BagItProfiles, profile)

		repo := core.NewRemoteRepository()
		repo.Name = fmt.Sprintf("Repo %d", i)
		repoIds[i] = repo.ID
		s.RemoteRepositories = append(s.RemoteRepositories, repo)

		ss := core.NewStorageService()
		ss.Name = fmt.Sprintf("Storage Service %d", i)
		ssIds[i] = ss.ID
		s.StorageServices = append(s.StorageServices, ss)
	}
	return ExportSettingsTestData{
		ExportSettings:    s,
		AppSettingIds:     appSettingIds,
		ProfileIds:        profileIds,
		RepoIds:           repoIds,
		StorageServiceIds: ssIds,
	}
}

func TestExportSettingsToForm(t *testing.T) {
	defer core.ClearDartTable()
	testData := getExportSettingsTestData()
	settings := testData.ExportSettings
	// Save the attached objects because they need to be in the DB
	// before we can build the form.
	for i := 0; i < 3; i++ {
		require.NoError(t, core.ObjSaveWithoutValidation(settings.AppSettings[i]))
		require.NoError(t, core.ObjSaveWithoutValidation(settings.BagItProfiles[i]))
		require.NoError(t, core.ObjSaveWithoutValidation(settings.RemoteRepositories[i]))
		require.NoError(t, core.ObjSaveWithoutValidation(settings.StorageServices[i]))
	}
	// Add some new objects that aren't part of the ExportSettings.
	// These should appear in the checkbox lists but not be selected.
	for i := 10; i < 13; i++ {
		appSetting := core.NewAppSetting(fmt.Sprintf("Name %d", i), fmt.Sprintf("Value %d", i))
		require.NoError(t, core.ObjSaveWithoutValidation(appSetting))

		profile := core.NewBagItProfile()
		profile.Name = fmt.Sprintf("Profile %d", i)
		require.NoError(t, core.ObjSaveWithoutValidation(profile))

		repo := core.NewRemoteRepository()
		repo.Name = fmt.Sprintf("Repo %d", i)
		require.NoError(t, core.ObjSaveWithoutValidation(repo))

		ss := core.NewStorageService()
		ss.Name = fmt.Sprintf("Storage Service %d", i)
		require.NoError(t, core.ObjSaveWithoutValidation(ss))
	}

	form := testData.ExportSettings.ToForm()
	require.NotNil(t, form)

	assert.Equal(t, settings.ID, form.Fields["ID"].Value)
	assert.Equal(t, settings.Name, form.Fields["Name"].Value)
	assert.Equal(t, 6, len(form.Fields["AppSettings"].Choices))
	assert.Equal(t, 6, len(form.Fields["BagItProfiles"].Choices))
	assert.Equal(t, 6, len(form.Fields["RemoteRepositories"].Choices))
	assert.Equal(t, 6, len(form.Fields["StorageServices"].Choices))

	// AppSettings
	selectedCount := 0
	notSelected := 0
	for _, choice := range form.Fields["AppSettings"].Choices {
		if util.StringListContains(testData.AppSettingIds, choice.Value) {
			assert.True(t, choice.Selected)
			selectedCount++
		} else {
			assert.False(t, choice.Selected)
			notSelected++
		}
	}
	assert.Equal(t, 3, selectedCount)
	assert.Equal(t, 3, notSelected)

	// BagIt Profiles
	selectedCount = 0
	notSelected = 0
	for _, choice := range form.Fields["BagItProfiles"].Choices {
		if util.StringListContains(testData.ProfileIds, choice.Value) {
			assert.True(t, choice.Selected)
			selectedCount++
		} else {
			assert.False(t, choice.Selected)
			notSelected++
		}
	}
	assert.Equal(t, 3, selectedCount)
	assert.Equal(t, 3, notSelected)

	// Remote Repos
	selectedCount = 0
	notSelected = 0
	for _, choice := range form.Fields["RemoteRepositories"].Choices {
		if util.StringListContains(testData.RepoIds, choice.Value) {
			assert.True(t, choice.Selected)
			selectedCount++
		} else {
			assert.False(t, choice.Selected)
			notSelected++
		}
	}
	assert.Equal(t, 3, selectedCount)
	assert.Equal(t, 3, notSelected)

	// Storage Services
	selectedCount = 0
	notSelected = 0
	for _, choice := range form.Fields["StorageServices"].Choices {
		if util.StringListContains(testData.StorageServiceIds, choice.Value) {
			assert.True(t, choice.Selected)
			selectedCount++
		} else {
			assert.False(t, choice.Selected)
			notSelected++
		}
	}
	assert.Equal(t, 3, selectedCount)
	assert.Equal(t, 3, notSelected)
}

func TestExportSettingsObjectIds(t *testing.T) {
	testData := getExportSettingsTestData()

	ids, err := testData.ExportSettings.ObjectIds(constants.TypeAppSetting)
	require.NoError(t, err)
	assert.EqualValues(t, testData.AppSettingIds, ids)

	ids, err = testData.ExportSettings.ObjectIds(constants.TypeBagItProfile)
	require.NoError(t, err)
	assert.EqualValues(t, testData.ProfileIds, ids)

	ids, err = testData.ExportSettings.ObjectIds(constants.TypeRemoteRepository)
	require.NoError(t, err)
	assert.EqualValues(t, testData.RepoIds, ids)

	ids, err = testData.ExportSettings.ObjectIds(constants.TypeStorageService)
	require.NoError(t, err)
	assert.EqualValues(t, testData.StorageServiceIds, ids)

	ids, err = testData.ExportSettings.ObjectIds("type does not exist")
	assert.Equal(t, constants.ErrUnknownType, err)
	assert.Empty(t, ids)
}

func TestContainsPlaintextPassword(t *testing.T) {
	settings := core.NewExportSettings()
	ss1 := &core.StorageService{}
	ss2 := &core.StorageService{}
	settings.StorageServices = append(settings.StorageServices, ss1, ss2)

	// False because passwords are empty
	assert.False(t, settings.ContainsPlaintextPassword())

	// False because password comes from env
	ss1.Password = "env:PASSWORD_FROM_ENV"
	assert.False(t, settings.ContainsPlaintextPassword())

	// True because we have a plaintext, non-env password
	ss2.Password = "eugenekrabs"
	assert.True(t, settings.ContainsPlaintextPassword())
}

func TestContainsPlaintextAPIToken(t *testing.T) {
	settings := core.NewExportSettings()
	repo1 := &core.RemoteRepository{}
	repo2 := &core.RemoteRepository{}
	settings.RemoteRepositories = append(settings.RemoteRepositories, repo1, repo2)

	// False because tokens are empty
	assert.False(t, settings.ContainsPlaintextAPIToken())

	// False because token comes from env
	repo1.APIToken = "env:TOKEN_FROM_ENV"
	assert.False(t, settings.ContainsPlaintextAPIToken())

	// True because we have a plaintext, non-env token
	repo2.APIToken = "patrickstar"
	assert.True(t, settings.ContainsPlaintextAPIToken())

}

func loadExportSettingsWithQuestions(t *testing.T) *core.ExportSettings {
	data, err := os.ReadFile(filepath.Join(util.PathToTestData(), "files", "export_settings_with_questions.json"))
	require.Nil(t, err)
	settings := &core.ExportSettings{}
	require.Nil(t, json.Unmarshal(data, settings))
	return settings
}

func TestExportSettingsSetValueFromResponse(t *testing.T) {
	settings := loadExportSettingsWithQuestions(t)
	require.NotEmpty(t, settings.Questions)

	responses := map[string]string{
		"Login":          "user@example.com",
		"Port":           "9899",
		"AllowsDownload": "no",
		"Value":          "dingo",
		"UserID":         "marge@example.com",
	}
	for _, q := range settings.Questions {
		response, ok := responses[q.Field]
		if !ok {
			response = "tag value " + q.Field
		}
		require.Nil(t, settings.SetValueFromResponse(q, response), q.Prompt)
	}

	ss := findStorageService(settings, "00000000-6870-48a3-9dd7-89e21410453b")
	assert.Equal(t, "user@example.com", ss.Login)
	assert.Equal(t, 9899, ss.Port)
	sftp := findStorageService(settings, "00000000-d761-4c03-ab5b-266bacc40f3f")
	assert.False(t, sftp.AllowsDownload)
	for _, repo := range settings.RemoteRepositories {
		if repo.ID == "00000000-9d6a-466e-91f0-541cf9894232" {
			assert.Equal(t, "marge@example.com", repo.UserID)
		}
	}
	for _, setting := range settings.AppSettings {
		if setting.ID == "00000000-4fd5-4543-9742-6d7b4469dde8" {
			assert.Equal(t, "dingo", setting.Value)
		}
	}
	for _, profile := range settings.BagItProfiles {
		for _, tag := range profile.Tags {
			if tag.ID == "00000000-6d9c-4825-92cf-e472fb05f3d4" {
				assert.Equal(t, "tag value "+tag.ID, tag.DefaultValue)
			}
		}
	}
}

// DART 2 exports use JSON-style field names in questions
// (login, port, allowsDownload, value, userId).
func TestExportSettingsSetValueFromResponseDart2FieldNames(t *testing.T) {
	settings := loadExportSettingsWithQuestions(t)
	for _, q := range settings.Questions {
		if q.ObjType != constants.TypeBagItProfile {
			q.Field = strings.ToLower(q.Field[:1]) + q.Field[1:]
			if q.Field == "userID" {
				q.Field = "userId"
			}
		}
	}
	for _, q := range settings.Questions {
		response := "x"
		switch q.Field {
		case "port":
			response = "22"
		case "allowsDownload":
			response = "true"
		}
		require.Nil(t, settings.SetValueFromResponse(q, response), q.Field)
	}
	ss := findStorageService(settings, "00000000-6870-48a3-9dd7-89e21410453b")
	assert.Equal(t, "x", ss.Login)
	assert.Equal(t, 22, ss.Port)
}

func TestExportSettingsSetValueFromResponseErrors(t *testing.T) {
	settings := loadExportSettingsWithQuestions(t)
	q := *settings.Questions[0] // StorageService Login question

	bad := q
	bad.Field = "NoSuchField"
	assert.ErrorIs(t, settings.SetValueFromResponse(&bad, "x"), constants.ErrNoSuchField)

	bad = q
	bad.ObjID = "11111111-1111-1111-1111-111111111111"
	assert.ErrorIs(t, settings.SetValueFromResponse(&bad, "x"), constants.ErrNoSuchField)

	bad = q
	bad.Field = "Port"
	assert.NotNil(t, settings.SetValueFromResponse(&bad, "not a number"))

	bad = q
	bad.ObjType = "Bogus"
	assert.ErrorIs(t, settings.SetValueFromResponse(&bad, "x"), constants.ErrUnknownType)

	assert.NotNil(t, settings.SetValueFromResponse(nil, "x"))
}

func findStorageService(settings *core.ExportSettings, id string) *core.StorageService {
	for _, ss := range settings.StorageServices {
		if ss.ID == id {
			return ss
		}
	}
	return nil
}
