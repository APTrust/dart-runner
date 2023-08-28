package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WorkflowCreate creates a new Workflow.
// Handles submission of new Workflow form.
// POST /workflows/new
func WorkflowCreate(c *gin.Context) {

}

// GET /workflows/delete/:id
// POST /workflows/delete/:id
func WorkflowDelete(c *gin.Context) {

}

// GET /workflows/edit/:id
func WorkflowEdit(c *gin.Context) {

}

// GET /workflows
func WorkflowIndex(c *gin.Context) {
	request := NewRequest(c)
	if request.HasErrors() {
		AbortWithErrorHTML(c, http.StatusInternalServerError, request.Errors[0])
		return
	}
	request.TemplateData["items"] = request.QueryResult.Workflow
	c.HTML(http.StatusOK, "workflow/list.html", request.TemplateData)
}

// GET /workflows/new
func WorkflowNew(c *gin.Context) {

}

// GET /workflows/show/:id
func WorkflowShow(c *gin.Context) {

}

// PUT /workflows/edit/:id
// POST /workflows/edit/:id
func WorkflowUpdate(c *gin.Context) {

}

// POST /workflows/run/:id
func WorkflowRun(c *gin.Context) {

}

// POST /workflows/runbatch/:id
func WorkflowRunBatch(c *gin.Context) {

}
