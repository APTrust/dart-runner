package controllers

import (
	"fmt"
	"net/http"

	"github.com/APTrust/dart-runner/core"
	"github.com/gin-gonic/gin"
)

// WorkflowCreate creates a new Workflow.
// Handles submission of new Workflow form.
// POST /workflows/new
func WorkflowCreate(c *gin.Context) {

}

// WorkflowCreateFromJob creates a new Workflow.
// Handles submission of new Workflow form.
// POST /workflows/from_job/:jobId
func WorkflowCreateFromJob(c *gin.Context) {
	jobId := c.Param("jobId")
	result := core.ObjFind(jobId)
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	workflow, err := core.WorkFlowFromJob(result.Job())
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}
	err = core.ObjSave(workflow)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}
	data := map[string]string{
		"status":   "OK",
		"location": fmt.Sprintf("/workflows/edit/%s", workflow.ID),
	}
	c.JSON(http.StatusCreated, data)
}

// GET /workflows/delete/:id
// POST /workflows/delete/:id
func WorkflowDelete(c *gin.Context) {

}

// GET /workflows/edit/:id
func WorkflowEdit(c *gin.Context) {
	request := NewRequest(c)
	if request.HasErrors() {
		AbortWithErrorHTML(c, http.StatusInternalServerError, request.Errors[0])
		return
	}
	c.HTML(http.StatusOK, "workflow/form.html", request.TemplateData)
}

// GET /workflows
func WorkflowIndex(c *gin.Context) {
	request := NewRequest(c)
	if request.HasErrors() {
		AbortWithErrorHTML(c, http.StatusInternalServerError, request.Errors[0])
		return
	}
	request.TemplateData["items"] = request.QueryResult.Workflows
	c.HTML(http.StatusOK, "workflow/list.html", request.TemplateData)
}

// GET /workflows/new
func WorkflowNew(c *gin.Context) {

}

// PUT /workflows/edit/:id
// POST /workflows/edit/:id
func WorkflowSave(c *gin.Context) {

}

// POST /workflows/run/:id
func WorkflowRun(c *gin.Context) {

}

// POST /workflows/runbatch/:id
func WorkflowRunBatch(c *gin.Context) {

}
