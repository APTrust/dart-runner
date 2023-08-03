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

// GET /jobs/delete/:id
// POST /jobs/delete/:id
func JobDelete(c *gin.Context) {
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	err := core.ObjDelete(result.Job())
	if err != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, err)
		return
	}
	c.Redirect(http.StatusFound, "/jobs")
}

// GET /jobs
func JobIndex(c *gin.Context) {
	request := NewRequest(c)
	if request.HasErrors() {
		AbortWithErrorHTML(c, http.StatusInternalServerError, request.Errors[0])
		return
	}
	request.TemplateData["items"] = request.QueryResult.Jobs
	c.HTML(http.StatusOK, "job/list.html", request.TemplateData)

}

// GET /jobs/new
func JobNew(c *gin.Context) {
	job := core.NewJob()
	err := core.ObjSaveWithoutValidation(job)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}
	err = core.ObjSaveWithoutValidation(job)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/jobs/files/%s", job.ID))
}

// GET /jobs/packaging/:id
func JobShowPackaging(c *gin.Context) {
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	job := result.Job()
	data := gin.H{
		"job":  job,
		"form": job.ToForm(),
	}
	c.HTML(http.StatusOK, "job/packaging.html", data)
}

// POST /jobs/packaging/:id
func JobSavePackaging(c *gin.Context) {
	jobId := c.Param("id")
	direction := c.PostForm("direction")
	nextPage := fmt.Sprintf("/jobs/metadata/%s", jobId)
	if direction == "back" {
		nextPage = fmt.Sprintf("/jobs/files/%s", jobId)
	}

	result := core.ObjFind(jobId)
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	job := result.Job()
	job.PackageOp.BagItSerialization = c.PostForm("Serialization")
	job.PackageOp.OutputPath = c.PostForm("OutputPath")
	job.PackageOp.PackageFormat = c.PostForm("PackageFormat")
	job.PackageOp.PackageName = c.PostForm("PackageName")

	bagItProfileID := c.PostForm("BagItProfileID")
	if job.BagItProfile == nil || job.BagItProfile.ID != bagItProfileID {
		profileResult := core.ObjFind(bagItProfileID)
		if profileResult.Error != nil {
			AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
			return
		}
		job.BagItProfile = profileResult.BagItProfile()
	}

	ok := job.PackageOp.Validate()
	if !ok {
		// Errors from sub-object have sub-object prefix for
		// display when running jobs from command line. We
		// want to strip that prefix here.
		errors := make(map[string]string)
		for key, value := range job.PackageOp.Errors {
			fieldName := strings.Replace(key, "PackageOperation.", "", 1)
			errors[fieldName] = value
		}
		form := job.ToForm()
		form.Errors = errors
		data := gin.H{
			"job":  job,
			"form": form,
		}
		c.HTML(http.StatusBadRequest, "job/packaging.html", data)
		return
	}
	err := core.ObjSaveWithoutValidation(job)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusFound, nextPage)
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
		fieldName := fmt.Sprintf("%s/%s", tagDef.TagFile, tagDef.TagName)
		tagDef.UserValue = c.PostForm(fieldName)
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

// GET /jobs/upload/:id
func JobShowUpload(c *gin.Context) {

}

// POST /jobs/upload/:id
func JobSaveUpload(c *gin.Context) {

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
			if tagDef.SystemMustSet || !util.IsEmpty(tagDef.DefaultValue) {
				formGroupClass = "form-group-hidden"
			}
			field := &core.Field{
				Attrs:          make(map[string]string),
				ID:             tagDef.ID,
				Name:           fmt.Sprintf("%s/%s", tagDef.TagFile, tagDef.TagName),
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
