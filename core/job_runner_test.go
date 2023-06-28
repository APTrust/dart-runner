package core_test

import (
	"os"
	"path"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getRunnerTestJob(t *testing.T, bagName string) *core.Job {
	workflow := loadJsonWorkflow(t)
	files := []string{
		path.Join(util.PathToTestData(), "files"),
	}
	outputPath := path.Join(os.TempDir(), bagName)
	tags := getTestTags()
	jobParams := core.NewJobParams(workflow, bagName, outputPath, files, tags)
	return jobParams.ToJob()
}

func testJobRunner(t *testing.T, bagName string, withCleanup bool) {
	job := getRunnerTestJob(t, bagName)
	defer func() {
		if withCleanup && util.LooksSafeToDelete(job.PackageOp.OutputPath, 12, 3) {
			os.Remove(job.PackageOp.OutputPath)
		}
	}()

	require.True(t, job.Validate(), job.Errors)
	retVal := core.RunJob(job, withCleanup, false)
	assert.Equal(t, constants.ExitOK, retVal)

	assert.True(t, job.PackageOp.Result.Succeeded())
	assert.True(t, job.ValidationOp.Result.Succeeded())
	for _, op := range job.UploadOps {
		assert.True(t, op.Result.Succeeded())
	}

	lastUpload := job.UploadOps[len(job.UploadOps)-1]
	if withCleanup {
		assert.Contains(t, lastUpload.Result.Info, "was deleted at")
	} else {
		assert.Contains(t, lastUpload.Result.Info, "Bag file(s) remain")
	}
}

func TestJobRunnerWithCleanup(t *testing.T) {
	testJobRunner(t, "bag_with_cleanup.tar", true)
}

func TestJobRunnerNoCleanup(t *testing.T) {
	testJobRunner(t, "bag_without_cleanup.tar", false)
}
