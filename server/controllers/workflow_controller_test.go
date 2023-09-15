package controllers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowNew(t *testing.T) {
	defer core.ClearDartTable()

	// Create some storage services to appear on the workflow form.
	services := CreateStorageServices(t, 3)

	// Confirm that there are currently zero workflows in the DB.
	result := core.ObjList(constants.TypeWorkflow, "obj_name", 10, 0)
	require.Nil(t, result.Error)
	assert.Empty(t, result.Workflows)

	// Get the new workflow page and make sure it includes expected items.
	expected := []string{
		"New Workflow",
		"PackageFormat",
		"BagIt",
		services[0].ID,
		services[0].Name,
		services[1].ID,
		services[1].Name,
		services[2].ID,
		services[2].Name,
	}
	DoSimpleGetTest(t, "/workflows/new", expected)

	// Make sure the new workflow exists in the DB.
	// The WorkflowNew endpoint should create and
	// save this before showing the form.
	result = core.ObjList(constants.TypeWorkflow, "obj_name", 10, 0)
	require.Nil(t, result.Error)
	assert.Equal(t, 1, len(result.Workflows))
	assert.Equal(t, "New Workflow", result.Workflow().Name)
}

func TestWorkflowCreateEditDelete(t *testing.T) {
	defer core.ClearDartTable()
	workflow := testWorkflowCreateFromJob(t)
	testWorkflowEdit(t, workflow)
	testWorkflowDelete(t, workflow)
}

func testWorkflowCreateFromJob(t *testing.T) *core.Workflow {
	// Save a job and its associated records, so we can
	// create a workflow from it.
	job := loadTestJob(t)
	assert.NoError(t, core.ObjSave(job.BagItProfile))
	for _, op := range job.UploadOps {
		assert.NoError(t, core.ObjSave(op.StorageService))
	}
	require.NoError(t, core.ObjSave(job))

	// Post to the endpoint and make sure we get the
	// expected redirect. Note that this endpoint is called
	// via AJAX from the front-end, so it returns JSON data
	// if it succeeds, and the front-end JS follows the
	// location URL in the JSON.
	endpointUrl := fmt.Sprintf("/workflows/from_job/%s", job.ID)
	params := url.Values{}
	w := httptest.NewRecorder()
	req, err := NewPostRequest(endpointUrl, params)
	require.Nil(t, err)
	dartServer.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	responseData := make(map[string]string)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &responseData))
	assert.True(t, strings.HasPrefix(responseData["location"], "/workflows/edit/"))

	// Now let's get the redirect URL and make sure it has
	// the form to edit this newly created workflow.
	expected := []string{
		"New Workflow",
		"PackageFormat",
		"BagIt",
		job.UploadOps[0].StorageService.ID,
		job.UploadOps[0].StorageService.Name,
	}
	DoSimpleGetTest(t, responseData["location"], expected)

	// Finally, make sure the workflow was saved to the DB.
	parts := strings.Split(responseData["location"], "/")
	uuid := parts[len(parts)-1]
	workflow := core.ObjFind(uuid).Workflow()
	require.NotNil(t, workflow)

	return workflow
}

func testWorkflowEdit(t *testing.T, workflow *core.Workflow) {
	// Create some storage services to appear on the workflow form.
	services := CreateStorageServices(t, 3)

	// First test that the workflow edit form displays correctly.
	expected := []string{
		"Name",
		workflow.Name,
		"Description",
		workflow.Description,
		"PackageFormat",
		workflow.PackageFormat,
		"BagItProfileID",
		workflow.BagItProfile.ID,
		workflow.StorageServiceIDs[0],
		services[0].ID,
		services[1].ID,
		services[2].ID,
	}
	DoSimpleGetTest(t, fmt.Sprintf("/workflows/edit/%s", workflow.ID), expected)

	// Now let's submit the form and make sure changes are saved.
	originalSSID := workflow.StorageServiceIDs[0]

	alternateProfile := core.BagItProfileClone(workflow.BagItProfile)
	alternateProfile.Name = "Alternate Profile"
	alternateProfile.Description = "Description of alternate profile"
	require.NoError(t, core.ObjSave(alternateProfile))

	params := url.Values{}
	params.Set("Name", "Marge Simpson")
	params.Set("PackageFormat", constants.PackageFormatBagIt)
	params.Set("Description", "Tall, blue hair, scratchy voice.")
	params.Add("StorageServiceIDs", services[1].ID)
	params.Add("StorageServiceIDs", services[2].ID)
	params.Set("BagItProfileID", alternateProfile.ID)
	settings := PostTestSettings{
		EndpointUrl:              fmt.Sprintf("/workflows/edit/%s", workflow.ID),
		Params:                   params,
		ExpectedResponseCode:     http.StatusFound,
		ExpectedRedirectLocation: "/workflows",
	}
	DoSimplePostTest(t, settings)

	// Make sure the edits were saved
	result := core.ObjFind(workflow.ID)
	workflow = result.Workflow()
	require.Nil(t, result.Error)
	require.NotNil(t, workflow)
	assert.Equal(t, "Marge Simpson", workflow.Name)
	assert.Equal(t, constants.PackageFormatBagIt, workflow.PackageFormat)
	assert.Equal(t, "Tall, blue hair, scratchy voice.", workflow.Description)
	assert.Contains(t, workflow.StorageServiceIDs, services[1].ID)
	assert.Contains(t, workflow.StorageServiceIDs, services[2].ID)
	assert.NotContains(t, workflow.StorageServiceIDs, originalSSID)
	assert.Equal(t, alternateProfile.ID, workflow.BagItProfile.ID)
}

func testWorkflowDelete(t *testing.T, workflow *core.Workflow) {

}

func TestWorkflowIndex(t *testing.T) {
	defer core.ClearDartTable()
}

func TestWorkflowSave(t *testing.T) {
	defer core.ClearDartTable()
}

func TestWorkflowExport(t *testing.T) {
	defer core.ClearDartTable()
}

func TestWorkflowRun(t *testing.T) {
	defer core.ClearDartTable()
}

func TestWorkflowRunBatch(t *testing.T) {
	// TODO: Implement "run batch" first
}
