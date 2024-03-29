package core_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createBatchFile(t *testing.T) string {
	batchFile := filepath.Join(util.PathToTestData(), "files", "test_batch.csv")
	batchData, err := util.ReadFile(batchFile)
	require.Nil(t, err)
	// Change placeholder PROJECT_ROOT to the actual project root
	// on this system, so the bagger can find the files it needs to bag.
	batchDataWithPath := strings.ReplaceAll(string(batchData), "PROJECT_ROOT", util.ProjectRoot())
	tmpBatchFile := filepath.Join(os.TempDir(), "test_batch.csv")
	err = os.WriteFile(tmpBatchFile, []byte(batchDataWithPath), 0644)
	require.Nil(t, err)
	return tmpBatchFile
}

func runnerCleanup() {
	os.Remove(filepath.Join(os.TempDir(), "test_batch.csv"))
	os.Remove(filepath.Join(os.TempDir(), "RunnerTestCore.tar"))
	os.Remove(filepath.Join(os.TempDir(), "RunnerTestServer.tar"))
	os.Remove(filepath.Join(os.TempDir(), "RunnerTestUtil.tar"))
}

func TestWorkflowRunner(t *testing.T) {
	workflowFile := filepath.Join(util.PathToTestData(), "files", "runner_test_workflow.json")
	batchFile := createBatchFile(t)
	defer runnerCleanup()

	runner, err := core.NewWorkflowRunner(workflowFile, batchFile, os.TempDir(), false, 3)
	require.Nil(t, err)
	require.NotNil(t, runner)

	// Don't redirect stdout/stderr to pipes on Windows
	// because some write calls will hang forever.
	// Write directly to these buffers instead.
	stdErr := new(bytes.Buffer)
	runner.SetStdErr(stdErr)
	stdOut := new(bytes.Buffer)
	runner.SetStdOut(stdOut)

	retVal := runner.Run()
	assert.Equal(t, retVal, constants.ExitOK)

	stdOutBytes, _ := io.ReadAll(stdOut)
	stdErrBytes, _ := io.ReadAll(stdErr)

	assert.NotEmpty(t, stdOutBytes)
	assert.Empty(t, stdErrBytes)

	// STDOUT should have three JSON objects,
	// each one representing the result of a job.
	// Parse and test these three...
	jsonStr := strings.TrimRight(string(stdOutBytes), "\r\n")
	jsonLines := strings.Split(jsonStr, "\n")
	assert.Equal(t, 3, len(jsonLines), "Workflow should have produced 3 JSON results.")

	for _, line := range jsonLines {
		result := &core.JobResult{}
		err = json.Unmarshal([]byte(line), result)
		require.Nil(t, err)

		assert.True(t, result.PayloadByteCount > 0)
		assert.True(t, result.PayloadFileCount > 0)
		assert.True(t, result.Succeeded)

		assert.True(t, result.PackageResult.Succeeded())
		require.NotEmpty(t, result.ValidationResults)
		assert.True(t, result.ValidationResults[0].Succeeded())

		for _, opResult := range result.UploadResults {
			assert.True(t, len(opResult.RemoteChecksum) >= 32)
			assert.True(t, strings.HasPrefix(opResult.RemoteURL, "s3://localhost:9899/dart-runner.test/RunnerTest"))
			assert.True(t, strings.HasSuffix(opResult.RemoteURL, ".tar"))
		}
	}
}
