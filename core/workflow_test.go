package core_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadJsonWorkflow(t *testing.T) *core.Workflow {
	pathToFile := filepath.Join(util.PathToTestData(), "files", "runner_test_workflow.json")
	workflow, err := core.WorkflowFromJson(pathToFile)
	require.Nil(t, err)
	require.NotNil(t, workflow)
	require.True(t, workflow.Validate(), workflow.Errors)
	require.Empty(t, workflow.Errors)
	require.NotNil(t, workflow.BagItProfile)
	return workflow
}

func TestWorkflowFromJson(t *testing.T) {
	workflow := loadJsonWorkflow(t)

	// Spot check the workflow's BagIt profile.
	pathToProfile := filepath.Join(util.ProjectRoot(), "profiles", "aptrust-v2.2.json")
	aptProfile, err := core.BagItProfileLoad(pathToProfile)
	require.Nil(t, err)

	assert.Equal(t, 14, len(workflow.BagItProfile.Tags))
	assert.EqualValues(t, aptProfile.ManifestsAllowed, workflow.BagItProfile.ManifestsAllowed)
	assert.EqualValues(t, aptProfile.AcceptBagItVersion, workflow.BagItProfile.AcceptBagItVersion)
	assert.EqualValues(t, aptProfile.BagItProfileInfo.BagItProfileIdentifier, workflow.BagItProfile.BagItProfileInfo.BagItProfileIdentifier)

	// Make sure the workflow has storage services
	require.NotEmpty(t, workflow.StorageServices)
	assert.Equal(t, 1, len(workflow.StorageServices))
	assert.Equal(t, "minioadmin", workflow.StorageServices[0].Login)
	assert.Equal(t, "minioadmin", workflow.StorageServices[0].Password)
}

func TestWorkflowFromJob(t *testing.T) {
	defer core.ClearDartTable()

	job := getTestJob(t)
	require.NotNil(t, job.BagItProfile)
	require.NotNil(t, job.PackageOp)
	require.NotEmpty(t, job.UploadOps)

	// Should get error here because BagItProfile is not in database
	workflow, err := core.WorkFlowFromJob(job)
	require.NotNil(t, err)
	require.Nil(t, workflow)

	// Save the profile, and then there should be no error.
	require.Nil(t, core.ObjSave(job.BagItProfile))

	workflow, err = core.WorkFlowFromJob(job)
	require.Nil(t, err)
	require.NotNil(t, workflow)

	assert.True(t, util.LooksLikeUUID(workflow.ID))

	require.NotNil(t, workflow.BagItProfile)
	assert.Equal(t, job.BagItProfile.ID, workflow.BagItProfile.ID)
	assert.Equal(t, job.PackageOp.PackageFormat, workflow.PackageFormat)

	assert.NotEmpty(t, workflow.StorageServices)
	assert.Equal(t, len(job.UploadOps), len(workflow.StorageServices))
	for i, op := range job.UploadOps {
		assert.Equal(t, op.StorageService.ID, workflow.StorageServices[i].ID)
	}
}

func TestWorkflowValidate(t *testing.T) {
	workflow := loadJsonWorkflow(t)
	workflow.BagItProfile.ManifestsAllowed = make([]string, 0)
	workflow.StorageServices[0].Host = ""
	assert.False(t, workflow.Validate())
	assert.Equal(t, 2, len(workflow.Errors))
	assert.Equal(t, "StorageService requires a hostname or IP address.", workflow.Errors["Local Test Receiving Bucket.StorageService.Host"])
	assert.Equal(t, "Profile must allow at least one manifest algorithm.", workflow.Errors["BagItProfile.ManifestsAllowed"])
}

func TestWorkflowLoadSaveDelete(t *testing.T) {
	defer core.ClearDartTable()
	workflow := loadJsonWorkflow(t)

	idFor := make(map[string]string)

	// Create and save five workflows.
	for i := 1; i < 6; i++ {
		w := workflow.Copy()
		w.ID = uuid.NewString()
		w.Name = fmt.Sprintf("Workflow %d", i)
		require.NoError(t, core.ObjSave(w))
		idFor[w.Name] = w.ID
	}

	// List all
	result := core.ObjList(constants.TypeWorkflow, "obj_name", 20, 0)
	require.Nil(t, result.Error)
	assert.Equal(t, 5, len(result.Workflows))
	assert.Equal(t, "Workflow 1", result.Workflows[0].Name)
	assert.Equal(t, "Workflow 5", result.Workflows[4].Name)

	// Find one
	result = core.ObjFind(idFor["Workflow 2"])
	require.Nil(t, result.Error)
	assert.NotNil(t, result.Workflow())
	assert.Equal(t, "Workflow 2", result.Workflow().Name)

	// Delete one
	w := result.Workflow()
	require.NoError(t, core.ObjDelete(w))
}

func TestWorkflowExportJson(t *testing.T) {
	defer core.ClearDartTable()
	workflow := loadJsonWorkflow(t)
	jsonBytes, err := workflow.ExportJson()
	assert.Nil(t, err)
	assert.NotEmpty(t, jsonBytes)
	assert.Equal(t, 10036, len(jsonBytes))
}

func TestWorkflowHasPlaintextPasswords(t *testing.T) {
	workflow := loadJsonWorkflow(t)
	// env passwords come from the environment,
	// so they're not embedded as plain text in
	// the JSON export.
	for i := range workflow.StorageServices {
		ss := workflow.StorageServices[i]
		ss.Password = "env:AWS_SECRET_KEY_ID"
	}
	assert.False(t, workflow.HasPlaintextPasswords())

	// Empty passwords are safe to put in the export
	// JSON. We assume a server admin will fill them
	// in when setting up the DART Runner workflow
	// on the server.
	for i := range workflow.StorageServices {
		ss := workflow.StorageServices[i]
		ss.Password = ""
	}
	assert.False(t, workflow.HasPlaintextPasswords())

	// Now this here is a no-no. Plain text password
	// will be exported in the JSON for all to see.
	for i := range workflow.StorageServices {
		ss := workflow.StorageServices[i]
		ss.Password = "secret"
	}
	assert.True(t, workflow.HasPlaintextPasswords())
}
