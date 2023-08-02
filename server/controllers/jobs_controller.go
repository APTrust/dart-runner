package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/gin-gonic/gin"
)

// JobCreate creates a new Job.
// Handles submission of new Job form.
// POST /jobs/new
func JobCreate(c *gin.Context) {

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

// GET /jobs/edit/:id
func JobEdit(c *gin.Context) {

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

	// This struct holds all the tag form inputs for a tag file.
	type TagFileForms struct {
		Name   string
		Fields []*core.Field
	}

	// TODO: Break this out and attach error messages
	// where necessary.
	//
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
				ID:             tagDef.ID,
				Name:           tagDef.TagName,
				Label:          tagDef.TagName,
				Value:          tagDef.GetValue(),
				Choices:        core.MakeChoiceList(tagDef.Values, tagDef.GetValue()),
				Required:       tagDef.Required,
				Help:           tagDef.Help,
				FormGroupClass: formGroupClass,
			}
			metadataTagFile.Fields[j] = field
		}
		tagFiles[i] = metadataTagFile
	}

	data := gin.H{
		"job":      job,
		"form":     job.ToForm(),
		"tagFiles": tagFiles,
	}
	c.HTML(http.StatusOK, "job/metadata.html", data)

}

// POST /jobs/metadata/:id
func JobSaveMetadata(c *gin.Context) {

}

// GET /jobs/show/:id
func JobShow(c *gin.Context) {

}

// PUT /jobs/edit/:id
// POST /jobs/edit/:id
func JobUpdate(c *gin.Context) {

}
