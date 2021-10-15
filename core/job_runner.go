package core

import (
	"github.com/APTrust/dart-runner/bagit"
	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

type Runner struct {
	Job *Job
}

func RunJob(job *Job) int {
	runner := &Runner{job}
	if !runner.ValidateJob() {
		return constants.ExitRuntimeErr
	}

	// ----------------------------------------
	// TODO: Delete bag when job completes.
	// Can do this in a simple defer.
	//
	// TODO: Fill in all available data in
	// OperationResult for each op. Or else
	// get rid of extraneous fields in that
	// struct.
	// ----------------------------------------

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
	//
	// For workflows, we only permit a single directory in op.SourceFiles.
	// Jobs may contain multiple. If there are overlapping directories,
	// we want to make sure their common files are not included twice.
	op := r.Job.PackageOp
	op.Result.Start()
	sourceFiles := make([]*util.ExtendedFileInfo, 0)
	for _, filepath := range op.SourceFiles {
		files, err := util.RecursiveFileList(filepath)
		if err != nil {
			errors := map[string]string{
				"SourceFiles": err.Error(),
			}
			op.Result.Finish(errors)
			return false
		}
		// TODO: Weed out duplicate files.
		sourceFiles = append(sourceFiles, files...)
	}
	bagger := bagit.NewBagger(op.OutputPath, r.Job.BagItProfile, sourceFiles)
	ok := bagger.Run()
	r.Job.ByteCount = bagger.PayloadBytes()
	r.Job.FileCount = bagger.PayloadFileCount()
	op.Result.Finish(bagger.Errors)
	return ok
}

func (r *Runner) RunValidationOp() bool {
	if r.Job.ValidationOp == nil {
		return true
	}
	// Validate the package / bag
	op := r.Job.ValidationOp
	op.Result.Start()
	ok := r.Job.ValidationOp.Validate()
	op.Result.Finish(op.Errors)
	return ok
}

func (r *Runner) RunUploadOps() bool {
	if r.Job.UploadOps == nil || len(r.Job.UploadOps) == 0 {
		return true
	}
	// Run upload ops in sequence. If any fails, continue
	// with remaining uploads.
	allSucceeded := true
	for _, op := range r.Job.UploadOps {
		op.Result.Start()
		ok := op.DoUpload()
		op.Result.Finish(op.Errors)
		if !ok {
			allSucceeded = false
		}
	}
	return allSucceeded
}
