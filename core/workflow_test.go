package core_test

import (
	"path"
	"testing"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowFromJson(t *testing.T) {
	pathToFile := path.Join(util.PathToTestData(), "files", "runner_test_workflow.json")
	workflow, err := core.WorkflowFromJson(pathToFile)
	require.Nil(t, err)
	require.NotNil(t, workflow)
	require.NotNil(t, workflow.BagItProfile)

	// Spot check the workflow's BagIt profile.
	pathToProfile := path.Join(util.ProjectRoot(), "profiles", "aptrust-v2.2.json")
	aptProfile, err := bagit.ProfileLoad(pathToProfile)
	require.Nil(t, err)

	assert.Equal(t, 14, len(workflow.BagItProfile.Tags))
	assert.EqualValues(t, aptProfile.ManifestsAllowed, workflow.BagItProfile.ManifestsAllowed)
	assert.EqualValues(t, aptProfile.AcceptBagItVersion, workflow.BagItProfile.AcceptBagItVersion)
	assert.EqualValues(t, aptProfile.BagItProfileInfo.BagItProfileIdentifier, workflow.BagItProfile.BagItProfileInfo.BagItProfileIdentifier)

	// Make sure the workflow has storage services
	require.NotEmpty(t, workflow.StorageServices)
	assert.Equal(t, 1, len(workflow.StorageServices))
	assert.Equal(t, "env:AWS_ACCESS_KEY_ID", workflow.StorageServices[0].Login)
	assert.Equal(t, "env:AWS_SECRET_ACCESS_KEY", workflow.StorageServices[0].Password)
}
