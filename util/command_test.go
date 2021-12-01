package util_test

import (
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
)

func TestExecCommand(t *testing.T) {
	stdout, stderr, exitCode := util.ExecCommand("ls", []string{"-la"})
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)
	assert.Equal(t, 0, exitCode)

	stdout, stderr, exitCode = util.ExecCommand("ls", []string{"-la", "/does-not-exist"})
	assert.Empty(t, stdout)
	assert.NotEmpty(t, stderr)
	assert.Equal(t, 1, exitCode)
}
