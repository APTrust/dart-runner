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
	require.True(t, job.Validate())
	retVal := core.RunJob(job)
	for k, v := range job.PackageOp.Result.Errors {
		fmt.Println(k, v)
	}
	for k, v := range job.Errors {
		fmt.Println(k, v)
	}
	assert.Equal(t, constants.ExitOK, retVal)
}
