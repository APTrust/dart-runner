package core_test

import (
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

	// ----------------------------------------
	// TODO: Travis needs creds to run uploads.
	// ----------------------------------------
	// When running in CI, don't try to upload items to S3,
	// because we don't currently have credentials.
	if util.RunningInCI() {
		fmt.Println("Running in CI. Skipping WorkflowRunner test until we have creds or a minio server.")
		return
	}

	workflowFile := path.Join(util.PathToTestData(), "files", "runner_test_workflow.json")
	batchFile := createBatchFile(t)
	defer runnerCleanup()

	// ----------------------------------------
	// TODO: Travis needs creds to run uploads.
	// ----------------------------------------

	runner, err := core.NewWorkflowRunner(workflowFile, batchFile, os.TempDir(), 3)
	require.Nil(t, err)
	require.NotNil(t, runner)

	// ----------------------------------------
	// TODO: Capture and test STDOUT properly
	// ----------------------------------------

	//origStdout := os.Stdout
	origStderr := os.Stderr

	//stdOutReader, stdOutWriter, _ := os.Pipe()
	stdErrReader, stdErrWriter, _ := os.Pipe()
	//os.Stdout = stdOutWriter
	os.Stderr = stdErrWriter

	defer func() {
		//os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	retVal := runner.Run()
	assert.Equal(t, retVal, constants.ExitOK)

	//stdOutWriter.Close()
	stdErrWriter.Close()
	//stdOutStr, _ := ioutil.ReadAll(stdOutReader)
	stdErrStr, _ := ioutil.ReadAll(stdErrReader)

	//fmt.Println("STDOUT:", stdOutStr)
	fmt.Println("STDERR:", stdErrStr)

	//assert.NotEmpty(t, stdOutStr)
	assert.Empty(t, string(stdErrStr))
}
