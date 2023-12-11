package util

import (
	"bytes"
	"os/exec"
	"runtime"
)

const STDIN_ERROR = -10000

// ExecCommand executes a command and returns the output of STDOUT and STDERR
// as well as the exit code.
//
// os/exec has some problems when piping stdin. These show up on Travis/Linux,
// but not Mac. See https://github.com/golang/go/issues/9307
func ExecCommand(command string, args []string, env []string, stdinData []byte) (stdout, stderr []byte, exitCode int) {
	cmd := exec.Command(command, args...)
	if runtime.GOOS == "windows" {
		windowsArgs := append([]string{"/C", command}, args...)
		cmd = exec.Command("cmd", windowsArgs...)
	}
	var outBuffer, errBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer
	cmd.Env = env

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

	cmd.Run()
	exitCode = cmd.ProcessState.ExitCode()
	return outBuffer.Bytes(), errBuffer.Bytes(), exitCode
}
