package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

type Runner struct {
	Job            *Job
	MessageChannel chan *EventMessage
}

// RunJobWithMessageChannel runs a job and pumps progress details
// into the messageChannel. That channel passes messages back to
// the front end UI. This returns the job's exit code, where zero
// means success and non-zero is failure. See the exit codes defined
// in constants for more info.
func RunJobWithMessageChannel(job *Job, deleteOnSuccess bool, messageChannel chan *EventMessage) int {
	runner := &Runner{
		Job:            job,
		MessageChannel: messageChannel,
	}
	if !runner.ValidateJob() {
		runner.writeExitMessagesAndSaveResults()
		return constants.ExitRuntimeErr
	}
	// DART calls RunJobWithMessageChannel instead of RunJob. For DART,
	// we always want to save artifacts. They go into the SQLite DB.
	if !runner.RunPackageOp(false) {
		runner.writeStageOutcome(constants.StagePackage, runner.Job.PackageOp.Result.Info, false)
		runner.writeExitMessagesAndSaveResults()
		return constants.ExitRuntimeErr
	}
	runner.writeStageOutcome(constants.StagePackage, runner.Job.PackageOp.Result.Info, true)

	if !runner.RunValidationOp() {
		runner.writeStageOutcome(constants.StageValidation, runner.Job.ValidationOp.Result.Info, false)
		runner.writeExitMessagesAndSaveResults()
		return constants.ExitRuntimeErr
	}
	runner.writeStageOutcome(constants.StageValidation, runner.Job.ValidationOp.Result.Info, true)

	if !runner.RunPostValidationOps() {
		runner.writeStageOutcome(constants.StageUpload, "One or more uploads failed", false)
		runner.writeExitMessagesAndSaveResults()
		return constants.ExitRuntimeErr
	}

	if !runner.RunUploadOps() {
		runner.writeStageOutcome(constants.StageUpload, "One or more uploads failed", false)
		runner.writeExitMessagesAndSaveResults()
		return constants.ExitRuntimeErr
	}
	if deleteOnSuccess {
		runner.cleanup()
	} else {
		runner.setNoCleanupMessage()
	}
	runner.writeStageOutcome(constants.StageUpload, "All uploads succeeded", true)

	runner.writeExitMessagesAndSaveResults()

	return constants.ExitOK
}

// This writes exit messages to the message channel, and saves the job and
// job result records to the DB.
func (r *Runner) writeExitMessagesAndSaveResults() {
	if r.MessageChannel == nil {
		panic("JobRunner.MessageChannel is nil")
	}
	err := ObjSave(r.Job)
	if err != nil {
		Dart.Log.Warningf("Error saving Job '%s' after running: %s", r.Job.Name, err.Error())
	}
	result := NewJobResult(r.Job)
	bagName := ""
	if r.Job.PackageOp != nil {
		bagName = r.Job.PackageOp.PackageName
	}
	artifact := NewJobResultArtifact(bagName, result)
	err = ArtifactSave(artifact)
	if err != nil {
		Dart.Log.Warningf("Error saving result artifact for job '%s' after running: %s", r.Job.Name, err.Error())
	}
	r.MessageChannel <- FinishEvent(result)
}

func (r *Runner) writeStageOutcome(stage, message string, succeeded bool) {
	if r.MessageChannel == nil {
		panic("JobRunner.MessageChannel is nil")
	}
	status := constants.StatusFailed
	if succeeded {
		status = constants.StatusSuccess
	}
	eventMessage := &EventMessage{
		EventType: constants.EventTypeFinish,
		Stage:     stage,
		Status:    status,
		Message:   message,
	}
	r.MessageChannel <- eventMessage
}

func RunJob(job *Job, deleteOnSuccess, skipArtifacts, printOutput bool) int {
	runner := &Runner{Job: job}
	if !runner.ValidateJob() {
		runner.printExitMessages()
		return constants.ExitRuntimeErr
	}
	if !runner.RunPackageOp(skipArtifacts) {
		runner.printExitMessages()
		return constants.ExitRuntimeErr
	}
	if !runner.RunValidationOp() {
		runner.printExitMessages()
		return constants.ExitRuntimeErr
	}
	if !runner.RunPostValidationOps() {
		runner.printExitMessages()
		return constants.ExitRuntimeErr
	}
	if !runner.RunUploadOps() {
		runner.printExitMessages()
		return constants.ExitRuntimeErr
	}
	if deleteOnSuccess {
		runner.cleanup()
	} else {
		runner.setNoCleanupMessage()
	}

	if printOutput {
		runner.writeResult()
	}

	return constants.ExitOK
}

// printExitMessages prints info about a job's exit status to
// stdout and/or stderr. This does not save job results to the
// database because DART Runner is intended to run on servers
// where there may be no database. If the user wants to save
// job results, they can capture the output from stdin/stderr
// and store it as they please.
func (r *Runner) printExitMessages() {
	stdOutMsg, stdErrMsg := r.Job.GetResultMessages()
	if len(stdOutMsg) > 0 {
		fmt.Println(stdOutMsg)
	}
	if len(stdErrMsg) > 0 {
		fmt.Fprintln(os.Stderr, stdErrMsg)
	}
}

func (r *Runner) ValidateJob() bool {
	return r.Job.Validate()
}

func (r *Runner) RunPackageOp(skipArtifacts bool) bool {
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
		files, err := util.RecursiveFileList(filepath, false)
		if err != nil {
			errors := map[string]string{
				"SourceFiles": err.Error(),
			}
			op.Result.Info = "Packaging failed."
			op.Result.Finish(errors)
			return false
		}
		// TODO: Weed out duplicate files.
		sourceFiles = append(sourceFiles, files...)
	}
	bagger := NewBagger(op.OutputPath, r.Job.BagItProfile, sourceFiles)
	bagger.MessageChannel = r.MessageChannel // Careful! This may be nil.
	ok := bagger.Run()
	if !skipArtifacts {
		r.saveBaggingArtifacts(bagger)
	} else {
		r.Job.ArtifactsDir = "~artifacts not saved~"
	}
	r.Job.ByteCount = bagger.PayloadBytes()
	r.Job.PayloadFileCount = bagger.PayloadFileCount()
	r.Job.TotalFileCount = bagger.GetTotalFilesBagged()
	r.setResultFileInfo(op.Result, op.OutputPath, bagger.Errors)
	op.Result.Finish(bagger.Errors)
	if ok {
		op.Result.Info = "Bag created"
	}
	return ok
}

func (r *Runner) RunValidationOp() bool {
	if r.Job.ValidationOp == nil {
		return true
	}
	// Validate the package / bag
	op := r.Job.ValidationOp

	// Note that calling Start() resets info like
	// op.Result.FileSize. So call this first, then
	// set the relevant info.
	op.Result.Start()

	op.Result.FilePath = r.Job.ValidationOp.PathToBag
	fileInfo, err := os.Stat(r.Job.ValidationOp.PathToBag)
	if err == nil && fileInfo != nil {
		op.Result.FileSize = fileInfo.Size()
		op.Result.FileMTime = fileInfo.ModTime()
	}
	ok := r.Job.ValidationOp.Validate()
	if !ok {
		r.setResultFileInfo(op.Result, op.PathToBag, op.Errors)
		op.Result.Finish(op.Errors)
		return false
	}
	validator, err := NewValidator(r.Job.PackageOp.OutputPath, r.Job.BagItProfile)
	if err != nil {
		op.Result.Finish(validator.Errors)
		return false
	}
	// When running from the UI, we'll have a message channel to pass
	// info back to the front end. When running from command line, we won't.
	if r.MessageChannel != nil {
		validator.MessageChannel = r.MessageChannel
	}
	err = validator.ScanBag()
	if err != nil {
		errors := make(map[string]string)
		if len(validator.Errors) > 0 {
			errors = validator.Errors
		} else {
			errors["Validator.Scan"] = err.Error()
		}
		op.Result.Finish(errors)
		return false
	}
	ok = validator.Validate()
	op.Result.Finish(validator.Errors)
	if ok {
		op.Result.Info = "Bag is valid."
	}
	return ok
}

func (r *Runner) RunUploadOps() bool {
	if len(r.Job.UploadOps) == 0 {
		return true
	}
	// Run upload ops in sequence. If any fails, continue
	// with remaining uploads.
	allSucceeded := true
	for _, op := range r.Job.UploadOps {
		err := op.CalculatePayloadSize()
		if err != nil {
			op.Result.Finish(map[string]string{"Upload.CalculatePayloadSize": err.Error()})
			allSucceeded = false
			continue
		}
		op.Result.Start()
		ok := op.DoUpload(r.MessageChannel)
		if op.SourceFiles != nil && len(op.SourceFiles) > 0 {
			r.setResultFileInfo(op.Result, op.SourceFiles[0], op.Errors)
		}
		op.Result.Finish(op.Errors)
		if !ok {
			allSucceeded = false
		}
		if r.MessageChannel != nil {
			r.writeStageOutcome(constants.StageUpload, op.StorageService.Name, ok)
		}
	}
	return allSucceeded
}

func (r *Runner) RunPostValidationOps() bool {
	if r.Job.PostValidationOps == nil || len(r.Job.PostValidationOps) == 0 {
		return true
	}
	// Run upload ops in sequence. If any fails, continue
	// with remaining uploads.
	allSucceeded := true
	for _, op := range r.Job.PostValidationOps {
		op.Result.Start()
		err := op.Run(r.MessageChannel)
		if err != nil {
			key := fmt.Sprintf("PostValidateOperation.%s", op.Command)
			op.Result.Finish(map[string]string{key: err.Error()})
			allSucceeded = false
			r.writeStageOutcome(constants.StagePostValidation, op.Command, false)
			continue
		} else {
			op.Result.Finish(op.Errors)
			r.writeStageOutcome(constants.StagePostValidation, op.Command, true)
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

// writeResult writes the result of a job to STDOUT and/or STDERR
func (r *Runner) writeResult() {
	stdoutMessage, stderrMessage := r.Job.GetResultMessages()
	if len(stdoutMessage) > 0 {
		fmt.Println(stdoutMessage)
	}
	if len(stderrMessage) > 0 {
		fmt.Fprintln(os.Stderr, stderrMessage)
	}
}

func (r *Runner) saveBaggingArtifacts(bagger *Bagger) {
	if bagger == nil {
		Dart.Log.Warningf("SaveBaggingArtifacts got nil bagger")
		return
	}
	// GUI mode has a local SQLite database; command line
	// modes for dart-runner and apt-cmd do not, so we save
	// artifacts to file system.
	if Dart.RuntimeMode == constants.ModeDartGUI {
		r.Job.ArtifactsDir = "~database~"
		r.saveArtifactsToDatabase(bagger)
	} else {
		r.Job.ArtifactsDir = bagger.ArtifactsDir()
		r.saveArtifactsToFileSystem(bagger)
	}
}

func (r *Runner) saveArtifactsToDatabase(bagger *Bagger) {
	for manifestName, content := range bagger.ManifestArtifacts {
		artifact := NewManifestArtifact(r.Job.Name(), r.Job.ID, manifestName, content)
		err := ArtifactSave(artifact)
		if err != nil {
			Dart.Log.Warningf("Error saving manifest artifact %s for job %s: %s", manifestName, r.Job.Name(), err.Error())
		}
	}
	for tagFileName, content := range bagger.TagFileArtifacts {
		artifact := NewTagFileArtifact(r.Job.Name(), r.Job.ID, tagFileName, content)
		err := ArtifactSave(artifact)
		if err != nil {
			Dart.Log.Warningf("Error saving tag file artifact %s for job %s: %s", tagFileName, r.Job.Name(), err.Error())
		}
	}
}

func (r *Runner) saveArtifactsToFileSystem(bagger *Bagger) {
	artifactsDir := bagger.ArtifactsDir()
	err := os.Mkdir(artifactsDir, 0755)
	if err != nil && !util.FileExists(artifactsDir) {
		Dart.Log.Warningf("Cannot create artifacts directory '%s': %s", artifactsDir, err.Error())
	}
	// Even if mkdir above failed, the dir might already exist.
	// We'll only bail on this operation if the dir isn't there.
	if !util.FileExists(artifactsDir) {
		Dart.Log.Warningf("Will not write artifacts for bag %s because outputDir %s does not exist", r.Job.Name(), artifactsDir)
		return
	}
	for manifestName, content := range bagger.ManifestArtifacts {
		outputPath := filepath.Join(artifactsDir, manifestName)
		err := os.WriteFile(outputPath, []byte(content), 0644)
		if err != nil {
			Dart.Log.Warningf("Error saving manifest file artifact %s for job %s: %s", manifestName, r.Job.Name(), err.Error())
		}
	}
	for tagFileName, content := range bagger.TagFileArtifacts {
		outputPath := filepath.Join(artifactsDir, tagFileName)
		err := os.WriteFile(outputPath, []byte(content), 0644)
		if err != nil {
			Dart.Log.Warningf("Error saving tag file artifact %s for job %s: %s", tagFileName, r.Job.Name(), err.Error())
		}
	}
}
