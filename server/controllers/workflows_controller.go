package controllers

import (
	"fmt"
	"net/http"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/gin-gonic/gin"
)

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

// PUT /workflows/delete/:id
// POST /workflows/delete/:id
func WorkflowDelete(c *gin.Context) {
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	err := core.ObjDelete(result.Workflow())
	if err != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, err)
		return
	}
	c.Redirect(http.StatusFound, "/workflows")
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

// PUT /workflows/edit/:id
// POST /workflows/edit/:id
func WorkflowSave(c *gin.Context) {
	workflow := &core.Workflow{}
	err := c.Bind(workflow)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusBadRequest, err)
		return
	}
	profileID := c.PostForm("BagItProfileID")
	if util.LooksLikeUUID(profileID) {
		result := core.ObjFind(profileID)
		if result.Error == nil && result.BagItProfile() != nil {
			workflow.BagItProfile = result.BagItProfile()
		}
	}
	err = core.ObjSave(workflow)
	if err != nil {
		objectExistsInDB, _ := core.ObjExists(workflow.ID)
		data := gin.H{
			"form":             workflow.ToForm(),
			"objectExistsInDB": objectExistsInDB,
		}
		c.HTML(http.StatusBadRequest, "workflow/form.html", data)
		return
	}
	c.Redirect(http.StatusFound, "/workflows")

}

// GET /workflows/export/:id
func WorkflowExport(c *gin.Context) {
	// Export for runner. Use full storage service def & bagit profile.
	// This will display in modal with a copy button.
}

// POST /workflows/run/:id
func WorkflowRun(c *gin.Context) {

}

// POST /workflows/runbatch/:id
func WorkflowRunBatch(c *gin.Context) {

}
