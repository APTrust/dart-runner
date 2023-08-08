package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
)

func getJobForJobResult() *core.Job {
	job := core.NewJob()
	job.PackageOp = core.NewPackageOperation("my_bag.tar", "bags", []string{})
	job.ValidationOp = core.NewValidationOperation("bags/my_bag.tar")
	job.UploadOps = make([]*core.UploadOperation, 3)
	for i := 0; i < 3; i++ {
		job.UploadOps[i] = core.NewUploadOperation(core.NewStorageService(), []string{})
	}
	job.ByteCount = 12345
	job.PayloadFileCount = 16
	return job
}

func TestJobResult(t *testing.T) {
	job := getJobForJobResult()
	jobResult := core.NewJobResult(job)
	assert.Equal(t, job.Name(), jobResult.JobName)
	assert.Equal(t, job.ByteCount, jobResult.PayloadByteCount)
	assert.Equal(t, job.PayloadFileCount, jobResult.PayloadFileCount)
	assert.NotNil(t, jobResult.PackageResult)
	assert.NotNil(t, jobResult.ValidationResult)
	assert.Equal(t, 3, len(jobResult.UploadResults))
	assert.Equal(t, 0, len(jobResult.ValidationErrors))
	assert.True(t, jobResult.Succeeded)

	job.Errors["oops"] = "Britney did it again."
	jobResult = core.NewJobResult(job)
	assert.False(t, jobResult.Succeeded)
	assert.Equal(t, 1, len(jobResult.ValidationErrors))

	jsonStr, err := jobResult.ToJson()
	assert.Nil(t, err)
	assert.True(t, len(jsonStr) > 100)
}
