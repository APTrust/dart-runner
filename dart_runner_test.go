package main_test

import (
	"path"
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	stdout, stderr, exitCode := util.ExecCommand(buildScript, nil)
	assert.NotEmpty(t, stdout)
	assert.Equal(t, 0, exitCode, stderr)
	if exitCode == 0 {
		setupSucceeded = true
	}
}

func runner() string {
	return path.Join(util.ProjectRoot(), "dist", "dart-runner")
}

func TestHelp(t *testing.T) {
	Setup(t)
	command := runner()
	args := []string{"--help"}
	stdout, stderr, exitCode := util.ExecCommand(command, args)
	assert.Contains(t, string(stdout), "DART Runner: Bag and ship files from the command line")
	assert.Empty(t, stderr)
	require.Equal(t, 0, exitCode)
}

func TestVersion(t *testing.T) {
	Setup(t)
	command := runner()
	args := []string{"--version"}
	stdout, stderr, exitCode := util.ExecCommand(command, args)
	assert.Contains(t, string(stdout), "DART Runner")
	assert.Contains(t, string(stdout), "Build")
	assert.Empty(t, stderr)
	require.Equal(t, 0, exitCode)
}

func TestJob(t *testing.T) {

}

func TestJobParams(t *testing.T) {

}

func TestWorkflowBatch(t *testing.T) {

}
