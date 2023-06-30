package controllers

import (
	"net/http"

	"github.com/APTrust/dart-runner/core"
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
	request := NewRequest(c)
	if request.HasErrors() {
		c.AbortWithError(http.StatusInternalServerError, request.Errors[0])
		return
	}
	profile := request.QueryResult.BagItProfile()
	request.TemplateData["tagFileNames"] = profile.TagFileNames()
	tagMap := make(map[string][]*core.TagDefinition)
	for _, name := range profile.TagFileNames() {
		tagMap[name] = profile.TagsInFile(name)
	}
	request.TemplateData["tagsInFile"] = tagMap

	c.HTML(http.StatusOK, "bagit_profile/form.html", request.TemplateData)
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
func BagItProfileSave(c *gin.Context) {
	profile := &core.BagItProfile{}
	err := c.Bind(profile)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	err = core.ObjSave(profile)
	if err != nil {
		objectExistsInDB, _ := core.ObjExists(profile.ID)
		data := gin.H{
			"form":             profile.ToForm(),
			"objectExistsInDB": objectExistsInDB,
		}
		c.HTML(http.StatusBadRequest, "bagit_profile/form.html", data)
		return
	}
	c.Redirect(http.StatusFound, "/bagit_profiles")

}
