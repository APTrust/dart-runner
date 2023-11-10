package controllers_test

import (
	"encoding/json"
	"path"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/require"
)

// These come from fixtures in testdata/files
var idOfSettingsWithQuestions = "00000000-97bb-470d-9fcd-7b55c5ae04bb"
var idOfSettingsWithoutQuestions = "11111111-97bb-470d-9fcd-7b55c5ae04bb"

func loadExportSettings(t *testing.T) {
	fixtures := []string{
		"export_settings_with_questions.json",
		"export_settings_no_questions.json",
	}
	for _, fixture := range fixtures {
		file := path.Join(util.ProjectRoot(), "testdata", "files", fixture)
		data, err := util.ReadFile(file)
		require.Nil(t, err)
		settings := &core.ExportSettings{}
		err = json.Unmarshal(data, settings)
		require.Nil(t, err)
		err = core.ObjSave(settings)
		require.Nil(t, err, settings.Name)
	}
}

func TestSettingsExportIndex(t *testing.T) {
	defer core.ClearDartTable()
	loadExportSettings(t)

}

func TestSettingsExportEdit(t *testing.T) {

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
