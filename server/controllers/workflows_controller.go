package controllers

import (
	"fmt"
	"net/http"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /workflow/new
func WorkflowNew(c *gin.Context) {
	workflow := &core.Workflow{
		ID:            uuid.NewString(),
		Name:          "New Workflow",
		PackageFormat: constants.PackageFormatBagIt,
	}
	err := core.ObjSaveWithoutValidation(workflow)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}
	data := gin.H{
		"form":                 workflow.ToForm(),
		"suppressDeleteButton": false,
	}
	c.HTML(http.StatusOK, "workflow/form.html", data)
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
	workflow.ID = c.Param("id")
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
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	workflow := result.Workflow()
	workflowJson, err := workflow.ExportJson()
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}
	passwordWarningDisplay := "none"
	if workflow.HasPlaintextPasswords() {
		passwordWarningDisplay = "block"
	}
	data := gin.H{
		"json":                   string(workflowJson),
		"passwordWarningDisplay": passwordWarningDisplay,
	}
	c.HTML(http.StatusOK, "settings/export_result.html", data)
}

// POST /workflows/run/:id
func WorkflowRun(c *gin.Context) {
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	job := core.JobFromWorkflow(result.Workflow())
	err := core.ObjSaveWithoutValidation(job)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}
	data := map[string]string{
		"status":   "OK",
		"location": fmt.Sprintf("/jobs/files/%s", job.ID),
	}
	c.JSON(http.StatusOK, data)
}

// GET /workflows/batch/choose
func WorkflowShowBatchForm(c *gin.Context) {
	wb := &core.WorkflowBatch{}
	form := wb.ToForm()
	data := gin.H{
		"form": form,
	}
	c.HTML(http.StatusOK, "workflow/batch.html", data)
}

// POST /workflows/batch/choose
func WorkflowInitBatch(c *gin.Context) {
	workflowID := c.PostForm("WorkflowID")
	pathToCSVFile := c.PostForm("PathToCSVFile")
	workflow := core.ObjFind(workflowID).Workflow() // may be nil if workflowID is empty

	// User may have attempted to run an earlier version of this same
	// workflow, in which case it will be saved. Update the saved
	// version instead of creating a new object.
	wbName := fmt.Sprintf("%s => %s", workflow.Name, pathToCSVFile)
	result := core.ObjByNameAndType(wbName, constants.TypeWorkflowBatch)
	wb := result.WorkflowBatch()
	if result.Error != nil {
		// Not found. Create a new one.
		wb = core.NewWorkflowBatch(workflow, pathToCSVFile)
	}
	if !wb.Validate() {
		form := wb.ToForm()
		data := gin.H{
			"form":         form,
			"batchIsValid": false,
			"batchErrors":  wb.Errors,
		}
		c.HTML(http.StatusBadRequest, "workflow/batch.html", data)
		return
	}

	err := core.ObjSave(wb)
	if err != nil && err != constants.ErrUniqueConstraint {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}

	// Create a dummy dummyJob here, so the divs display on the
	// front end. If the worklflow has a packaging step, dummy dummyJob
	// should have a PacakageOp. Ditto for the workflow's upload ops.
	dummyJob := core.NewJob()
	if wb.Workflow.PackageFormat == "" {
		dummyJob.PackageOp = nil
	} else {
		dummyJob.PackageOp.PackageFormat = wb.Workflow.PackageFormat
	}
	if len(wb.Workflow.StorageServiceIDs) > 0 {
		dummyUploadOp := &core.UploadOperation{
			StorageService: core.NewStorageService(),
		}
		dummyJob.UploadOps = []*core.UploadOperation{dummyUploadOp}
	}

	form := wb.ToForm()
	data := gin.H{
		"form":            form,
		"batchIsValid":    true,
		"workflowBatchId": wb.ID,
		"job":             dummyJob,
	}
	c.HTML(http.StatusOK, "workflow/batch.html", data)
}

// GET /workflows/batch/run/:id
func WorkflowRunBatch(c *gin.Context) {
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	// wb := result.WorkflowBatch()
	// messageChannel := make(chan *core.EventMessage)

	// TODO:
	//
	// Parse CSV file
	// Convert each line to jobParams
	// Convert each jobParams to job
	// In event emitter go routine, execute each job
	// Save job outcome?
	// Save job artifacts?
	// Be sure the disconnect event is not emitted until all jobs are complete
	//
	// See job_run_controller.go for details on the emitter.
}
