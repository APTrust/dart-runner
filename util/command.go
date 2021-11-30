package util

import (
	"bytes"
	"os/exec"
)

// ExecCommand executes a command and returns the output of STDOUT and STDERR
// as well as the exit code.
func ExecCommand(command string, args []string) (stdout, stderr []byte, exitCode int) {
	cmd := exec.Command(command, args...)
	var outBuffer, errBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer
	cmd.Run()
	exitCode = cmd.ProcessState.ExitCode()
	return outBuffer.Bytes(), errBuffer.Bytes(), exitCode
}
