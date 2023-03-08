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
	Cleanup      bool
	Concurrency  int
	SuccessCount int
	FailureCount int
	parseError   error
	jobChannel   chan *Job
	waitGroup    sync.WaitGroup
	outMutex     sync.Mutex
	errMutex     sync.Mutex
	fCountMutex  sync.Mutex
	sCountMutex  sync.Mutex
}

// NewWorkflowRunner creates a new WorkFlowRunner object. Param workflowFile
// is the path the JSON file that contains a description of the workflow.
// Param csvFile is the path to the file contains a CSV list of directories
// to package. (That file also contains tag values for each package.)
// Param outputDir is the path to the directory into which the packages
// should be written. Param concurrency is the number of jobs to run in
// parallel.
//
// This creates (1 * concurrency) goroutines to do the work. You probably
// shouldn't set concurrency too high because bagging and other forms of
// packaging do a lot of disk reading and writing. Concurrency significantly
// above 2 will probably lead to disk thrashing.
func NewWorkflowRunner(workflowFile, csvFile, outputDir string, cleanup bool, concurrency int) (*WorkflowRunner, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("Concurrency must be >= 1.")
	}
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

	// Create the runner. Note that the channel size is limited
	// because the job objects pushed into the channel can use
	// 10-20k of memory each. If you're running a workflow on
	// 5,000 items, you don't want 50-100 MB of data sitting in
	// memory waiting to be processed. Better to call jobParams.ToJob()
	// just before the job is going to be executed. The 10-20k
	// goes out of scope as soon as the job completes.
	runner := &WorkflowRunner{
		Workflow:    workflow,
		CSVFile:     workflowCSVFile,
		OutputDir:   outputDir,
		Cleanup:     cleanup,
		Concurrency: concurrency,
		jobChannel:  make(chan *Job, concurrency*2),
	}
	// Create one or more workers to run jobs.
	for i := 0; i < concurrency; i++ {
		go runner.runAsync()
	}
	return runner, nil
}

// Run runs the workflow, writing one line of JSON output per job
// to STDOUT. The output is a serialized JobResult object. Errors
// will be written to STDERR, though there **should** also be
// JobResult written to STDOUT if a job fails.
func (r *WorkflowRunner) Run() int {
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
	return r.getExitCode()
}

// runAsync listens for new jobs on on a go channel and runs
// those jobs as they appear. It should run up to concurrency jobs
// at once.
func (r *WorkflowRunner) runAsync() {
	for job := range r.jobChannel {
		retVal := RunJob(job, r.Cleanup, false)
		if retVal == constants.ExitOK {
			r.sCountMutex.Lock()
			r.SuccessCount++
			r.sCountMutex.Unlock()
		} else {
			r.fCountMutex.Lock()
			r.FailureCount++
			r.fCountMutex.Unlock()
		}
		r.writeResult(job)
		r.waitGroup.Done()
	}
}

func (r *WorkflowRunner) getJobParams(entry *WorkflowCSVEntry) *JobParams {
	return NewJobParams(
		r.Workflow.Copy(),
		entry.BagName,
		path.Join(r.OutputDir, entry.BagName),
		[]string{entry.RootDir},
		entry.Tags)
}

func (r *WorkflowRunner) getExitCode() int {
	if r.parseError != nil {
		errMsg := fmt.Sprintf("Error parsing CSV batch file: %s", r.parseError.Error())
		r.writeStdErr(errMsg)
		return constants.ExitRuntimeErr
	}
	if r.FailureCount > 0 {
		errMsg := fmt.Sprintf("%d job(s) failed", r.FailureCount)
		r.writeStdErr(errMsg)
		return constants.ExitRuntimeErr
	}
	return constants.ExitOK
}

// writeResult writes the result of a job to STDOUT and/or STDERR
func (r *WorkflowRunner) writeResult(job *Job) {
	stdoutMessage, stderrMessage := job.GetResultMessages()
	if len(stdoutMessage) > 0 {
		r.writeStdOut(stdoutMessage)
	}
	if len(stderrMessage) > 0 {
		r.writeStdErr(stderrMessage)
	}
}

// writeStdOut safely writes to STDOUT from concurrent go routines.
func (r *WorkflowRunner) writeStdOut(msg string) {
	r.outMutex.Lock()
	fmt.Println(msg)
	r.outMutex.Unlock()
}

// writeStdErr safely writes to STDOUT from concurrent go routines.
func (r *WorkflowRunner) writeStdErr(msg string) {
	r.errMutex.Lock()
	fmt.Fprintln(os.Stderr, msg)
	r.errMutex.Unlock()
}
