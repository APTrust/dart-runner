package core

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

// WorkflowRunner runs all jobs in CSVFile through Workflow.
type WorkflowRunner struct {
	Workflow      *Workflow
	CSVFile       *WorkflowCSVFile
	OutputDir     string
	Cleanup       bool
	SkipArtifacts bool
	Concurrency   int
	SuccessCount  int
	FailureCount  int
	parseError    error
	jobChannel    chan *Job
	waitGroup     sync.WaitGroup
	outMutex      sync.Mutex
	errMutex      sync.Mutex
	fCountMutex   sync.Mutex
	sCountMutex   sync.Mutex
	stdErrWriter  *bytes.Buffer
	stdOutWriter  *bytes.Buffer
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
//
// Note that this constructor is for running a workflow without a GUI.
// See NewWorkflowRunnerWithMessageChannel() to run a workflow that sends
// a running stream of status events back to the GUI.
//
// Param workflowFile is the path to a file containing a JSON representation of
// the workflow you want to run.
//
// csvFile is the path to the csv file containing info about what items
// to bag and what tag values to apply to each bag. See
// https://aptrust.github.io/dart-docs/users/workflows/batch_jobs/
// for a description of the csv file and an example of what it looks like.
// This should be an absolute path.
//
// outputDir is the path the output directory where DART will write
// the bags it creates. This should be an absolute path.
//
// cleanup describes whether or not DART should delete the bags it
// creates after successful upload.
//
// concurrency describes how many jobs you want to run at once. In some
// environments, you can set this to 2 or more to get better throughput.
// (E.g. if you're reading from and writing to network attached storage
// on a high-performance file server). But in most cases where you're
// reading from and writing to a single local disk, you'll want to set
// this to 1.
func NewWorkflowRunner(workflowFile, csvFile, outputDir string, cleanup, skipArtifacts bool, concurrency int) (*WorkflowRunner, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("concurrency must be >= 1")
	}
	if !util.FileExists(outputDir) {
		return nil, fmt.Errorf("output directory '%s' does not exist; you must create it first", outputDir)
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
		Workflow:      workflow,
		CSVFile:       workflowCSVFile,
		OutputDir:     outputDir,
		Cleanup:       cleanup,
		SkipArtifacts: skipArtifacts,
		Concurrency:   concurrency,
		jobChannel:    make(chan *Job, concurrency*2),
	}
	// Create one or more workers to run jobs.
	for i := 0; i < concurrency; i++ {
		go runner.listenForJobs(nil)
	}
	return runner, nil
}

// NewWorkflowRunnerWithMessageChannel creates a new workflow runner
// that sends status events back to the GUI. These events tell the user
// which files are being packaged and validated, and they update the
// packaging, validation, and upload progress bars.
//
// This constructor force concurrency to 1 because at the moment,
// the UI is capable of reporting on only one process at a time.
//
// Param workflowID is the id of the workflow you want to run.
//
// csvFile is the path to the csv file containing info about what items
// to bag and what tag values to apply to each bag. See
// https://aptrust.github.io/dart-docs/users/workflows/batch_jobs/
// for a description of the csv file and an example of what it looks like.
// This should be an absolute path.
//
// outputDir is the path the output directory where DART will write
// the bags it creates. This should be an absolute path.
//
// cleanup describes whether or not DART should delete the bags it
// creates after successful upload.
//
// messageChannel is a channel to send status/progress messages back
// to the front end, so the user can see what's happening.
func NewWorkflowRunnerWithMessageChannel(workflowID, csvFile, outputDir string, cleanup bool, messageChannel chan *EventMessage) (*WorkflowRunner, error) {
	if !util.FileExists(outputDir) {
		return nil, fmt.Errorf("output directory '%s' does not exist; you must create it first", outputDir)
	}
	workflowCSVFile, err := NewWorkflowCSVFile(csvFile)
	if err != nil {
		return nil, err
	}
	result := ObjFind(workflowID)
	if result.Error != nil {
		return nil, fmt.Errorf("could not load the specified workflow: %v", result.Error)
	}
	workflow := result.Workflow()

	// See note in NewWorkflowRunner above about creating workflow runner.
	runner := &WorkflowRunner{
		Workflow:      workflow,
		CSVFile:       workflowCSVFile,
		OutputDir:     outputDir,
		Cleanup:       cleanup,
		SkipArtifacts: false,
		Concurrency:   1,
		jobChannel:    make(chan *Job, 1),
	}
	go runner.listenForJobs(messageChannel)
	return runner, nil
}

// SetStdOut causes messages to be written to buffer b instead of
// os.Stdout. We use this in testing to capture output that would go
// to os.Stdout. On Windows, redirecting os.Stdout to a pipe using
// os.Pipe() causes some writes to hang forever. This happens deep
// inside Syscall6, so there's nothing we can do about it. The pipe
// works fine on Linux and MacOS, but not on Windows. Hence, this.
func (r *WorkflowRunner) SetStdOut(b *bytes.Buffer) {
	r.stdOutWriter = b
}

// SetStdErr causes messages to be written to buffer b instead of
// os.Stderr. See the doc comments for SetStdOut above for why
// this exists.
func (r *WorkflowRunner) SetStdErr(b *bytes.Buffer) {
	r.stdErrWriter = b
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

// listenForJobs listens for new jobs on on a go channel and runs
// those jobs as they appear. It should run up to
// WorkflowRunner.Concurrency jobs at once.
func (r *WorkflowRunner) listenForJobs(messageChannel chan *EventMessage) {
	for job := range r.jobChannel {
		var retVal int
		if messageChannel != nil {
			retVal = RunJobWithMessageChannel(job, r.Cleanup, messageChannel)
		} else {
			retVal = RunJob(job, r.Cleanup, r.SkipArtifacts, false)
		}
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
		filepath.Join(r.OutputDir, entry.BagName),
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
	defer r.outMutex.Unlock()
	if r.stdOutWriter != nil {
		_, err := fmt.Fprintln(r.stdOutWriter, msg)
		if err != nil {
			Dart.Log.Errorf("Error writing to stdout builder: %v", err.Error())
			Dart.Log.Warningf("stdout message was: %s", msg)
		}
	} else {
		fmt.Println(msg)
	}
}

// writeStdErr safely writes to STDOUT from concurrent go routines.
func (r *WorkflowRunner) writeStdErr(msg string) {
	r.errMutex.Lock()
	defer r.errMutex.Unlock()
	if r.stdErrWriter != nil {
		_, err := fmt.Fprintln(r.stdErrWriter, msg)
		if err != nil {
			Dart.Log.Errorf("Error writing to stderr builder: %v", err.Error())
			Dart.Log.Warningf("stderr message was: %s", msg)
		}
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
}
