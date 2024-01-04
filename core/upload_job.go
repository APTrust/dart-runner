package core

import (
	"fmt"

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
	UploadOps         []*UploadOperation
	Name              string
	Errors            map[string]string
}

func NewUploadJob() *UploadJob {
	id := uuid.NewString()
	return &UploadJob{
		ID:                id,
		PathsToUpload:     make([]string, 0),
		StorageServiceIDs: make([]string, 0),
		UploadOps:         make([]*UploadOperation, 0),
		Name:              fmt.Sprintf("Upload Job - %s", id),
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

	ssidField := form.AddMultiValueField("StorageServiceIDs", "Upload Targets", job.StorageServiceIDs, true)
	ssidField.Choices = ObjChoiceList(constants.TypeStorageService, job.StorageServiceIDs)

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

func (job *UploadJob) Run(messageChannel chan *EventMessage) int {
	job.UploadOps = make([]*UploadOperation, 0)

	if !job.Validate() {
		return constants.ExitUsageErr
	}

	//
	// TODO: If path is directory, expand by listing all files recursively.
	//

	uploadTargets := make(map[string]*StorageService)
	for i, ssid := range job.StorageServiceIDs {
		result := ObjFind(ssid)
		if result.Error != nil {
			errName := fmt.Sprintf("StorageService #%d", i+1)
			job.Errors[errName] = result.Error.Error()
		} else {
			uploadTargets[ssid] = result.StorageService()
		}
	}
	if len(uploadTargets) == 0 {
		// User didn't provide any valid Storage Services
		job.Errors["StorageService"] = "No valid storage services found."
		return constants.ExitUsageErr
	}

	// Try all uploads, and keep going if any fail.
	exitCode := constants.ExitOK
	for _, storageService := range uploadTargets {
		if !job.runOne(storageService, messageChannel) {
			exitCode = constants.ExitRuntimeErr
		}
	}
	return exitCode
}

func (job *UploadJob) runOne(storageService *StorageService, messageChannel chan *EventMessage) bool {
	uploadOp := NewUploadOperation(storageService, job.PathsToUpload)
	job.UploadOps = append(job.UploadOps, uploadOp)
	uploadOp.Result.Start()
	if !uploadOp.Validate() {
		uploadOp.Result.Finish(uploadOp.Errors)
		return false
	}

	err := uploadOp.CalculatePayloadSize()
	if err != nil {
		uploadOp.Result.Finish(map[string]string{"Upload.CalculatePayloadSize": err.Error()})
		return false
	}
	ok := uploadOp.DoUpload(messageChannel)
	uploadOp.Result.Finish(uploadOp.Errors)
	if messageChannel != nil {
		status := constants.StatusFailed
		if ok {
			status = constants.StatusSuccess
		}
		eventMessage := &EventMessage{
			EventType: constants.EventTypeFinish,
			Stage:     constants.StageUpload,
			Status:    status,
			Message:   storageService.Name,
		}
		messageChannel <- eventMessage
	}
	return ok
}
