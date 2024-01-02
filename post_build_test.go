package main_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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
	os.Setenv("DART_ENV", "test")
}

func build(t *testing.T) {
	if runtime.GOOS == "windows" {
		buildWindows(t)
	} else {
		buildScript := filepath.Join(util.ProjectRoot(), "scripts", "build.sh")
		stdout, stderr, exitCode := util.ExecCommand(buildScript, nil, os.Environ(), nil)
		assert.NotEmpty(t, stdout)
		assert.Equal(t, 0, exitCode, string(stderr))
		if exitCode == 0 {
			setupSucceeded = true
		}
	}
}

func buildWindows(t *testing.T) {
	buildDir := filepath.Join(util.ProjectRoot(), "dist", "windows")
	require.NoError(t, os.MkdirAll(buildDir, 0755))
	//command := `go build -o dist/windows/dart-runner.exe -ldflags "-X 'main.Version=TEST'" -tags windows`
	version := fmt.Sprintf("DART Runner TEST for WINDOWS (Build TEST %s)", time.Now().Format(time.RFC3339))
	command := "go"
	args := []string{
		"build",
		"-o",
		".\\dist\\windows\\dart-runner.exe",
		"-ldflags",
		fmt.Sprintf("-X 'main.Version=%s'", version),
		"-tags",
		"windows",
	}
	stdout, stderr, exitCode := util.ExecCommand(command, args, os.Environ(), nil)
	assert.Equal(t, 0, exitCode, "stdout: %s \n stderr: %s", string(stdout), string(stderr))
	setupAttempted = true
	setupSucceeded = exitCode == 0
}

// runner returns the path to the dart-runner executable created by build()
func runner() string {
	osName := "windows"
	exeName := "dart-runner"

	switch runtime.GOOS {
	case "linux":
		osName = "linux"
	case "darwin":
		if runtime.GOARCH == "amd-64" {
			osName = "mac-x64"
		} else {
			osName = "mac-arm64"
		}
	case "windows":
		osName = "windows"
		exeName = "dart-runner.exe"
	}
	return filepath.Join(util.ProjectRoot(), "dist", osName, exeName)
}

// When we run post-build tests, DART needs to know it's running in
// a test environment, so it uses an in-memory database instead of
// polluting the on-disk database.
func envForRunner() []string {
	env := os.Environ()
	env = append(env, "DART_ENV=test")
	return env
}

// dirs returns a list of directories commonly used in tests
func dirs(t *testing.T) (filesDir, homeDir, outputDir string) {
	var err error
	filesDir = filepath.Join(util.PathToTestData(), "files")
	homeDir, err = os.UserHomeDir()
	require.Nil(t, err)

	// NOTE: scripts/test.rb should create this dir before tests start.
	outputDir = filepath.Join(homeDir, "tmp", "bags")

	return filesDir, homeDir, outputDir
}

func TestHelpCommand(t *testing.T) {
	Setup(t)
	command := runner()
	args := []string{"--help"}
	stdout, stderr, exitCode := util.ExecCommand(command, args, envForRunner(), nil)
	assert.Contains(t, string(stdout), "DART Runner: Bag and ship files from the command line")
	assert.Empty(t, stderr, string(stderr))
	require.Equal(t, 0, exitCode)
}

func TestVersionCommand(t *testing.T) {
	Setup(t)
	command := runner()
	args := []string{"--version"}
	stdout, stderr, exitCode := util.ExecCommand(command, args, envForRunner(), nil)
	assert.Contains(t, string(stdout), "DART Runner")
	assert.Contains(t, string(stdout), "Build")
	assert.Empty(t, stderr)
	require.Equal(t, 0, exitCode)
}

func TestRunJobCommand(t *testing.T) {
	Setup(t)
	filesDir, homeDir, outputDir := dirs(t)
	command := runner()
	jobParamsJson, err := util.ReadFile(filepath.Join(filesDir, "postbuild_test_params.json"))
	require.Nil(t, err, command)
	require.True(t, len(string(jobParamsJson)) > 100)
	args := []string{
		fmt.Sprintf("--workflow=%s/postbuild_test_workflow.json", filesDir),
		fmt.Sprintf("--output-dir=%s", outputDir),
	}
	stdout, stderr, exitCode := util.ExecCommand(command, args, envForRunner(), jobParamsJson)
	assert.NotEmpty(t, stdout)
	if len(stderr) > 0 {
		fmt.Println("JSON output from failed workflow test:")
		fmt.Println(string(stdout))
	}
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
	stdout, stderr, exitCode := util.ExecCommand(command, args, envForRunner(), nil)
	assert.NotEmpty(t, stdout)

	assert.Empty(t, stderr, string(stderr))
	require.Equal(t, 0, exitCode)

	results := strings.Split(string(stdout), "\n")
	assert.Equal(t, 4, len(results))

	for _, data := range results {
		testJsonOutput(t, data)
	}

	outputFiles := []string{
		"RunnerTestServer.tar",
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

	accessKeyId := "minioadmin"
	secretKey := "minioadmin"
	options := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyId, secretKey, ""),
		Secure: false,
	}
	client, err := minio.New("127.0.0.1:9899", options)
	require.NoError(t, err)

	objInfo, err := client.StatObject(context.Background(), "dart-runner.test", file, minio.StatObjectOptions{})
	require.NoError(t, err)
	require.NotNil(t, objInfo)

	assert.True(t, objInfo.Size > int64(10000), file)
	assert.WithinDuration(t, time.Now(), objInfo.LastModified, 30*time.Second)
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
	assert.NotEmpty(t, result.ValidationResults)
	assert.NotNil(t, result.UploadResults)
	assert.Equal(t, 1, len(result.UploadResults))
	assert.True(t, result.Succeeded)
}
