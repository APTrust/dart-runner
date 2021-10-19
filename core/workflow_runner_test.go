package core_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createBatchFile(t *testing.T) string {
	batchFile := path.Join(util.PathToTestData(), "files", "test_batch.csv")
	batchData, err := util.ReadFile(batchFile)
	require.Nil(t, err)
	// Change placeholder PROJECT_ROOT to the actual project root
	// on this system, so the bagger can find the files it needs to bag.
	batchDataWithPath := strings.ReplaceAll(string(batchData), "PROJECT_ROOT", util.ProjectRoot())
	tmpBatchFile := path.Join(os.TempDir(), "test_batch.csv")
	err = os.WriteFile(tmpBatchFile, []byte(batchDataWithPath), 0644)
	require.Nil(t, err)
	return tmpBatchFile
}

func runnerCleanup() {
	os.Remove(path.Join(os.TempDir(), "test_batch.csv"))
	os.Remove(path.Join(os.TempDir(), "RunnerTestCore.tar"))
	os.Remove(path.Join(os.TempDir(), "RunnerTestBagIt.tar"))
	os.Remove(path.Join(os.TempDir(), "RunnerTestUtil.tar"))
}

func TestWorkflowRunner(t *testing.T) {
	workflowFile := path.Join(util.PathToTestData(), "files", "runner_test_workflow.json")
	batchFile := createBatchFile(t)
	defer runnerCleanup()

	runner, err := core.NewWorkflowRunner(workflowFile, batchFile, os.TempDir(), 3)
	require.Nil(t, err)
	require.NotNil(t, runner)

	// ----------------------------------------
	// TODO: Capture and test STDOUT properly
	// ----------------------------------------

	origStdout := os.Stdout
	origStderr := os.Stderr

	stdOutReader, stdOutWriter, _ := os.Pipe()
	stdErrReader, stdErrWriter, _ := os.Pipe()
	os.Stdout = stdOutWriter
	os.Stderr = stdErrWriter

	retVal := runner.Run()
	assert.Equal(t, retVal, constants.ExitOK)

	stdOutWriter.Close()
	stdErrWriter.Close()
	stdOutBytes, _ := ioutil.ReadAll(stdOutReader)
	stdErrBytes, _ := ioutil.ReadAll(stdErrReader)

	os.Stdout = origStdout
	os.Stderr = origStderr

	fmt.Println("STDOUT:", string(stdOutBytes))
	fmt.Println("STDERR:", string(stdErrBytes))

	assert.NotEmpty(t, stdOutBytes)
	assert.Empty(t, string(stdErrBytes))

	// STDOUT should have three JSON objects,
	// each one representing the result of a job.
	// Parse and test these three...
	jsonStr := strings.TrimRight(string(stdOutBytes), "\r\n")
	jsonLines := strings.Split(jsonStr, util.NewLine())
	assert.Equal(t, 3, len(jsonLines), "Workflow should have produced 3 JSON results.")

	for _, line := range jsonLines {
		result := &core.JobResult{}
		err = json.Unmarshal([]byte(line), result)
		require.Nil(t, err)

		assert.True(t, result.PayloadByteCount > 0)
		assert.True(t, result.PayloadFileCount > 0)
		assert.True(t, result.Succeeded)

		for _, opResult := range result.Results {
			if opResult.Operation == "upload" {
				assert.True(t, len(opResult.RemoteChecksum) >= 32)
				assert.True(t, strings.HasPrefix(opResult.RemoteURL, "s3://localhost:9899/dart-runner.test/RunnerTest"))
				assert.True(t, strings.HasSuffix(opResult.RemoteURL, ".tar"))
			}
		}
	}
	//assert.True(t, false)
}

// func formatJsonOutput(str string) string {
// 	splitAt := fmt.Sprintf("}%s{", util.NewLine())
// 	parts := strings.Split(str, splitAt)
// 	return fmt.Sprintf("[%s]", strings.Join(parts, "},{"))
// }
