package controllers

import (
	"fmt"
	"net/http"

	"github.com/APTrust/dart-runner/core"
	"github.com/gin-gonic/gin"
)

type JobListItem struct {
	Job       *core.Job
	Artifacts []core.NameIDPair
}

// PUT /jobs/delete/:id
// POST /jobs/delete/:id
func JobDelete(c *gin.Context) {
	jobID := c.Param("id")
	result := core.ObjFind(jobID)
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	err := core.ObjDelete(result.Job())
	if err != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, err)
		return
	}
	err = core.ArtifactsDeleteByJobID(jobID)
	if err != nil {
		deletionErr := fmt.Errorf("job was deleted but artifacts were not: %v", err)
		AbortWithErrorHTML(c, http.StatusInternalServerError, deletionErr)
		return
	}
	SetFlashCookie(c, fmt.Sprintf("Deleted job %s.", result.Job().Name()))
	c.Redirect(http.StatusFound, "/jobs")
}

// GET /jobs
func JobIndex(c *gin.Context) {
	request := NewRequest(c)
	if request.HasErrors() {
		AbortWithErrorHTML(c, http.StatusInternalServerError, request.Errors[0])
		return
	}
	request.TemplateData["jobs"] = request.QueryResult.Jobs
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
	c.Redirect(http.StatusFound, fmt.Sprintf("/jobs/files/%s", job.ID))
}

// GET /jobs/show_json/:id
func JobShowJson(c *gin.Context) {
	jobID := c.Param("id")
	result := core.ObjFind(jobID)
	if result.Error != nil {
		AbortWithErrorJSON(c, http.StatusInternalServerError, result.Error)
		return
	}
	c.JSON(http.StatusOK, result.Job())
}
