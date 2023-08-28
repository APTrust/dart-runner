package core

import (
	"encoding/json"
	"fmt"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

type Workflow struct {
	ID              string            `json:"id"`
	BagItProfile    *BagItProfile     `json:"bagItProfile"`
	Description     string            `json:"description"`
	Errors          map[string]string `json:"-"`
	Name            string            `json:"name"`
	PackageFormat   string            `json:"packageFormat"`
	StorageServices []*StorageService `json:"storageServices"`
}

func WorkflowFromJson(pathToFile string) (*Workflow, error) {
	workflow := &Workflow{}
	data, err := util.ReadFile(pathToFile)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, workflow)
	return workflow, err
}

// WorkflowFromJob creates a new workflow based on param job. It does
// not save the workflow. That's up to the caller.
//
// When a user defines a job that does what they want, they will often
// convert it to a workflow, so they can run the same packaging and
// upload operations on a large of set of materials.
func WorkFlowFromJob(job *Job) (*Workflow, error) {
	workflow := &Workflow{
		Name:            "",
		Description:     "",
		PackageFormat:   constants.PackageFormatNone,
		StorageServices: make([]*StorageService, len(job.UploadOps)),
	}
	if job.PackageOp != nil {
		workflow.PackageFormat = job.PackageOp.PackageFormat
	}
	// Load a fresh copy of the BagIt profile, because the copy in the
	// job may have custom tag values assigned.
	if job.BagItProfile != nil {
		result := ObjFind(job.BagItProfile.ID)
		if result.Error != nil {
			return nil, fmt.Errorf("Can't find BagIt Profile for this workflow: %s", result.Error.Error())
		}
		workflow.BagItProfile = result.BagItProfile()
	}
	for i, uploadOp := range job.UploadOps {
		workflow.StorageServices[i] = uploadOp.StorageService
	}
	return workflow, nil
}

func (w *Workflow) Validate() bool {
	w.Errors = make(map[string]string)
	if w.BagItProfile != nil && !w.BagItProfile.Validate() {
		for key, value := range w.BagItProfile.Errors {
			w.Errors["BagItProfile."+key] = value
		}
	}
	if w.StorageServices != nil {
		for _, ss := range w.StorageServices {
			if !ss.Validate() {
				for key, value := range ss.Errors {
					uniqueKey := fmt.Sprintf("%s.StorageService.%s", ss.Name, key)
					w.Errors[uniqueKey] = value
				}
			}
		}
	}
	return len(w.Errors) == 0
}

func (w *Workflow) Copy() *Workflow {
	ssCopy := make([]*StorageService, len(w.StorageServices))
	for i, ss := range w.StorageServices {
		ssCopy[i] = ss.Copy()
	}
	profile := BagItProfileClone(w.BagItProfile)
	return &Workflow{
		ID:              w.ID,
		BagItProfile:    profile,
		Description:     w.Description,
		Errors:          w.Errors,
		Name:            w.Name,
		PackageFormat:   w.PackageFormat,
		StorageServices: ssCopy,
	}

}

func (w *Workflow) ToForm() *Form {

	// TODO: Implement this!

	// Name, Description, PackageFormat, BagItProfileID, StorageServices

	return nil
}

// ObjID returns this w's object id (uuid).
func (w *Workflow) ObjID() string {
	return w.ID
}

// ObjName returns this object's name, so names will be
// searchable and sortable in the DB.
func (w *Workflow) ObjName() string {
	return w.Name
}

// ObjType returns this object's type name.
func (w *Workflow) ObjType() string {
	return constants.TypeWorkflow
}

func (w *Workflow) String() string {
	return fmt.Sprintf("Workflow '%s'", w.Name)
}

func (w *Workflow) GetErrors() map[string]string {
	return w.Errors
}

func (w *Workflow) IsDeletable() bool {
	return true
}
