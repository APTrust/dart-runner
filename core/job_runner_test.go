package core_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getRunnerTestJob(t *testing.T) *core.Job {
	workflow := loadJsonWorkflow(t)
	files := []string{
		path.Join(util.PathToTestData(), "files"),
	}
	packageName := "runner_test_bag.tar"
	outputPath := path.Join(os.TempDir(), packageName)
	tags := getTestTags()
	jobParams := core.NewJobParams(workflow, packageName, outputPath, files, tags)
	return jobParams.ToJob()
}

func TestJobRunner(t *testing.T) {
	job := getRunnerTestJob(t)
	defer func() {
		if util.LooksSafeToDelete(job.PackageOp.OutputPath, 12, 3) {
			os.Remove(job.PackageOp.OutputPath)
		}
	}()

	// ----------------------------------------
	// TODO: Travis needs creds to run uploads.
	// ----------------------------------------
	// When running in CI, don't try to upload items to S3,
	// because we don't currently have credentials.
	if util.RunningInCI() {
		fmt.Println("Running in CI. Skipping upload operations in JobRunner test.")
		job.UploadOps = make([]*core.UploadOperation, 0)
	}

	require.True(t, job.Validate())
	retVal := core.RunJob(job)
	assert.Equal(t, constants.ExitOK, retVal)

	assert.True(t, job.PackageOp.Result.Succeeded())
	assert.True(t, job.ValidationOp.Result.Succeeded())
	for _, op := range job.UploadOps {
		fmt.Println(op.Errors)
		assert.True(t, op.Result.Succeeded())
	}
}
