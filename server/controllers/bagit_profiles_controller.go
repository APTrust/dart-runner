package controllers

import (
	"net/http"
	"strings"

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
	// Load the existing profile here, so we get the existing
	// tag definitions, which will not be on the submitted form.
	profile := &core.BagItProfile{}
	result := core.ObjFind(c.Param("id"))
	if result.BagItProfile() != nil {
		profile = result.BagItProfile()
	}

	err := c.Bind(profile)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	// For tag files allowed, split lines into individual file names.
	// This data comes from a textarea, with one file name per line.
	if len(profile.TagFilesAllowed) == 1 {
		rawString := strings.TrimSpace(profile.TagFilesAllowed[0])
		allowed := strings.Split(rawString, "\n")
		for i, _ := range allowed {
			allowed[i] = strings.TrimSpace(allowed[i])
		}
		profile.TagFilesAllowed = allowed
	}
	err = core.ObjSave(profile)
	if err != nil {
		objectExistsInDB, _ := core.ObjExists(profile.ID)
		data := gin.H{
			"form":             profile.ToForm(),
			"objectExistsInDB": objectExistsInDB,
			"errMsg":           "Please correct the following errors",
			"errors":           profile.Errors,
		}
		c.HTML(http.StatusBadRequest, "bagit_profile/form.html", data)
		return
	}
	c.Redirect(http.StatusFound, "/profiles")

}
