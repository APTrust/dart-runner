package core_test

import (
	"path"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadJsonWorkflow(t *testing.T) *core.Workflow {
	pathToFile := path.Join(util.PathToTestData(), "files", "runner_test_workflow.json")
	workflow, err := core.WorkflowFromJson(pathToFile)
	require.Nil(t, err)
	require.NotNil(t, workflow)
	return workflow
}

func TestWorkflowFromJson(t *testing.T) {
	workflow := loadJsonWorkflow(t)
	require.NotNil(t, workflow.BagItProfile)

	require.True(t, workflow.Validate(), workflow.Errors)

	// Spot check the workflow's BagIt profile.
	pathToProfile := path.Join(util.ProjectRoot(), "profiles", "aptrust-v2.2.json")
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
	assert.True(t, workflow.Validate())
	assert.Empty(t, workflow.Errors)

	workflow.BagItProfile.ManifestsAllowed = make([]string, 0)
	workflow.StorageServices[0].Host = ""
	assert.False(t, workflow.Validate())
	assert.Equal(t, 2, len(workflow.Errors))
	assert.Equal(t, "StorageService requires a hostname or IP address.", workflow.Errors["Local Test Receiving Bucket.StorageService.Host"])
	assert.Equal(t, "Profile must allow at least one manifest algorithm.", workflow.Errors["BagItProfile.ManifestsAllowed"])
}
