package core

import (
	"fmt"
	"os"

	"github.com/APTrust/dart-runner/constants"
)

type Runner struct {
	Job *Job
}

func RunJob(job *Job) int {
	runner := &Runner{job}
	if !runner.ValidateJob() {
		return constants.ExitRuntimeErr
	}
	if !runner.RunPackageOp() {
		return constants.ExitRuntimeErr
	}
	if !runner.RunValidationOp() {
		return constants.ExitRuntimeErr
	}
	if !runner.RunUploadOps() {
		return constants.ExitRuntimeErr
	}
	return constants.ExitOK
}

func (r *Runner) ValidateJob() bool {
	if !r.Job.Validate() {
		r.PrintErrors(r.Job.Errors)
		return false
	}
	return true
}

func (r *Runner) RunPackageOp() bool {
	if r.Job.PackageOp == nil {
		return true
	}
	// Build the package / bag
	// Set the bag path on the validation op
	return true
}

func (r *Runner) RunValidationOp() bool {
	if r.Job.ValidationOp == nil {
		return true
	}
	// Validate the package / bag
	return true
}

func (r *Runner) RunUploadOps() bool {
	if r.Job.UploadOps == nil || len(r.Job.UploadOps) == 0 {
		return true
	}
	// Run upload ops on sequence. If any fails, continue
	// with remaining uploads and set the error messages.
	return true
}

func (r *Runner) PrintErrors(errors map[string]string) {
	for key, value := range errors {
		fmt.Fprintln(os.Stderr, key, value)
	}
}
