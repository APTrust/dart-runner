package core_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewArtifact(t *testing.T) {
	artifact := core.NewArtifact()
	assert.True(t, util.LooksLikeUUID(artifact.ID))
	assert.NotEmpty(t, artifact.UpdatedAt)
}

func TestNewJobResultArtifact(t *testing.T) {
	job := loadTestJob(t)
	jobResult := core.NewJobResult(job)

	artifact := core.NewJobResultArtifact("blah-blah-blah", jobResult)
	assert.True(t, util.LooksLikeUUID(artifact.ID))
	assert.Equal(t, jobResult.JobID, artifact.JobID)
	assert.Equal(t, "blah-blah-blah", artifact.BagName)
	assert.Equal(t, constants.ItemTypeJobResult, artifact.ItemType)
	assert.Equal(t, fmt.Sprintf("Job Result %s", jobResult.JobName), artifact.FileName)
	assert.Equal(t, constants.FileTypeJsonData, artifact.FileType)
	assert.NotEmpty(t, artifact.UpdatedAt)

	resultJson, err := json.Marshal(jobResult)
	require.NoError(t, err)
	assert.Equal(t, string(resultJson), artifact.RawData)
}

func TestNewManifestArtifact(t *testing.T) {
	artifact := core.NewManifestArtifact("baggy", constants.EmptyUUID, "manny-fest.txt", "this here is the content")
	assert.True(t, util.LooksLikeUUID(artifact.ID))
	assert.Equal(t, constants.EmptyUUID, artifact.JobID)
	assert.Equal(t, "baggy", artifact.BagName)
	assert.Equal(t, constants.ItemTypeManifest, artifact.ItemType)
	assert.Equal(t, "manny-fest.txt", artifact.FileName)
	assert.Equal(t, constants.FileTypeManifest, artifact.FileType)
	assert.NotEmpty(t, artifact.UpdatedAt)
}

func TestNewTagFileArtifact(t *testing.T) {
	artifact := core.NewTagFileArtifact("baggsy", constants.EmptyUUID, "tag-file.txt", "this here is the content")
	assert.True(t, util.LooksLikeUUID(artifact.ID))
	assert.Equal(t, constants.EmptyUUID, artifact.JobID)
	assert.Equal(t, "baggsy", artifact.BagName)
	assert.Equal(t, constants.ItemTypeTagFile, artifact.ItemType)
	assert.Equal(t, "tag-file.txt", artifact.FileName)
	assert.Equal(t, constants.FileTypeTag, artifact.FileType)
	assert.NotEmpty(t, artifact.UpdatedAt)
}
