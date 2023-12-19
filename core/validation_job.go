package core

import (
	"fmt"
	"strings"

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
	ValidationOps   []*ValidationOperation
	Name            string
	Errors          map[string]string
}

func NewValidationJob() *ValidationJob {
	id := uuid.NewString()
	return &ValidationJob{
		ID:              id,
		PathsToValidate: make([]string, 0),
		ValidationOps:   make([]*ValidationOperation, 0),
		Errors:          make(map[string]string),
		Name:            fmt.Sprintf("Validation Job: %s", id),
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

	profileField := form.AddField("BagItProfileID", "BagIt Profile", job.BagItProfileID, true)
	profileField.Choices = ObjChoiceList(constants.TypeBagItProfile, []string{job.BagItProfileID})

	pathsField := form.AddMultiValueField("PathsToValidate", "Items to Validate", job.PathsToValidate, true)
	pathsField.Values = job.PathsToValidate

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

// Run runs this validation job, validating all PathsToValidate against
// the selected BagItProfile. This returns constants.ExitOK if all
// operations succeed. If any params are invalid or the required profile
// cannot be found, it returns constants.ExitUsage error because the
// user didn't supply valid info to complete the job. For any other
// runtime failure, it returns constants.ExitRuntimeError.
//
// Note that as long as the user supplies a valid profile and paths to
// validate, this will attempt to validate all bags. It's possible that
// some bags will be valid and some will not. Check the results of each
// ValidationOperation.Result if you get a non-zero exit code.
func (job *ValidationJob) Run(messageChannel chan *EventMessage) int {
	job.ValidationOps = make([]*ValidationOperation, 0)
	if !job.Validate() {
		// job.Errors is set inside call to Validate()
		return constants.ExitUsageErr
	}
	result := ObjFind(job.BagItProfileID)
	if result.Error != nil {
		job.Errors["BagItProfile"] = result.Error.Error()
		return constants.ExitRuntimeErr
	}
	profile := result.BagItProfile()
	if !profile.Validate() {
		job.Errors["BagItProfile"] = "BagIt profile is not valid"
		for key, value := range profile.Errors {
			job.Errors[key] = value
		}
		return constants.ExitUsageErr
	}
	exitCode := constants.ExitOK
	for _, pathToValidate := range job.PathsToValidate {
		if !job.runOne(pathToValidate, profile, messageChannel) {
			exitCode = constants.ExitRuntimeErr
		}
	}
	return exitCode
}

func (job *ValidationJob) runOne(pathToBag string, profile *BagItProfile, messageChannel chan *EventMessage) bool {
	op := NewValidationOperation(pathToBag)
	job.ValidationOps = append(job.ValidationOps, op)
	op.Result.Start()

	// Get a validator object to do the work. If this returns an
	// error, it's usually "file not found."
	validator, err := NewValidator(pathToBag, profile)
	if err != nil {
		errMap := map[string]string{
			pathToBag: err.Error(),
		}
		op.Result.Finish(errMap)
		return false
	}

	// When running from the UI, we'll have a message channel to pass
	// info back to the front end. When running from command line, we won't.
	if messageChannel != nil {
		validator.MessageChannel = messageChannel
	}

	// Scan the bag first, to build up an idea of what's in it.
	// This man return an error if the path is unreadable or if
	// we're trying to read a corrupt tar file.
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

	// Now that we know what's in the bag, validate it.
	// If the contents are invalid, validator.Errors will
	// contain specific info about what's wrong.
	ok := validator.Validate()
	op.Result.Finish(validator.Errors)
	if ok {
		op.Result.Info = "Bag is valid."
	}
	return ok
}
