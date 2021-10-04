package core

import (
	"fmt"
	"path"
	"time"

	"github.com/APTrust/dart-runner/bagit"
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
	BagItProfile *bagit.Profile       `json:"bagItProfile"`
	ByteCount    int64                `json:"byteCount"`
	DirCount     int                  `json:"dirCount"`
	Errors       map[string]string    `json:"errors"`
	FileCount    int                  `json:"fileCount"`
	PackageOp    *PackageOperation    `json:"packageOp"`
	UploadOps    []*UploadOperation   `json:"uploadOps"`
	ValidationOp *ValidationOperation `json:"validationOp"`
	WorkflowID   string               `json:"workflowId"`
}

func NewJob() *Job {
	return &Job{}
}

// Title returns a title for display purposes. It will use the first
// available non-empty value of: 1) the name of the file that the job
// packaged, 2) the name of the file that the job uploaded, or 3) a
// title or description of the bag from within the bag's tag files.
// If none of those is available, this will return "Job of <timestamp>",
// where timestamp is date and time the job was created.
func (job *Job) Title() string {
	var name = fmt.Sprintf("Job of %s", time.Now().Format(time.RFC3339))
	if name == "" && job.PackageOp != nil && job.PackageOp.PackageName != "" {
		name = path.Base(job.PackageOp.PackageName)
	}
	if name == "" && len(job.UploadOps) > 0 && len(job.UploadOps[0].SourceFiles) > 0 {
		name = path.Base(job.UploadOps[0].SourceFiles[0])
	}
	// Try to get a title from the bag.
	if name == "" && job.BagItProfile != nil {
		for _, tagName := range TitleTags {
			tag, _ := job.BagItProfile.FirstMatchingTag("tagName", tagName)
			if tag != nil && tag.UserValue != "" {
				name = tag.UserValue
				break
			}
		}
	}
	return name
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
			for _, result := range uploadOp.Results {
				if result != nil && !result.Completed.IsZero() {
					uploadedAt = result.Completed
				}
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
			for _, result := range op.Results {
				if result.Attempt > 0 {
					return true
				}
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
			for _, result := range op.Results {
				if result.Attempt > 0 {
					anyAttempted = true
				}
				if !result.Succeeded() {
					allSucceeded = false
				}
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
		if job.PackageOp.PackageFormat == "BagIt" && job.BagItProfile == nil {
			job.Errors["Job.bagItProfile"] = "BagIt packaging requires a BagItProfile."
		}
	}
	if job.ValidationOp != nil {
		job.ValidationOp.Validate()
		for key, value := range job.ValidationOp.Errors {
			job.Errors[key] = value
		}
		if job.BagItProfile == nil && job.ValidationOp.Result.Errors["Job.BagItProfile"] == "" {
			job.Errors["Job.BagItProfile"] = "Validation requires a BagItProfile."
		}
	}
	opNum := 1
	for _, uploadOp := range job.UploadOps {
		uploadOp.Validate()
		for key, value := range uploadOp.Errors {
			uniqueKey := fmt.Sprintf("%s-%d", key, opNum)
			job.Errors[uniqueKey] = value
		}
		opNum++
	}
	return len(job.Errors) == 0
}

// GetRunErrors returns a list of errors from all of this job's operations.
func (job *Job) GetRunErrors() map[string]string {
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
