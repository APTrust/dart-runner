package util

import (
	"bytes"
	"os/exec"
)

const STDIN_ERROR = -10000

// ExecCommand executes a command and returns the output of STDOUT and STDERR
// as well as the exit code.
//
// os/exec has some problems when piping stdin. These show up on Travis/Linux,
// but not Mac. See https://github.com/golang/go/issues/9307
func ExecCommand(command string, args []string, stdinData []byte) (stdout, stderr []byte, exitCode int) {
	cmd := exec.Command(command, args...)
	var outBuffer, errBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	if len(stdinData) > 0 {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, nil, STDIN_ERROR
		}
		_, err = stdin.Write(stdinData)
		if err != nil {
			return nil, nil, STDIN_ERROR
		}
		stdin.Close()
	}

	cmd.Start()
	cmd.Wait()
	exitCode = cmd.ProcessState.ExitCode()
	return outBuffer.Bytes(), errBuffer.Bytes(), exitCode
}
