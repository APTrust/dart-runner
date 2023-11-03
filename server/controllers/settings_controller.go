package controllers

import (
	"encoding/json"
	"fmt"
	"html/template"
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
	exportSettings, err := getExportSettings(c.Param("id"))
	if err != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, err)
		return
	}
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
	exportSettings, err := getExportSettings(c.Param("id"))
	if err != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, err)
		return
	}
	exportSettings.Name = c.PostForm("Name")

	// Include collections of settings that the user
	// specified on the HTML form.
	// Note that we're not dealing with questions here.
	// We'll deal with those in the questions endpoints.

	err = setExportSettingsCollections(c, exportSettings)
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
	exportSettings, err := getExportSettings(c.Param("id"))
	if err != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, err)
		return
	}
	err = core.ObjDelete(exportSettings)
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
	exportSettings, err := getExportSettings(c.Param("id"))
	if err != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, err)
		return
	}
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
// func SettingsExportShowQuestions(c *gin.Context) {
// 	exportSettings, err := getExportSettings(c.Param("id"))
// 	if err != nil {
// 		AbortWithErrorHTML(c, http.StatusNotFound, err)
// 		return
// 	}
// 	if len(exportSettings.Questions) == 0 {
// 		exportSettings.Questions = append(exportSettings.Questions, core.NewExportQuestion())
// 	}

// 	// We show options related to the export settings only, not all options.
// 	// Showing all confuses the user because many don't apply to the settings at hand.
// 	opts := core.NewExportOptionsFromSettings(exportSettings)
// 	optionsJson, err := json.Marshal(opts)
// 	if err != nil {
// 		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
// 		return
// 	}
// 	data := gin.H{
// 		"settings":    exportSettings,
// 		"optionsJson": template.JS(string(optionsJson)),
// 	}
// 	c.HTML(http.StatusOK, "settings/question_form.html", data)
// }

// GET /settings/export/questions/new/:id
func SettingsExportNewQuestion(c *gin.Context) {
	exportSettings, err := getExportSettings(c.Param("id"))
	if err != nil {
		AbortWithErrorJSON(c, http.StatusNotFound, err)
		return
	}
	question := core.NewExportQuestion()

	// We show options related to the export settings only, not all options.
	// Showing all confuses the user because many don't apply to the settings at hand.
	opts := core.NewExportOptionsFromSettings(exportSettings)
	optionsJson, err := json.Marshal(opts)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}

	data := gin.H{
		"settings":    exportSettings,
		"question":    question,
		"form":        question.ToForm(),
		"optionsJson": template.JS(string(optionsJson)),
	}
	c.HTML(http.StatusOK, "settings/question.html", data)
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
	exportSettings, err := getExportSettings(c.Param("settings_id"))
	if err != nil {
		AbortWithErrorJSON(c, http.StatusNotFound, err)
		return
	}
	questions := make([]*core.ExportQuestion, 0)
	for _, question := range exportSettings.Questions {
		if question.ID != c.Param("question_id") {
			questions = append(questions, question)
		}
	}
	exportSettings.Questions = questions
	err = core.ObjSave(exportSettings)
	if err != nil {
		AbortWithErrorJSON(c, http.StatusInternalServerError, err)
		return
	}
	data := map[string]string{
		"status":   "OK",
		"location": fmt.Sprintf("/settings/export/questions/%s", c.Param("settings_id")),
	}
	c.JSON(http.StatusOK, data)
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

// GET /settings/profile_tags
func SettingsProfileTagList(c *gin.Context) {
	profileID := c.Query("profileID")
	list, err := core.UserSettableTagsForProfile(profileID)
	if err != nil {
		AbortWithErrorJSON(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, list)
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

func getExportSettings(id string) (*core.ExportSettings, error) {
	result := core.ObjFind(id)
	if result.Error != nil {
		return nil, result.Error
	}
	return result.ExportSetting(), nil
}
