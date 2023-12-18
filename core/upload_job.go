package core

import (
	"fmt"
	"time"

	"github.com/APTrust/dart-runner/constants"
	"github.com/google/uuid"
)

// UploadJob represents an upload-only job, in
// which we may be sending multiple files to multiple
// targets. This type of job has no packaging or validation
// step.
type UploadJob struct {
	ID                string
	PathsToUpload     []string
	StorageServiceIDs []string
	UploadOps         []UploadOperation
	Name              string
	Errors            map[string]string
}

func NewUploadJob() *UploadJob {
	return &UploadJob{
		ID:                uuid.NewString(),
		PathsToUpload:     make([]string, 0),
		StorageServiceIDs: make([]string, 0),
		UploadOps:         make([]UploadOperation, 0),
		Name:              fmt.Sprintf("Upload Job - %s", time.Now().Format(time.RFC3339Nano)),
		Errors:            make(map[string]string),
	}
}

// ObjID returns this job's object id (uuid).
func (job *UploadJob) ObjID() string {
	return job.ID
}

// ObjName returns this object's name, so names will be
// searchable and sortable in the DB.
func (job *UploadJob) ObjName() string {
	return job.Name
}

// ObjType returns this object's type name.
func (job *UploadJob) ObjType() string {
	return constants.TypeUploadJob
}

// String returns a string representation of this UploadJob,
// which is the same as Name().
func (job *UploadJob) String() string {
	return job.Name
}

// IsDeletable describes whether users can delete this
// object from the database. All UploadJobs are deletable.
func (job *UploadJob) IsDeletable() bool {
	return true
}

// ToForm returns a form object through which users can
// edit this UploadJob.
func (job *UploadJob) ToForm() *Form {
	form := NewForm(constants.TypeUploadJob, job.ID, job.Errors)
	form.UserCanDelete = true
	form.AddMultiValueField("PathsToUpload", "Items to Upload", job.PathsToUpload, true)
	form.AddMultiValueField("StorageServiceIDs", "Upload Targets", job.StorageServiceIDs, true)
	return form
}

// Validate returns true if this UploadJob is valid, false if not.
// Check the value of Errors or GetErrors() after calling this
// to see why validation failed.
func (job *UploadJob) Validate() bool {
	job.Errors = make(map[string]string)
	if len(job.PathsToUpload) == 0 {
		job.Errors["PathsToUpload"] = "You must select at least one item to upload."
	}
	// We can test that all files exist, but it's a pain to make
	// the user correct that. Instead, when the job runs, we'll
	// record an error if any local files are missing or unreadable.
	if len(job.StorageServiceIDs) == 0 {
		job.Errors["StorageServiceIDs"] = "Please choose at least one upload target."
	}
	return len(job.Errors) == 0
}

// GetErrors returns a map of errors describing why this
// UploadJob is not valid.
func (job *UploadJob) GetErrors() map[string]string {
	return job.Errors
}
