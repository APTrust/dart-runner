package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
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
	return core.RunJob(job, opts.DeleteAfterUpload, true)
}

func runJobFromJson(opts *core.Options) int {
	params, err := initParams(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating job: %s\n", err.Error())
		return constants.ExitRuntimeErr
	}
	return core.RunJob(params.ToJob(), opts.DeleteAfterUpload, true)
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

func initParams(opts *core.Options) (*core.JobParams, error) {
	if !util.FileExists(opts.OutputDir) {
		return nil, fmt.Errorf("Output directory '%s' does not exist. You must create it first.", opts.OutputDir)
	}
	workflow, err := core.WorkflowFromJson(opts.WorkflowFilePath)
	if err != nil {
		return nil, err
	}
	partialParams := &core.JobParams{}
	err = json.Unmarshal(opts.StdinData, partialParams)
	if err != nil {
		return nil, fmt.Errorf("JobParams JSON: %s", err.Error())
	}
	params := core.NewJobParams(
		workflow,
		partialParams.PackageName,
		path.Join(opts.OutputDir, partialParams.PackageName),
		partialParams.Files,
		partialParams.Tags)
	return params, nil
}

func showHelp() {
	fmt.Println(help)
}

func showVersion() {
	fmt.Println(Version)
}
