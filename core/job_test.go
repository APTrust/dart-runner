package core_test

import (
	"os"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestJob(t *testing.T) *core.Job {
	workflow := getTestWorkflow(t)
	files := getTestFileList()
	tags := getTestTags()
	params := core.NewJobParams(workflow, "job_unit_test.tar", os.TempDir(), files, tags)
	return params.ToJob()
}

func TestJobValidate(t *testing.T) {
	job := core.NewJob()
	assert.False(t, job.Validate())
	require.Equal(t, 1, len(job.Errors))
	assert.Equal(t, "Job has nothing to package, validate, or upload.", job.Errors["Job"])

	job = getTestJob(t)
	assert.Equal(t, 0, len(job.Errors))

	// BagItJob without Profile should cause error
	job.BagItProfile = nil
	assert.False(t, job.Validate())

	assert.Equal(t, 2, len(job.Errors))
	assert.Equal(t, "BagIt packaging requires a BagItProfile.", job.Errors["Job.Package.BagItProfile"])
	assert.Equal(t, "Validation requires a BagItProfile.", job.Errors["Job.Validate.BagItProfile"])

	// Let's cause some more problems, shall we?
	for _, uploadOp := range job.UploadOps {
		uploadOp.StorageService.Login = ""
	}
	assert.False(t, job.Validate())
	assert.Equal(t, 3, len(job.Errors))
	assert.Equal(t, "StorageService requires a login name or access key id.", job.Errors["StorageService.Login"])

	for _, uploadOp := range job.UploadOps {
		uploadOp.StorageService = nil
	}
	assert.False(t, job.Validate())
	assert.Equal(t, 4, len(job.Errors))
	assert.Equal(t, "UploadOperation requires a StorageService", job.Errors["UploadOp[0].StorageService"])
	assert.Equal(t, "UploadOperation requires a StorageService", job.Errors["UploadOp[1].StorageService"])
}
