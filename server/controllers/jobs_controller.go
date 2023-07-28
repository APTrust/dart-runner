package controllers

import (
	"net/http"

	"github.com/APTrust/dart-runner/core"
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
	data := gin.H{
		"job": job,
	}
	c.HTML(http.StatusOK, "job/files.html", data)
}

// GET /jobs/packaging/:id
func JobShowPackaging(c *gin.Context) {
	job := core.NewJob()
	data := gin.H{
		"job":  job,
		"form": job.ToForm(),
	}
	c.HTML(http.StatusOK, "job/packaging.html", data)
}

// POST /jobs/packaging/:id
func JobSavePackaging(c *gin.Context) {

}

// GET /jobs/show/:id
func JobShow(c *gin.Context) {

}

// PUT /jobs/edit/:id
// POST /jobs/edit/:id
func JobUpdate(c *gin.Context) {

}
