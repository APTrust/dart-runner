package controllers

import (
	"net/http"

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

}

// GET /jobs/show/:id
func JobShow(c *gin.Context) {

}

// PUT /jobs/edit/:id
// POST /jobs/edit/:id
func JobUpdate(c *gin.Context) {

}
