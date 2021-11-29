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

// Version value is injected at build time.
var Version string

func main() {
	constants.AppVersion = Version
	exitCode := constants.ExitOK
	options := core.ParseOptions()
	if !options.AreValid() {
		showHelp()
	} else if options.Version {
		showVersion()
	} else if options.JobFilePath != "" {
		exitCode = runJob(options)
	} else if len(options.StdinData) > 0 {
		exitCode = runJobFromJson(options)
	} else {
		exitCode = runWorkflow(options)
	}
	os.Exit(exitCode)
}

func runJob(opts *core.Options) int {
	job, err := core.JobFromJson(opts.JobFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"Error reading job file %s: %s\n",
			opts.JobFilePath, err.Error())
		return constants.ExitRuntimeErr
	}
	return core.RunJob(job, opts.DeleteAfterUpload)
}

func runJobFromJson(opts *core.Options) int {
	//
	// 1. Parse json into JobParams
	// 2. Read workflow from file
	// 3. Create full-fledged JobParams with Workflow object.
	// 4. Create job using JobParams.ToJob()
	// 5. return core.RunJob(), as below
	//
	// --------------------------------------------
	job, err := core.JobFromJson(opts.JobFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"Error reading job file %s: %s\n",
			opts.JobFilePath, err.Error())
		return constants.ExitRuntimeErr
	}
	return core.RunJob(job, opts.DeleteAfterUpload)
}

func runWorkflow(opts *core.Options) int {
	runner, err := core.NewWorkflowRunner(
		opts.WorkflowFilePath,
		opts.BatchFilePath,
		opts.OutputDir,
		opts.DeleteAfterUpload,
		opts.Concurrency,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot start workflow: %s\n", err.Error())
		return constants.ExitRuntimeErr
	}
	return runner.Run()
}

func showHelp() {
	fmt.Println(help)
}

func showVersion() {
	fmt.Println(Version)
}
