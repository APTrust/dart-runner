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

// GET /profiles/export/:id
func BagItProfileExport(c *gin.Context) {
	data := map[string]string{
		"modalTitle":   "This here is the title",
		"modalContent": "<p>And this is <b>modal content</b> with some formatting.</p>",
	}
	c.JSON(http.StatusOK, data)
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

// GET /profiles/new_tag/:profile_id
func BagItProfileNewTag(c *gin.Context) {

}

// POST /profiles/new_tag/:profile_id
func BagItProfileCreateTag(c *gin.Context) {

}

// GET /profiles/edit_tag/:profile_id/:tag_id
func BagItProfileEditTag(c *gin.Context) {
	// TODO: Finish TagDefinition.ToForm() & use "tag_definition/form.html"
	// This one displays in a modal.
	result := core.ObjFind(c.Param("profile_id"))
	if result.Error != nil {
		c.AbortWithError(http.StatusNotFound, result.Error)
		return
	}
	profile := result.BagItProfile()
	tag, err := profile.FirstMatchingTag("ID", c.Param("tag_id"))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// sb := &strings.Builder{}
	// templateData := gin.H{
	// 	"bagItProfileID": profile.ID,
	// 	"tag":            tag,
	// 	"form":           tag.ToForm(),
	// }
	// err = util.Template.Lookup("tag_definition/form.html").Execute(sb, templateData)
	// if err != nil {
	// 	c.AbortWithError(http.StatusInternalServerError, err)
	// 	return
	// }

	data := map[string]string{
		"modalTitle":   "Edit Tag Definition",
		"modalContent": tag.TagName,
	}

	c.JSON(http.StatusOK, data)
}

// POST /profiles/edit_tag/:profile_id/:tag_id
// PUT  /profiles/edit_tag/:profile_id/:tag_id
func BagItProfileSaveTag(c *gin.Context) {

}

// POST /profiles/delete_tag/:profile_id/:tag_id
// PUT  /profiles/delete_tag/:profile_id/:tag_id
func BagItProfileDeleteTag(c *gin.Context) {

}

// POST /profiles/new_tag_file/:profile_id
func BagItProfileCreateTagFile(c *gin.Context) {

}

// POST /profiles/delete_tag_file/:profile_id
// PUT  /profiles/delete_tag_file/:profile_id
func BagItProfileDeleteTagFile(c *gin.Context) {

}
