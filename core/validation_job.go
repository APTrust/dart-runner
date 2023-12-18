package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/APTrust/dart-runner/constants"
	"github.com/google/uuid"
)

// ValidationJob is a job that only validates bags.
// This type of job may validate multiple bags, but
// it includes no package or upload operations.
type ValidationJob struct {
	ID              string
	BagItProfileID  string
	PathsToValidate []string
	ValidationOps   []ValidationOperation
	Name            string
	Errors          map[string]string
}

func NewValidationJob() *ValidationJob {
	return &ValidationJob{
		ID:              uuid.NewString(),
		PathsToValidate: make([]string, 0),
		ValidationOps:   make([]ValidationOperation, 0),
		Errors:          make(map[string]string),
		Name:            fmt.Sprintf("Validation Operation: %s", time.Now().Format(time.RFC3339Nano)),
	}
}

// ObjID returns this job's object id (uuid).
func (job *ValidationJob) ObjID() string {
	return job.ID
}

// ObjName returns this object's name, so names will be
// searchable and sortable in the DB.
func (job *ValidationJob) ObjName() string {
	return job.Name
}

// ObjType returns this object's type name.
func (job *ValidationJob) ObjType() string {
	return constants.TypeValidationJob
}

// String returns a string representation of this ValidationJob,
// which is the same as Name().
func (job *ValidationJob) String() string {
	return job.Name
}

// IsDeletable describes whether users can delete this
// object from the database. All ValidationJobs are deletable.
func (job *ValidationJob) IsDeletable() bool {
	return true
}

// ToForm returns a form object through which users can
// edit this ValidationJob.
func (job *ValidationJob) ToForm() *Form {
	form := NewForm(constants.TypeValidationJob, job.ID, job.Errors)
	form.UserCanDelete = true
	form.AddField("BagItProfileID", "BagIt Profile", job.BagItProfileID, true)
	form.AddMultiValueField("PathsToValidate", "Items to Validate", job.PathsToValidate, true)
	return form
}

// Validate returns true if this ValidationJob is valid, false if not.
// Check the value of Errors or GetErrors() after calling this
// to see why validation failed.
func (job *ValidationJob) Validate() bool {
	job.Errors = make(map[string]string)
	if len(job.PathsToValidate) == 0 {
		job.Errors["PathsToValidate"] = "You must select at least one item to validate."
	}
	if strings.TrimSpace(job.BagItProfileID) == "" {
		job.Errors["BagItProfileID"] = "Please choose a BagIt profile."
	}
	return len(job.Errors) == 0
}

// GetErrors returns a map of errors describing why this
// ValidationJob is not valid.
func (job *ValidationJob) GetErrors() map[string]string {
	return job.Errors
}
