package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
	"github.com/google/uuid"
)

type Workflow struct {
	ID                string            `json:"id"`
	BagItProfile      *BagItProfile     `json:"bagItProfile"`
	Description       string            `json:"description"`
	Errors            map[string]string `json:"-"`
	Name              string            `json:"name"`
	PackageFormat     string            `json:"packageFormat"`
	StorageServiceIDs []string          `json:"storageServiceIds"`
	StorageServices   []*StorageService `json:"storageServices"`
}

func WorkflowFromJson(pathToFile string) (*Workflow, error) {
	workflow := &Workflow{}
	data, err := util.ReadFile(pathToFile)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, workflow)
	if err == nil {
		workflow.resolveStorageServices()
	}
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
		ID:              uuid.NewString(),
		Name:            "New Workflow",
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
	workflow.resolveStorageServices()
	return workflow, nil
}

func (w *Workflow) Validate() bool {
	w.Errors = make(map[string]string)
	if strings.TrimSpace(w.Name) == "" {
		w.Errors["Name"] = "Workflow requires a name."
	}
	if w.PackageFormat == "" {
		w.Errors["PackageFormat"] = "Workflow requires a pacakage format."
	}
	if w.PackageFormat == constants.PackageFormatBagIt && w.BagItProfile == nil {
		w.Errors["BagItProfile"] = "Workflow requires a BagIt profile."
	}
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
		ID:                w.ID,
		BagItProfile:      profile,
		Description:       w.Description,
		Errors:            w.Errors,
		Name:              w.Name,
		PackageFormat:     w.PackageFormat,
		StorageServiceIDs: w.StorageServiceIDs,
		StorageServices:   ssCopy,
	}
}

func (w *Workflow) ToForm() *Form {
	w.resolveStorageServices()
	form := NewForm(constants.TypeWorkflow, w.ID, w.Errors)
	form.UserCanDelete = true

	form.AddField("ID", "ID", w.ID, true)
	form.AddField("Name", "Name", w.Name, true)
	form.AddField("Description", "Description", w.Description, false)

	// NOTE: For now, we're working only with BagIt format. That may change
	// in future if we support OCFL.
	packageFormatField := form.AddField("PackageFormat", "Package Format", w.PackageFormat, true)
	packageFormatField.Choices = MakeChoiceList(constants.PackageFormats, w.PackageFormat)

	selectedProfileIds := make([]string, 0)
	if w.BagItProfile != nil {
		selectedProfileIds = []string{w.BagItProfile.ID}
	}

	bagItProfileField := form.AddField("BagItProfileID", "BagIt Profile", "", false)
	bagItProfileField.Choices = ObjChoiceList(constants.TypeBagItProfile, selectedProfileIds)

	storageServicesField := form.AddMultiValueField("StorageServiceIDs", "Storage Services", w.StorageServiceIDs, false)
	storageServicesField.Choices = ObjChoiceList(constants.TypeStorageService, w.StorageServiceIDs)
	return form
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

// resolveStorageServices handles a potential problem in the
// transition to the new DART. When DART and DART Runner were separate
// apps, DART attached only StorageService IDs to a workflow, while
// DART Runner attached entire StorageService objects to the workflow.
//
// From here on, SS = StorageService.
//
// Runner used SS objects because it had no local database and needed
// all of the relevant info to be embedded in the JSON blob to do its
// work.
//
// DART stored SS ids in the workflow object and loaded the full
// StorageService objects at runtime to do its work. The advantage
// here was that if credentials or other details of the SS object
// changed, DART would always be loading the most current data.
//
// Now that DART and DART Runner share the same codebase, we have
// to handle the case of legacy workflows from both products. That
// means we may be loading an old DART workflow with SS ids or an
// old Runner workflow with SS objects. In either case, we want to
// load the SS objects from the DB if possible. If not, we'll stick
// with the SS objects that are already attached. We also want to make
// sure that StorageServiceIDs is populated in either case.
func (w *Workflow) resolveStorageServices() {
	if w.StorageServiceIDs == nil {
		w.StorageServiceIDs = make([]string, 0)
	}
	for _, ss := range w.StorageServices {
		if !util.StringListContains(w.StorageServiceIDs, ss.ID) {
			w.StorageServiceIDs = append(w.StorageServiceIDs, ss.ID)
		}
	}
	for _, ssid := range w.StorageServiceIDs {
		result := ObjFind(ssid)
		// If DART Runner is running on a server, it may not
		// even have a database, so we do nothing here. If we
		// did get a result, replace a potentially stale copy
		// of the StorageService with the freshest one from the
		// database. If the StorageService doesn't exist in
		// w.StorageServices, then add it. This is important
		// when users are editing workflows for export.
		if result.Error == nil && result.StorageService() != nil {
			found := false
			ssFromDatabase := result.StorageService()
			for i, ss := range w.StorageServices {
				if ss.ID == ssFromDatabase.ID {
					found = true
					w.StorageServices[i] = ssFromDatabase
				}
			}
			if !found {
				w.StorageServices = append(w.StorageServices, ssFromDatabase)
			}
		}
	}
}

// ExportJson returns formatted JSON describing this workflow,
// including full StorageService records. Use this to export
// a workflow to be run on an external DART Runner server.
func (w *Workflow) ExportJson() ([]byte, error) {
	w.resolveStorageServices()
	return json.MarshalIndent(w, "", "  ")
}

// HasPlaintextPasswords returns true if any of the storage services
// in this workflow contain plaintext passwords. The front end will
// warn the user about plain text passwords when they export a workflow.
func (w *Workflow) HasPlaintextPasswords() bool {
	w.resolveStorageServices()
	for _, ss := range w.StorageServices {
		if ss.Password != "" && !strings.HasPrefix(ss.Password, "env:") {
			return true
		}
	}
	return false
}
