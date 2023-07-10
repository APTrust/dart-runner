package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/APTrust/dart-runner/core"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	request.TemplateData["activeTab"] = c.DefaultQuery("tab", "navAboutTab")
	request.TemplateData["activeTagFile"] = c.Query("tagFile")
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
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		c.AbortWithError(http.StatusNotFound, result.Error)
		return
	}
	profile := result.BagItProfile()
	profileJson, err := profile.ToStandardFormat().ToJSON()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, result.Error)
		return
	}
	templateData := gin.H{
		"json": profileJson,
	}
	c.HTML(http.StatusOK, "bagit_profile/export.html", templateData)
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

// GET /profiles/new_tag/:profile_id/:tag_file
func BagItProfileNewTag(c *gin.Context) {
	tag := &core.TagDefinition{
		ID:        uuid.NewString(),
		IsBuiltIn: false,
		TagFile:   c.Param("tag_file"),
		TagName:   "New-Tag",
	}
	templateData := gin.H{
		"bagItProfileID": c.Param("profile_id"),
		"tag":            tag,
		"form":           tag.ToForm(),
	}
	c.HTML(http.StatusOK, "tag_definition/form.html", templateData)
}

// POST /profiles/new_tag/:profile_id
func BagItProfileCreateTag(c *gin.Context) {

}

// GET /profiles/edit_tag/:profile_id/:tag_id
func BagItProfileEditTag(c *gin.Context) {
	// This displays in a modal.
	profile, tag, err := loadProfileAndTag(c)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	templateData := gin.H{
		"bagItProfileID": profile.ID,
		"tag":            tag,
		"form":           tag.ToForm(),
	}

	c.HTML(http.StatusOK, "tag_definition/form.html", templateData)
}

// POST /profiles/edit_tag/:profile_id/:tag_id
// PUT  /profiles/edit_tag/:profile_id/:tag_id
func BagItProfileSaveTag(c *gin.Context) {

	// TODO: Respond with JSON? Or HTML?

	profile, tag, err := loadProfileAndTag(c)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if tag == nil {
		tag = &core.TagDefinition{}
	}
	c.Bind(tag)
	if !tag.Validate() {
		templateData := gin.H{
			"bagItProfileID": profile.ID,
			"tag":            tag,
			"form":           tag.ToForm(),
		}
		c.HTML(http.StatusBadRequest, "tag_definition/form.html", templateData)
		return
	}

	tagExists := false
	for i, existingTag := range profile.Tags {
		if existingTag.ID == tag.ID {
			profile.Tags[i] = tag
			tagExists = true
			break
		}
	}
	if !tagExists {
		profile.Tags = append(profile.Tags, tag)
	}

	// Validation error here should not be possible, since we just
	// pulled a valid profile from the DB and only altered
	// or added a single tag, which we know by now is valid.
	err = core.ObjSave(profile)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// tagMap := make(map[string][]*core.TagDefinition)
	// for _, name := range profile.TagFileNames() {
	// 	tagMap[name] = profile.TagsInFile(name)
	// }
	// templateData := gin.H{
	// 	"currentUrl":   c.Request.URL.Path,
	// 	"showAsModal":  c.Query("modal") == "true",
	// 	"tagFileNames": profile.TagFileNames(),
	// 	"tagsInFile":   tagMap,
	// }

	// Encode tab and name of tag file in query string
	// so we display the correct view to the user.
	query := url.Values{}
	query.Set("tab", "navTagFilesTab")
	query.Set("tagFile", tag.TagFile)
	c.Redirect(http.StatusFound, fmt.Sprintf("/profiles/edit/%s?%s", profile.ID, query.Encode()))

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

func loadProfileAndTag(c *gin.Context) (*core.BagItProfile, *core.TagDefinition, error) {
	result := core.ObjFind(c.Param("profile_id"))
	if result.Error != nil {
		return nil, nil, result.Error
	}
	profile := result.BagItProfile()
	tag, err := profile.FirstMatchingTag("ID", c.Param("tag_id"))
	return profile, tag, err
}
