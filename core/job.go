package core

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
	"github.com/google/uuid"
)

// TitleTags is a list of BagItProfile tags to check to try to find a
// meaningful title for this job. DART checks them in order and
// returns the first one that has a non-empty user-defined value.
var TitleTags = []string{
	"Title",
	"Internal-Sender-Identifier",
	"External-Identifier",
	"Internal-Sender-Description",
	"External-Description",
	"Description",
}

type Job struct {
	ID               string               `json:"id"`
	BagItProfile     *BagItProfile        `json:"bagItProfile"`
	ByteCount        int64                `json:"byteCount"`
	DirCount         int64                `json:"dirCount"`
	Errors           map[string]string    `json:"errors"`
	PayloadFileCount int64                `json:"fileCount"` // still fileCount in JSON for legacy compatibility
	TotalFileCount   int64                `json:"totalFileCount"`
	PackageOp        *PackageOperation    `json:"packageOp"`
	UploadOps        []*UploadOperation   `json:"uploadOps"`
	ValidationOp     *ValidationOperation `json:"validationOp"`
	WorkflowID       string               `json:"workflowId"`
}

// NewJob creates a new Job with a unique ID.
func NewJob() *Job {
	return &Job{
		ID:           uuid.NewString(),
		Errors:       make(map[string]string),
		PackageOp:    NewPackageOperation("", "", make([]string, 0)),
		UploadOps:    make([]*UploadOperation, 0),
		ValidationOp: NewValidationOperation(""),
	}
}

// JobFromJson converts the JSON data in file pathToFile
// into a Job object.
func JobFromJson(pathToFile string) (*Job, error) {
	job := &Job{}
	data, err := util.ReadFile(pathToFile)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, job)
	return job, err
}

// JobFromWorkflow creates a new job based on the specified
// workflow. Note that the new job will have no name or
// source files.
func JobFromWorkflow(workflow *Workflow) *Job {
	workflow.resolveStorageServices()
	job := NewJob()
	job.WorkflowID = workflow.ID
	job.BagItProfile = workflow.BagItProfile
	for _, ss := range workflow.StorageServices {
		files := make([]string, 0)
		uploadOp := NewUploadOperation(ss, files)
		job.UploadOps = append(job.UploadOps, uploadOp)
	}
	return job
}

// Name returns a name for this job, which is usually the file name of
// the package being built, validated, or uploaded.
func (job *Job) Name() string {
	if job.PackageOp != nil && job.PackageOp.PackageName != "" {
		return path.Base(job.PackageOp.PackageName)
	}
	if job.ValidationOp != nil && job.ValidationOp.PathToBag != "" {
		return path.Base(job.ValidationOp.PathToBag)
	}
	if job.UploadOps != nil && len(job.UploadOps) > 0 && len(job.UploadOps[0].SourceFiles) > 0 {
		return path.Base(job.UploadOps[0].SourceFiles[0])
	}
	// Try to get a title from the bag.
	if job.BagItProfile != nil {
		for _, tagName := range TitleTags {
			tag, _ := job.BagItProfile.FirstMatchingTag("tagName", tagName)
			if tag != nil && tag.UserValue != "" {
				return tag.UserValue
			}
		}
	}
	return fmt.Sprintf("Job of %s", time.Now().Format(time.RFC3339))
}

// PackagedAt returns the datetime on which this job's package
// operation completed.
func (job *Job) PackagedAt() time.Time {
	if job.PackageOp != nil && job.PackageOp.Result != nil {
		return job.PackageOp.Result.Completed
	}
	return time.Time{}
}

// PackageAttempted returns true if DART attempted to execute
// this job's package operation.
func (job *Job) PackageAttempted() bool {
	if job.PackageOp != nil && job.PackageOp.Result != nil {
		return job.PackageOp.Result.Attempt > 0
	}
	return false
}

// PackageSucceeded returns true if DART successfully completed
// this job's package operation. Note that this will return false
// if packaging failed and if packaging was never attempted, so check
// PackageAttempted as well.
func (job *Job) PackageSucceeded() bool {
	if job.PackageOp != nil && job.PackageOp.Result != nil {
		return job.PackageOp.Result.Succeeded()
	}
	return false
}

// ValidatedAt returns the datetime on which this job's validation
// operation completed.
func (job *Job) ValidatedAt() time.Time {
	if job.ValidationOp != nil && job.ValidationOp.Result != nil {
		return job.ValidationOp.Result.Completed
	}
	return time.Time{}
}

// ValidationAttempted returns true if DART attempted to execute
// this job's validation operation.
func (job *Job) ValidationAttempted() bool {
	if job.ValidationOp != nil && job.ValidationOp.Result != nil {
		return job.ValidationOp.Result.Attempt > 0
	}
	return false
}

// ValidationSucceeded returns true if DART successfully completed
// this job's validation operation. See ValidationAttempted as well.
func (job *Job) ValidationSucceeded() bool {
	if job.ValidationOp != nil && job.ValidationOp.Result != nil {
		return job.ValidationOp.Result.Succeeded()
	}
	return false
}

// UploadedAt returns the datetime on which this job's last upload
// operation completed.
func (job *Job) UploadedAt() time.Time {
	uploadedAt := time.Time{}
	if len(job.UploadOps) > 0 {
		for _, uploadOp := range job.UploadOps {
			if uploadOp.Result != nil && !uploadOp.Result.Completed.IsZero() {
				uploadedAt = uploadOp.Result.Completed
			}
		}
	}
	return uploadedAt
}

// UploadAttempted returns true if DART attempted to execute any of
// this job's upload operations.
func (job *Job) UploadAttempted() bool {
	if job.UploadOps != nil {
		for _, op := range job.UploadOps {
			if op.Result.Attempt > 0 {
				return true
			}
		}
	}
	return false
}

// UploadSucceeded returns true if DART successfully completed all of
// this job's upload operations. See UploadAttempted as well.
func (job Job) UploadSucceeded() bool {
	anyAttempted := false
	allSucceeded := true
	if job.UploadOps != nil {
		for _, op := range job.UploadOps {
			if op.Result.Attempt > 0 {
				anyAttempted = true
			}
			if !op.Result.Succeeded() {
				allSucceeded = false
			}
		}
	}
	return anyAttempted && allSucceeded
}

// Validate returns true or false, indicating whether this object
// contains complete and valid data. If it returns false, check
// the errors property for specific errors.
func (job *Job) Validate() bool {
	job.Errors = make(map[string]string)
	if job.PackageOp != nil {
		job.PackageOp.Validate()
		job.Errors = job.PackageOp.Errors
		if job.PackageOp.PackageFormat == constants.PackageFormatBagIt && job.BagItProfile == nil {
			job.Errors["Job.Package.BagItProfile"] = "BagIt packaging requires a BagItProfile."
		}
	}
	// ValidationOp.PathToBag should be defined, but it won't exist
	// until PackageOp finishes.
	if job.ValidationOp != nil {
		if strings.TrimSpace(job.ValidationOp.PathToBag) == "" {
			job.Errors["Job.Validate.PathToBag"] = "Validation requires a file or bag to validate."
		}
		if job.BagItProfile == nil {
			job.Errors["Job.Validate.BagItProfile"] = "Validation requires a BagItProfile."
		}
	}

	// UploadOp validation ensures that files exist. They don't yet, so we
	// don't want to run full validation. Just ensure we have valid storage
	// service records.
	if job.UploadOps != nil {
		for i, uploadOp := range job.UploadOps {
			errKey := fmt.Sprintf("UploadOp[%d].StorageService", i)
			if uploadOp.StorageService == nil {
				job.Errors[errKey] = "UploadOperation requires a StorageService"
			} else if !uploadOp.StorageService.Validate() {
				for key, errMsg := range uploadOp.StorageService.Errors {
					ssErrKey := "Job.StorageService." + key
					job.Errors[ssErrKey] = errMsg
				}
			}
		}
	}
	if job.PackageOp == nil && job.ValidationOp == nil && (job.UploadOps == nil || len(job.UploadOps) == 0) {
		job.Errors["Job"] = "Job has nothing to package, validate, or upload."
	}
	return len(job.Errors) == 0
}

// RuntimeErrors returns a list of errors from all of this job's operations.
func (job *Job) RuntimeErrors() map[string]string {
	errs := make(map[string]string)
	if job.PackageOp != nil && job.PackageOp.Result != nil {
		for key, value := range job.PackageOp.Errors {
			job.Errors[key] = value
		}
	}
	if job.ValidationOp != nil && job.ValidationOp.Result != nil {
		for key, value := range job.ValidationOp.Errors {
			job.Errors[key] = value
		}
	}
	if job.UploadOps != nil && len(job.UploadOps) > 0 {
		opNum := 1
		for _, uploadOp := range job.UploadOps {
			for key, value := range uploadOp.Errors {
				uniqueKey := fmt.Sprintf("%s-%d", key, opNum)
				job.Errors[uniqueKey] = value
			}
		}
		opNum++
	}
	return errs
}

func (job *Job) GetResultMessages() (stdoutMessage, stdErrMessage string) {
	result := NewJobResult(job)
	stdoutMessage, err := result.ToJson()

	// If we can't serialize the JobResult, tell the user.
	if err != nil {
		stdErrMessage = fmt.Sprintf("Error getting result for job %s: %s", job.Name(), err.Error())
		status := "succeeded"
		if !result.Succeeded {
			status = "failed"
		}
		stdoutMessage = fmt.Sprintf("Job %s %s, but dart runner encountered an error when trying to report detailed results.", job.Name(), status)
		return stdoutMessage, stdErrMessage
	}

	// OK, we can serialize the the JobResult. If there were any errors,
	// make a note in STDERR.
	if !result.Succeeded {
		stdErrMessage = fmt.Sprintf("Job %s encountered one or more errors. See the JSON results in stdout.", job.Name())

	}
	return stdoutMessage, stdErrMessage
}

// HasPackageOp returns true if this job includes a package operation.
func (job *Job) HasPackageOp() bool {
	return job.PackageOp != nil
}

// HasUploadOps returns true if this job includes one or more upload operations.
func (job *Job) HasUploadOps() bool {
	return len(job.UploadOps) > 0 && job.UploadOps[0] != nil
}

// PersistentObject interface

// ObjID returns this job's object id (uuid).
func (job *Job) ObjID() string {
	return job.ID
}

// ObjName returns this object's name, so names will be
// searchable and sortable in the DB.
func (job *Job) ObjName() string {
	return job.Name()
}

// ObjType returns this object's type name.
func (job *Job) ObjType() string {
	return constants.TypeJob
}

func (job *Job) IsDeletable() bool {
	return true
}

func (job *Job) String() string {
	return fmt.Sprintf("Job: %s [%s]", job.Name(), job.ID)
}

func (job *Job) GetErrors() map[string]string {
	return job.Errors
}

func (job *Job) ToForm() *Form {
	form := NewForm(constants.TypeJob, job.ID, job.Errors)
	form.AddField("ID", "ID", job.ID, true)

	bagitProfileID := ""
	serializationOptions := constants.AcceptSerialization
	if job.BagItProfile != nil {
		bagitProfileID = job.BagItProfile.ID
		// TODO: Include serializations allowed in profile, if DART supports them.
	}
	bagItProfile := form.AddField("BagItProfileID", "BagIt Profile ID", bagitProfileID, false)
	bagItProfile.Choices = ObjChoiceList(constants.TypeBagItProfile, []string{bagitProfileID})
	bagItProfile.Help = "Choose the BagIt profile to which your bag should conform."

	// PackageOp
	packageFormat := form.AddField("PackageFormat", "Package Format", job.PackageOp.PackageFormat, true)
	packageFormat.Choices = MakeChoiceList(constants.PackageFormats, job.PackageOp.PackageFormat)

	serialization := form.AddField("BagItSerialization", "BagItSerialization", job.PackageOp.BagItSerialization, true)
	serialization.Choices = MakeChoiceList(serializationOptions, job.PackageOp.BagItSerialization)
	serialization.Help = "How should this bag be serialized or compressed?"

	packageName := form.AddField("PackageName", "Package Name", job.PackageOp.PackageName, true)
	packageName.Help = "The name of the folder or file to create. E.g. 'photos', 'photos.tar', etc."

	outputPath := form.AddField("OutputPath", "Output Path", job.PackageOp.OutputPath, true)
	outputPath.Help = "This is where DART will create the bag. This field updates automatically when you update the package name."

	// UploadOps

	// ValidationOp

	return form
}
