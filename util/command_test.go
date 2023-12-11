package util_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
)

func TestExecCommand(t *testing.T) {
	command := "ls"
	args := []string{"-la"}
	if runtime.GOOS == "windows" {
		command = "dir"
		args = []string{"C:\\"}
	}
	stdout, stderr, exitCode := util.ExecCommand(command, args, os.Environ(), nil)
	assert.NotEmpty(t, stdout, string(stdout))
	assert.Empty(t, stderr, string(stderr))
	assert.Equal(t, 0, exitCode)

	args = []string{"-la", "/does-not-exist"}
	if runtime.GOOS == "windows" {
		args = []string{"C:\\does-not-exist-no-no-no"}
	}
	stdout, stderr, exitCode = util.ExecCommand(command, args, os.Environ(), nil)
	// Windows sends a warning to STDOUT in addition to the error message on stderr
	if runtime.GOOS != "windows" {
		assert.Empty(t, stdout, string(stdout))
	}
	assert.NotEmpty(t, stderr, string(stderr))
	assert.NotEqual(t, 0, exitCode)

	if systemHasAwk() {
		// Note: `awk //` copies stdin to stdout.
		// This tests that stdinData actually gets passed to our command.
		stdinData := []byte("Cletus Spuckler lost a game of tic-tac-toe to a chicken.\n")
		args = []string{"//"}
		stdout, stderr, exitCode = util.ExecCommand("awk", args, os.Environ(), stdinData)
		assert.Equal(t, stdinData, stdout)
		assert.Empty(t, stderr)
		assert.Equal(t, 0, exitCode)
	}
}

func systemHasAwk() bool {
	_, _, exitCode := util.ExecCommand("which", []string{"awk"}, os.Environ(), nil)
	return exitCode == 0
}
