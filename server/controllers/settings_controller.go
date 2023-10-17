package controllers

import "github.com/gin-gonic/gin"

// GET /settings/export
func SettingsExport(c *gin.Context) {

}

// GET /settings/export/result
func SettingsExportResult(c *gin.Context) {

}

// GET /settings/export/questions
func SettingsExportShowQuestions(c *gin.Context) {

}

// POST /settings/export/questions/:id
func SettingsExportSaveQuestion(c *gin.Context) {

}

// POST /settings/export/questions/delete/:id
func SettingsExportDeleteQuestion(c *gin.Context) {

}

// GET /settings/import
func SettingsImport(c *gin.Context) {
	// Show form to import settings from URL or JSON
}

// POST /settings/import/url
func SettingsImportFromUrl(c *gin.Context) {

}

// POST /settings/import/json
func SettingsImportFromJson(c *gin.Context) {

}
