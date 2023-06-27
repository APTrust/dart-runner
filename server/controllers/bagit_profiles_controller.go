package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProfileCreate creates a new Profile.
// Handles submission of new Profile form.
// POST /profiles/new
func ProfileCreate(c *gin.Context) {

}

// GET /profiles/delete/:id
// POST /profiles/delete/:id
func ProfileDelete(c *gin.Context) {

}

// GET /profiles/edit/:id
func ProfileEdit(c *gin.Context) {

}

// GET /profiles
func ProfileIndex(c *gin.Context) {
	request := NewRequest(c)
	if request.HasErrors() {
		c.AbortWithError(http.StatusInternalServerError, request.Errors[0])
		return
	}
	request.TemplateData["items"] = request.QueryResult.BagItProfiles
	c.HTML(http.StatusOK, "bagit_profile/list.html", request.TemplateData)
}

// GET /profiles/new
func ProfileNew(c *gin.Context) {

}

// GET /profiles/import_start
func ProfileImportStart(c *gin.Context) {

}

// POST /profiles/import
func ProfileImport(c *gin.Context) {

}

// GET /profiles/export
func ProfileExport(c *gin.Context) {

}

// PUT /profiles/edit/:id
// POST /profiles/edit/:id
func ProfileUpdate(c *gin.Context) {

}
