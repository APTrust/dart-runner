package controllers

import (
	"fmt"
	"net/http"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/gin-gonic/gin"
)

// SettingsExportIndex displays a list of ExportSettings
// objects.
//
// GET /settings/export
func SettingsExportIndex(c *gin.Context) {
	result := core.ObjList(constants.TypeExportSettings, "obj_name", 0, 100)
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, result.Error)
		return
	}
	data := gin.H{
		"items": result.ExportSettings,
	}
	c.HTML(http.StatusOK, "settings/list.html", data)
}

// SettingsExportEdit shows a form on which user can edit
// the specified ExportSettings.
//
// GET /settings/export/edit/:id
func SettingsExportEdit(c *gin.Context) {
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	exportSettings := result.ExportSetting()
	data := gin.H{
		"settings": exportSettings,
		"form":     exportSettings.ToForm(),
	}
	c.HTML(http.StatusOK, "settings/export.html", data)
}

// SettingsExportSave saves ExportSettings.
//
// POST /settings/export/save
func SettingsExportSave(c *gin.Context) {

}

// SettingsExportNew creates a new ExportSettings object
// and then redirects to the edit form.
//
// GET /settings/export/new
func SettingsExportNew(c *gin.Context) {
	exportSettings := core.NewExportSettings()
	err := core.ObjSave(exportSettings)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/settings/export/edit/%s", exportSettings.ID))
}

// SettingsExportShowJson shows the JSON representation of
// an ExportSettings object. This is the value a user will
// copy to share settings with others.
//
// GET /settings/export/show_json/:id
func SettingsExportShowJson(c *gin.Context) {

}

// SettingsExportShowQuestions displays a page on which the
// user can create, edit and delete ExportQuestions.
//
// GET /settings/export/questions/:id
func SettingsExportShowQuestions(c *gin.Context) {

}

// SettingsExportSaveQuestions saves questions attached
// to the specified ExportSettings object.
//
// POST /settings/export/questions/:id
func SettingsExportSaveQuestions(c *gin.Context) {

}

// SettingsExportDeleteQuestion deletes a question from ExportSettings.
//
// POST /settings/export/questions/delete/:settings_id/:question_id
func SettingsExportDeleteQuestion(c *gin.Context) {

}

// SettingsImport shows a form on which user can specify a URL
// from which to import settings, or a blob of JSON to be imported
// directly.
//
// GET /settings/import
func SettingsImport(c *gin.Context) {
	c.HTML(http.StatusOK, "settings/import.html", gin.H{})
}

// SettingsImportFromUrl imports settings from a URL.
//
// POST /settings/import/url
func SettingsImportFromUrl(c *gin.Context) {

}

// SettingsImportFromJson imports JSON from a blob that the
// user pasted into a textarea on the settings/import page.
//
// POST /settings/import/json
func SettingsImportFromJson(c *gin.Context) {

}
