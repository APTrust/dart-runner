package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
)

//go:embed help.txt
var help string

func main() {
	options := core.ParseOptions()
	fmt.Println(options)
	if !options.AreValid() {
		showHelp()
		os.Exit(constants.ExitOK)
	}
}

func runJob(opts *core.Options) int {

	return constants.ExitOK
}

func runWorkflow(opts *core.Options) int {

	return constants.ExitOK
}

func showHelp() {
	fmt.Println(help)
}
