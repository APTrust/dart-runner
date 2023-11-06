package core

import (
	"fmt"
	"time"

	"github.com/APTrust/dart-runner/constants"
	"github.com/google/uuid"
)

// ExportSettings are settings that users can export as JSON to share
// with other users.
type ExportSettings struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	AppSettings        []*AppSetting       `json:"appSettings"`
	BagItProfiles      []*BagItProfile     `json:"bagItProfiles"`
	Questions          []*ExportQuestion   `json:"questions"`
	RemoteRepositories []*RemoteRepository `json:"remoteRepositories"`
	StorageServices    []*StorageService   `json:"storageServices"`
	Errors             map[string]string   `json:"-"`
}

// NewExportSettings returns a new ExportSettings object.
func NewExportSettings() *ExportSettings {
	return &ExportSettings{
		ID:                 uuid.NewString(),
		Name:               fmt.Sprintf("Export Settings - %s", time.Now().Format(time.RFC822)),
		AppSettings:        make([]*AppSetting, 0),
		BagItProfiles:      make([]*BagItProfile, 0),
		Questions:          make([]*ExportQuestion, 0),
		RemoteRepositories: make([]*RemoteRepository, 0),
		StorageServices:    make([]*StorageService, 0),
		Errors:             make(map[string]string),
	}
}

// ObjectIds returns the ids (uuids) for all the objects of the specified type
// in this group of export settings.
func (settings *ExportSettings) ObjectIds(objType string) ([]string, error) {
	var ids []string
	switch objType {
	case constants.TypeAppSetting:
		ids = make([]string, len(settings.AppSettings))
		for i, item := range settings.AppSettings {
			ids[i] = item.ID
		}
	case constants.TypeBagItProfile:
		ids = make([]string, len(settings.BagItProfiles))
		for i, item := range settings.BagItProfiles {
			ids[i] = item.ID
		}
	case constants.TypeRemoteRepository:
		ids = make([]string, len(settings.RemoteRepositories))
		for i, item := range settings.RemoteRepositories {
			ids[i] = item.ID
		}
	case constants.TypeStorageService:
		ids = make([]string, len(settings.StorageServices))
		for i, item := range settings.StorageServices {
			ids[i] = item.ID
		}
	default:
		return nil, constants.ErrUnknownType
	}
	return ids, nil
}

// ContainsPlaintextPassword returns true if any StorageService
// in these settings contains a plaintext password.
func (settings *ExportSettings) ContainsPlaintextPassword() bool {
	for _, ss := range settings.StorageServices {
		if ss.HasPlaintextPassword() {
			return true
		}
	}
	return false
}

// ContainsPlaintextAPIToken returns true if any RemoteRepository
// in these settings contains a plaintext API token.
func (settings *ExportSettings) ContainsPlaintextAPIToken() bool {
	for _, repo := range settings.RemoteRepositories {
		if repo.HasPlaintextAPIToken() {
			return true
		}
	}
	return false
}

// GetErrors returns a map of validation errors for this object.
func (settings *ExportSettings) GetErrors() map[string]string {
	return settings.Errors
}

// IsDeletable returns a boolean indicating whether users are
// allowed to delete this object from the database.
func (settings *ExportSettings) IsDeletable() bool {
	return true
}

// ObjID returns this object's id (uuid).
func (settings *ExportSettings) ObjID() string {
	return settings.ID
}

// ObjName returns this object's name.
func (settings *ExportSettings) ObjName() string {
	return settings.Name
}

// ObjType returns this object's type.
func (settings *ExportSettings) ObjType() string {
	return constants.TypeExportSettings
}

// String returns a string describing this object's type and name.
func (settings *ExportSettings) String() string {
	return fmt.Sprintf("Export Settings: %s", settings.Name)
}

// ToForm returns a Form object that can represent this ExportSettings object.
// Note that the form does not include Questions. Because Questions are more
// complex than simple ID-Name object, the templates must handle them differently.
func (settings *ExportSettings) ToForm() *Form {
	form := NewForm(constants.TypeExportSettings, settings.ID, settings.Errors)
	form.UserCanDelete = settings.IsDeletable()

	form.AddField("ID", "ID", settings.ID, true)
	nameField := form.AddField("Name", "Name", settings.Name, true)
	nameField.Help = "Names allow you to differentiate your export settings. For example, 'External Donor Settings' and 'Internal Team Settings.'"

	appSettingIds, _ := settings.ObjectIds(constants.TypeAppSetting)
	appSettingsField := form.AddMultiValueField("AppSettings", "Application Settings", appSettingIds, false)
	appSettingsField.Choices = ObjChoiceList(constants.TypeAppSetting, appSettingIds)

	bagItProfileIds, _ := settings.ObjectIds(constants.TypeBagItProfile)
	bagItProfilesField := form.AddMultiValueField("BagItProfiles", "BagIt Profiles", bagItProfileIds, false)
	bagItProfilesField.Choices = ObjChoiceList(constants.TypeBagItProfile, bagItProfileIds)

	remoteRepoIds, _ := settings.ObjectIds(constants.TypeRemoteRepository)
	remoteReposField := form.AddMultiValueField("RemoteRepositories", "Remote Repositories", remoteRepoIds, false)
	remoteReposField.Choices = ObjChoiceList(constants.TypeRemoteRepository, remoteRepoIds)

	storageServiceIds, _ := settings.ObjectIds(constants.TypeStorageService)
	storageServicesField := form.AddMultiValueField("StorageServices", "Storage Services", storageServiceIds, false)
	storageServicesField.Choices = ObjChoiceList(constants.TypeStorageService, storageServiceIds)

	return form
}

// Validate always returns true for ExportSettings because we want
// to allow users to export any settings they choose.
func (settings *ExportSettings) Validate() bool {
	return true
}

func (settings *ExportSettings) SetValueFromResponse(question *ExportQuestion, response string) error {
	// TODO: Implement this.
	return nil
}
