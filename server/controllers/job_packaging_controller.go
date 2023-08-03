package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/APTrust/dart-runner/core"
	"github.com/gin-gonic/gin"
)

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
