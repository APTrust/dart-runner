package core

import (
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

// WorkflowRunner runs all jobs in CSVFile through Workflow.
type WorkflowRunner struct {
	Workflow     *Workflow
	CSVFile      *WorkflowCSVFile
	OutputDir    string
	Concurrency  int
	SuccessCount int
	FailureCount int
	parseError   error
	jobChannel   chan *Job
	waitGroup    sync.WaitGroup
}

func NewWorkflowRunner(workflowFile, csvFile, outputDir string, concurrency int) (*WorkflowRunner, error) {
	if !util.FileExists(outputDir) {
		return nil, fmt.Errorf("Output directory '%s' does not exist. You must create it first.", outputDir)
	}
	workflowCSVFile, err := NewWorkflowCSVFile(csvFile)
	if err != nil {
		return nil, err
	}
	workflow, err := WorkflowFromJson(workflowFile)
	if err != nil {
		return nil, err
	}
	return &WorkflowRunner{
		Workflow:    workflow,
		CSVFile:     workflowCSVFile,
		OutputDir:   outputDir,
		Concurrency: concurrency,
		jobChannel:  make(chan *Job, concurrency),
	}, nil
}

func (r *WorkflowRunner) Run() int {
	go r.runAsync()
	for {
		entry, err := r.CSVFile.ReadNext()
		if err == io.EOF {
			break
		} else if err != nil {
			r.parseError = err
			break
		}
		jobParams := r.getJobParams(entry)
		r.waitGroup.Add(1)
		r.jobChannel <- jobParams.ToJob()
	}
	r.waitGroup.Wait()

	if r.parseError != nil || r.FailureCount > 0 {
		return constants.ExitRuntimeErr
	}
	return constants.ExitOK
}

func (r *WorkflowRunner) runAsync() {
	for job := range r.jobChannel {
		retVal := RunJob(job)
		if retVal == constants.ExitOK {
			r.SuccessCount++
		} else {
			r.FailureCount++
		}
		r.waitGroup.Done()
	}
}

func (r *WorkflowRunner) getJobParams(entry *WorkflowCSVEntry) *JobParams {
	return NewJobParams(
		r.Workflow,
		entry.BagName,
		path.Join(r.OutputDir, entry.BagName),
		[]string{entry.RootDir},
		entry.Tags)
}

func (r *WorkflowRunner) getExitCode() int {
	if r.parseError != nil {
		errMsg := fmt.Sprintf("Error parsing CSV batch file: %s", r.parseError.Error())
		fmt.Fprintf(os.Stderr, errMsg)
		return constants.ExitRuntimeErr
	}
	if r.FailureCount > 0 {
		errMsg := fmt.Sprintf("%d job(s) failed", r.FailureCount)
		fmt.Fprintf(os.Stderr, errMsg)
		return constants.ExitRuntimeErr
	}
	return constants.ExitOK
}
