package controllers_test

import (
	"encoding/json"
	"fmt"
	"path"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/require"
)

const (
	NO_QUESTIONS   = 0
	WITH_QUESTIONS = 1
)

// This loads settings fixtures into the database and returns
// a slice of the two settings objects. Note that settings without
// questions is the first item in the list, and settings with
// questions is the second item, as in the consts above.
func loadExportSettings(t *testing.T) []*core.ExportSettings {
	settingsList := make([]*core.ExportSettings, 2)
	fixtures := []string{
		"export_settings_no_questions.json",
		"export_settings_with_questions.json",
	}
	for i, fixture := range fixtures {
		file := path.Join(util.ProjectRoot(), "testdata", "files", fixture)
		data, err := util.ReadFile(file)
		require.Nil(t, err)
		settings := &core.ExportSettings{}
		err = json.Unmarshal(data, settings)
		require.Nil(t, err)
		err = core.ObjSave(settings)
		require.Nil(t, err, settings.Name)
		settingsList[i] = settings

		// We need to save the object attached to these
		// settings so they'll appear as options on the
		// settings export page.
		for _, appSetting := range settings.AppSettings {
			require.NoError(t, core.ObjSave(appSetting))
		}
		for _, profile := range settings.BagItProfiles {
			require.NoError(t, core.ObjSave(profile))
		}
		for _, repo := range settings.RemoteRepositories {
			require.NoError(t, core.ObjSave(repo))
		}
		for _, ss := range settings.StorageServices {
			require.NoError(t, core.ObjSave(ss))
		}
	}
	return settingsList
}

func loadObjectsForExportTests(t *testing.T) {
	_, err := core.CreateAppSettings(2)
	require.NoError(t, err)

	_, err = core.CreateBagItProfiles(2)
	require.NoError(t, err)

	_, err = core.CreateRemoteRepos(2)
	require.NoError(t, err)

	_, err = core.CreateStorageServices(2)
	require.NoError(t, err)
}

func setUpExportTest(t *testing.T) []*core.ExportSettings {
	loadObjectsForExportTests(t)
	return loadExportSettings(t)
}

func TestSettingsExportIndex(t *testing.T) {
	defer core.ClearDartTable()
	settings := setUpExportTest(t)
	expected := []string{
		settings[0].ID,
		settings[0].Name,
		settings[1].ID,
		settings[1].Name,
	}
	DoSimpleGetTest(t, "/settings/export", expected)
}

func TestSettingsExportEdit(t *testing.T) {
	defer core.ClearDartTable()
	settings := loadExportSettings(t)

	expected := []string{
		settings[0].ID,
		settings[0].Name,
	}

	// This page should display all available objects of the following
	// types as checkboxes.
	types := []string{
		constants.TypeAppSetting,
		constants.TypeBagItProfile,
		constants.TypeRemoteRepository,
		constants.TypeStorageService,
	}
	for _, objType := range types {
		checkboxName := fmt.Sprintf("%ss", objType)
		if objType == constants.TypeRemoteRepository {
			checkboxName = "RemoteRepositories"
		}
		for _, item := range core.ObjNameIdList(objType) {
			html := fmt.Sprintf(`type="checkbox" name="%s" value="%s"`, checkboxName, item.ID)
			expected = append(expected, html)
			expected = append(expected, item.Name)
		}
	}
	pageUrl := fmt.Sprintf("/settings/export/edit/%s", settings[0].ID)
	DoSimpleGetTest(t, pageUrl, expected)
}

func TestSettingsExportSave(t *testing.T) {

}

func TestSettingsExportNew(t *testing.T) {

}

func TestSettingsExportDelete(t *testing.T) {

}

func TestSettingsExportShowJson(t *testing.T) {

}

func TestSettingsExportNewQuestion(t *testing.T) {

}

func TestSettingsExportSaveQuestion(t *testing.T) {

}

func TestSettingsExportEditQuestion(t *testing.T) {

}

func TestSettingsExportDeleteQuestion(t *testing.T) {

}

func TestSettingsImportShow(t *testing.T) {

}

func TestSettingsImportRun(t *testing.T) {

}

func TestSettingsImportAnswers(t *testing.T) {

}
