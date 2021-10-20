package core

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

type Runner struct {
	Job *Job
}

func RunJob(job *Job, deleteOnSuccess bool) int {
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
	if deleteOnSuccess {
		runner.cleanup()
	} else {
		runner.setNoCleanupMessage()
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
	r.setResultFileInfo(op.Result, op.OutputPath, bagger.Errors)
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
	r.setResultFileInfo(op.Result, op.PathToBag, op.Errors)
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
		if op.SourceFiles != nil && len(op.SourceFiles) > 0 {
			r.setResultFileInfo(op.Result, op.SourceFiles[0], op.Errors)
		}
		op.Result.Finish(op.Errors)
		if !ok {
			allSucceeded = false
		}
	}
	return allSucceeded
}

func (r *Runner) setResultFileInfo(opResult *OperationResult, filePath string, errMap map[string]string) {
	opResult.FilePath = filePath
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		errMap["OutputFile.Stat"] = fmt.Sprintf("Can't stat output file at %s: %s", filePath, err.Error())
	} else {
		opResult.FileMTime = fileInfo.ModTime()
		opResult.FileSize = fileInfo.Size()
	}
}

// cleanup cleans up bags from the output directory after successful
// upload.
func (r *Runner) cleanup() {
	// If we didn't create bag, there's nothing to delete
	if r.Job.PackageOp == nil || r.Job.PackageOp.OutputPath == "" {
		return
	}

	bagFile := r.Job.PackageOp.OutputPath

	// If we didn't upload anything, it was just a bagging job.
	// The user probably wants to do something with this bag,
	// since we haven't done anything with it. So don't delete it.
	// This means we don't delete anything for bagging-only,
	// validation-only, or bagging + validation jobs, even when
	// the --delete flag is set to true.
	if r.Job.UploadOps == nil || len(r.Job.UploadOps) == 0 {
		r.Job.PackageOp.Result.Warning = fmt.Sprintf(
			"Output file at %s was not deleted because there was no upload",
			bagFile)
		return
	}

	lastUpload := r.Job.UploadOps[len(r.Job.UploadOps)-1]

	if r.Job.UploadSucceeded() {
		if !util.FileExists(bagFile) {
			return
		}
		if util.IsDirectory(bagFile) {
			lastUpload.Result.Warning = fmt.Sprintf(
				"Output file at %s was not deleted because it is a directory",
				bagFile)
			return
		}
		err := os.Remove(bagFile)
		if err != nil {
			lastUpload.Result.Warning = fmt.Sprintf(
				"Error deleting output file at %s: %s. You should delete this manually.",
				bagFile, err.Error())

		} else {
			lastUpload.Result.Info = fmt.Sprintf(
				"Output file at %s was deleted at %s",
				bagFile, time.Now().Format(time.RFC3339))
		}
	}
}

func (r *Runner) setNoCleanupMessage() {
	if r.Job.UploadOps != nil && len(r.Job.UploadOps) > 0 {
		for _, op := range r.Job.UploadOps {
			op.Result.Info = fmt.Sprintf("Bag file(s) remain at %s.",
				strings.Join(op.SourceFiles, ", "))
		}
	} else if r.Job.ValidationOp != nil && r.Job.ValidationOp.PathToBag != "" {
		r.Job.ValidationOp.Result.Info = fmt.Sprintf(
			"Bag file remains in %s.", r.Job.ValidationOp.PathToBag)
	} else if r.Job.PackageOp == nil || r.Job.PackageOp.OutputPath == "" {
		r.Job.PackageOp.Result.Info = fmt.Sprintf(
			"Bag file remains in %s.", r.Job.PackageOp.OutputPath)
	}
}
