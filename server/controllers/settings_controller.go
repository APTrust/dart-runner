package controllers

import (
	"encoding/json"
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
	result := core.ObjList(constants.TypeExportSettings, "obj_name", 100, 0)
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
		"flash":    GetFlashCookie(c),
	}
	c.HTML(http.StatusOK, "settings/export.html", data)
}

// SettingsExportSave saves ExportSettings.
//
// POST /settings/export/save
func SettingsExportSave(c *gin.Context) {
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		if result.Error != nil {
			AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
			return
		}
	}
	exportSettings := result.ExportSetting()
	exportSettings.Name = c.PostForm("Name")

	// Include collections of settings that the user
	// specified on the HTML form.
	// Note that we're not dealing with questions here.
	// We'll deal with those in the questions endpoints.

	err := setExportSettingsCollections(c, exportSettings)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, err)
		return
	}

	err = core.ObjSave(exportSettings)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}

	SetFlashCookie(c, "Settings have been saved.")
	c.Redirect(http.StatusFound, fmt.Sprintf("/settings/export/edit/%s", exportSettings.ID))
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

// SettingsExportDelete deletes the ExportSettings record with the specified ID.
//
// POST /settings/export/delete/:id
func SettingsExportDelete(c *gin.Context) {
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	exportSettings := result.ExportSetting()
	err := core.ObjDelete(exportSettings)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusFound, "/settings/export/")
}

// SettingsExportShowJson shows the JSON representation of
// an ExportSettings object. This is the value a user will
// copy to share settings with others.
//
// GET /settings/export/show_json/:id
func SettingsExportShowJson(c *gin.Context) {
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	exportSettings := result.ExportSetting()
	jsonData, err := json.MarshalIndent(exportSettings, "", "  ")
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}
	displayPasswordWarning := exportSettings.ContainsPlaintextPassword()
	displayTokenWarning := exportSettings.ContainsPlaintextAPIToken()
	data := gin.H{
		"settings":                exportSettings,
		"json":                    string(jsonData),
		"displayPasswordWarning":  displayPasswordWarning,
		"displayTokenWarning":     displayTokenWarning,
		"displayPlaintextWarning": displayPasswordWarning || displayTokenWarning,
	}
	c.HTML(http.StatusOK, "settings/export_result.html", data)
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

// setExportSettingsCollections sets AppSettings, BagItProfiles,
// RemoteRepositories, and StorageServices on the exportSettings
// object based on values the user submitted in the HTML form.
func setExportSettingsCollections(c *gin.Context, exportSettings *core.ExportSettings) error {

	exportSettings.AppSettings = make([]*core.AppSetting, 0)
	for _, uuid := range c.PostFormArray("AppSettings") {
		result := core.ObjFind(uuid)
		if result.Error != nil {
			return result.Error
		}
		exportSettings.AppSettings = append(exportSettings.AppSettings, result.AppSetting())
	}

	exportSettings.BagItProfiles = make([]*core.BagItProfile, 0)
	for _, uuid := range c.PostFormArray("BagItProfiles") {
		result := core.ObjFind(uuid)
		if result.Error != nil {
			return result.Error
		}
		exportSettings.BagItProfiles = append(exportSettings.BagItProfiles, result.BagItProfile())
	}

	exportSettings.RemoteRepositories = make([]*core.RemoteRepository, 0)
	for _, uuid := range c.PostFormArray("RemoteRepositories") {
		result := core.ObjFind(uuid)
		if result.Error != nil {
			return result.Error
		}
		exportSettings.RemoteRepositories = append(exportSettings.RemoteRepositories, result.RemoteRepository())
	}

	exportSettings.StorageServices = make([]*core.StorageService, 0)
	for _, uuid := range c.PostFormArray("StorageServices") {
		result := core.ObjFind(uuid)
		if result.Error != nil {
			return result.Error
		}
		exportSettings.StorageServices = append(exportSettings.StorageServices, result.StorageService())
	}

	return nil
}
