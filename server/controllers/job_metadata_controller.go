package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/gin-gonic/gin"
)

// This struct holds all the tag form inputs for a tag file.
// This is used on the metadata page.
type TagFileForms struct {
	Name   string
	Fields []*core.Field
}

// GET /jobs/metadata/:id
func JobShowMetadata(c *gin.Context) {
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	job := result.Job()
	tagFiles := GetTagFileForms(job, false)
	data := gin.H{
		"job":      job,
		"form":     job.ToForm(),
		"tagFiles": tagFiles,
	}
	c.HTML(http.StatusOK, "job/metadata.html", data)
}

// POST /jobs/metadata/:id
func JobSaveMetadata(c *gin.Context) {
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	job := result.Job()
	for _, tagDef := range job.BagItProfile.Tags {
		tagDef.UserValue = c.PostForm(tagDef.FullyQualifiedName())
	}
	err := core.ObjSaveWithoutValidation(job)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}
	// Go to next or previous page, as specified by user
	direction := c.PostForm("direction")
	nextPage := fmt.Sprintf("/jobs/upload/%s", job.ID)

	// If user wants to go back to the packaging page,
	// let them go. We don't need to display the errors
	// because they'll come back through this page again.
	if direction == "previous" {
		nextPage = fmt.Sprintf("/jobs/packaging/%s", job.ID)
		c.Redirect(http.StatusFound, nextPage)
	}
	if TagErrorsExist(job.BagItProfile.Tags) && direction == "next" {
		tagFiles := GetTagFileForms(job, true)
		data := gin.H{
			"job":      job,
			"form":     job.ToForm(),
			"tagFiles": tagFiles,
		}
		c.HTML(http.StatusOK, "job/metadata.html", data)
	}
	c.Redirect(http.StatusFound, nextPage)
}

func GetTagFileForms(job *core.Job, withErrors bool) []TagFileForms {
	// Get the list of tag files, in alpha order.
	tagFileNames := job.BagItProfile.TagFileNames()
	tagFiles := make([]TagFileForms, len(tagFileNames))
	for i, tagFileName := range tagFileNames {
		// Get list of tags in this file, in alpha order
		tagDefs := job.BagItProfile.TagsInFile(tagFileName)
		metadataTagFile := TagFileForms{
			Name:   tagFileName,
			Fields: make([]*core.Field, len(tagDefs)),
		}
		for j, tagDef := range tagDefs {
			formGroupClass := ""
			if tagDef.SystemMustSet() || !util.IsEmpty(tagDef.DefaultValue) {
				formGroupClass = "form-group-hidden"
			}
			field := &core.Field{
				Attrs:          make(map[string]string),
				ID:             tagDef.ID,
				Name:           tagDef.FullyQualifiedName(),
				Label:          tagDef.TagName,
				Value:          tagDef.GetValue(),
				Choices:        core.MakeChoiceList(tagDef.Values, tagDef.GetValue()),
				Required:       tagDef.Required,
				Help:           tagDef.Help,
				FormGroupClass: formGroupClass,
			}
			if withErrors {
				field.Error = ValidateTagValue(tagDef)
			}
			if strings.Contains(strings.ToLower(tagDef.TagName), "description") {
				field.Attrs["ControlType"] = "textarea"
			}
			if tagDef.WasAddedForJob {
				field.Attrs["WasAddedForJob"] = "true"
			}
			if tagDef.SystemMustSet() {
				field.Attrs["readonly"] = "readonly"
			}
			metadataTagFile.Fields[j] = field
		}
		tagFiles[i] = metadataTagFile
	}
	return tagFiles
}

func ValidateTagValue(tagDef *core.TagDefinition) string {
	tagValue := tagDef.GetValue()
	if !tagDef.IsLegalValue(tagValue) {
		return fmt.Sprintf("Tag has illegal value '%s'. Allowed values are: %s", tagValue, strings.Join(tagDef.Values, ","))
	}
	if tagDef.Required && !tagDef.EmptyOK && util.IsEmpty(tagValue) {
		return "This tag requires a value."
	}
	return ""
}

func TagErrorsExist(tags []*core.TagDefinition) bool {
	for _, tagDef := range tags {
		if ValidateTagValue(tagDef) != "" {
			return true
		}
	}
	return false
}
