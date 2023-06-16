package util_test

import (
	"os"
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
)

func TestExecCommand(t *testing.T) {
	args := []string{"-la"}
	stdout, stderr, exitCode := util.ExecCommand("ls", args, os.Environ(), nil)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)
	assert.Equal(t, 0, exitCode)

	args = []string{"-la", "/does-not-exist"}
	stdout, stderr, exitCode = util.ExecCommand("ls", args, os.Environ(), nil)
	assert.Empty(t, stdout)
	assert.NotEmpty(t, stderr)
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
