package main_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests compile a binary of dart runner, call the program like a user
// would, and then test a number of output conditions. These tests rely on the
// setup done by scripts/test.rb, including:
//
// 1. Cleaning out and re-creating the test directories under ~/tmp
// 2. Running a local version of Minio to receive uploads.

var setupAttempted = false
var setupSucceeded = false

func Setup(t *testing.T) {
	if !setupAttempted {
		build(t)
	}
	setupAttempted = true
}

func build(t *testing.T) {
	buildScript := path.Join(util.ProjectRoot(), "scripts", "build.sh")
	stdout, stderr, exitCode := util.ExecCommand(buildScript, nil, nil)
	assert.NotEmpty(t, stdout)
	assert.Equal(t, 0, exitCode, stderr)
	if exitCode == 0 {
		setupSucceeded = true
	}
}

// runner returns the path to the dart-runner executable created by build()
func runner() string {
	return path.Join(util.ProjectRoot(), "dist", "dart-runner")
}

// dirs returns a list of directories commonly used in tests
func dirs(t *testing.T) (filesDir, homeDir, outputDir string) {
	var err error
	filesDir = path.Join(util.ProjectRoot(), "testdata", "files")
	homeDir, err = os.UserHomeDir()
	require.Nil(t, err)

	// NOTE: scripts/test.rb should create this dir before tests start.
	outputDir = path.Join(homeDir, "tmp", "bags")

	return filesDir, homeDir, outputDir
}

func TestHelpCommand(t *testing.T) {
	Setup(t)
	command := runner()
	args := []string{"--help"}
	stdout, stderr, exitCode := util.ExecCommand(command, args, nil)
	assert.Contains(t, string(stdout), "DART Runner: Bag and ship files from the command line")
	assert.Empty(t, stderr)
	require.Equal(t, 0, exitCode)
}

func TestVersionCommand(t *testing.T) {
	Setup(t)
	command := runner()
	args := []string{"--version"}
	stdout, stderr, exitCode := util.ExecCommand(command, args, nil)
	assert.Contains(t, string(stdout), "DART Runner")
	assert.Contains(t, string(stdout), "Build")
	assert.Empty(t, stderr)
	require.Equal(t, 0, exitCode)
}

func TestRunJobCommand(t *testing.T) {
	Setup(t)
	filesDir, homeDir, outputDir := dirs(t)
	jobParamsJson, err := util.ReadFile(path.Join(filesDir, "postbuild_test_params.json"))
	require.Nil(t, err)
	require.True(t, len(jobParamsJson) > 100) // Make sure we read this right.

	workflow := fmt.Sprintf("--workflow=%s/postbuild_test_workflow.json", filesDir)
	output := fmt.Sprintf("--output-dir=%s", outputDir)
	command := fmt.Sprintf("echo '%s' | %s %s %s", string(jobParamsJson), runner(), workflow, output)

	args := []string{
		"-c",
		command,
	}
	stdout, stderr, exitCode := util.ExecCommand("bash", args, jobParamsJson)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr, string(stderr))
	require.Equal(t, 0, exitCode)

	require.True(t, strings.HasPrefix(string(stdout), "{"), "Command output doesn't look like JSON")

	testJsonOutput(t, string(stdout))
	testOutputFile(t, homeDir, "job_params_test.tar")
}

func TestWorkflowBatchCommand(t *testing.T) {
	Setup(t)
	filesDir, homeDir, outputDir := dirs(t)
	command := runner()
	args := []string{
		fmt.Sprintf("--workflow=%s/postbuild_test_workflow.json", filesDir),
		fmt.Sprintf("--batch=%s/postbuild_test_batch.csv", filesDir),
		fmt.Sprintf("--output-dir=%s", outputDir),
		"--concurrency=1",
		"--delete=true",
	}
	stdout, stderr, exitCode := util.ExecCommand(command, args, nil)
	assert.NotEmpty(t, stdout)
	fmt.Println(string(stderr))
	fmt.Println(string(stdout))
	assert.Empty(t, stderr, string(stderr))
	require.Equal(t, 0, exitCode)

	results := strings.Split(string(stdout), "\n")
	assert.Equal(t, 4, len(results))

	for _, data := range results {
		testJsonOutput(t, data)
	}

	outputFiles := []string{
		"RunnerTestBagIt.tar",
		"RunnerTestCore.tar",
		"RunnerTestUtil.tar",
	}
	for _, file := range outputFiles {
		testOutputFile(t, homeDir, file)
	}
}

func testOutputFile(t *testing.T, homeDir, file string) {
	// This directory is also created by scripts/test.rb.
	// Post-build tests upload to the dart-runner.test bucket.
	fullPath := path.Join(homeDir, "tmp", "minio", "dart-runner.test", file)
	require.True(t, util.FileExists(fullPath), fullPath)

	fileInfo, err := os.Stat(fullPath)
	require.Nil(t, err, fullPath)

	// Make sure size is sane and modtime is fresh (so we know file isn't
	// left over from a prior test run... the test script at scripts/test.rb
	// should delete contents of this dir before each run, but let's be sure).
	assert.True(t, fileInfo.Size() > int64(10000), fullPath)
	assert.WithinDuration(t, time.Now(), fileInfo.ModTime(), 30*time.Second)
}

func testJsonOutput(t *testing.T, data string) {
	if len(data) == 0 {
		return // empty newline at end of output
	}
	result := &core.JobResult{}
	err := json.Unmarshal([]byte(data), result)
	require.Nil(t, err)
	assert.NotEmpty(t, result.JobName)
	assert.True(t, result.PayloadFileCount > 0)
	assert.True(t, result.PayloadByteCount > 0)
	assert.NotNil(t, result.PackageResult)
	assert.NotNil(t, result.ValidationResult)
	assert.NotNil(t, result.UploadResults)
	assert.Equal(t, 1, len(result.UploadResults))
	assert.True(t, result.Succeeded)
}
