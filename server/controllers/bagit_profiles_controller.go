package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BagItProfileCreate creates a new Profile.
// Handles submission of new Profile form.
// POST /profiles/new
func BagItProfileCreate(c *gin.Context) {

}

// GET /profiles/delete/:id
// POST /profiles/delete/:id
func BagItProfileDelete(c *gin.Context) {

}

// GET /profiles/edit/:id
func BagItProfileEdit(c *gin.Context) {

}

// GET /profiles
func BagItProfileIndex(c *gin.Context) {
	request := NewRequest(c)
	if request.HasErrors() {
		c.AbortWithError(http.StatusInternalServerError, request.Errors[0])
		return
	}
	request.TemplateData["items"] = request.QueryResult.BagItProfiles
	c.HTML(http.StatusOK, "bagit_profile/list.html", request.TemplateData)
}

// GET /profiles/new
func BagItProfileNew(c *gin.Context) {

}

// GET /profiles/import_start
func BagItProfileImportStart(c *gin.Context) {

}

// POST /profiles/import
func BagItProfileImport(c *gin.Context) {

}

// GET /profiles/export
func BagItProfileExport(c *gin.Context) {

}

// PUT /profiles/edit/:id
// POST /profiles/edit/:id
func BagItProfileUpdate(c *gin.Context) {

}
