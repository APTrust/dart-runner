package core

import (
	"sort"

	"github.com/APTrust/dart-runner/constants"
)

var AppSettingSettableFields = []string{"Value"}
var RemoteRepositorySettableFields = []string{"APIToken", "LoginExtra", "Name", "Url", "UserID"}
var StorageServiceSettableFields = []string{"AllowsDownload", "AllowsUpload", "Bucket", "Host", "Login", "LoginExtra", "Name", "Password", "Port", "Protocol"}

// ExportOptions returns options to populate lists on
// the export settings page.
type ExportOptions struct {
	AppSettings            []NameIDPair            `json:"appSettings"`
	AppSettingFields       []string                `json:"appSettingFields"`
	BagItProfiles          []NameIDPair            `json:"bagItProfiles"`
	BagItProfileFields     map[string][]NameIDPair `json:"bagItProfileFields"`
	RemoteRepositories     []NameIDPair            `json:"remoteRepositories"`
	RemoteRepositoryFields []string                `json:"remoteRepositoryFields"`
	StorageServices        []NameIDPair            `json:"storageServices"`
	StorageServiceFields   []string                `json:"storageServiceFields"`
}

// NewExportOptions returns an object containing the data
// we need to populate lists on the export settings page.
func NewExportOptions() *ExportOptions {
	profiles := ObjNameIdList(constants.TypeBagItProfile)
	opts := &ExportOptions{
		AppSettings:            ObjNameIdList(constants.TypeAppSetting),
		AppSettingFields:       AppSettingSettableFields,
		BagItProfiles:          profiles,
		BagItProfileFields:     make(map[string][]NameIDPair),
		RemoteRepositories:     ObjNameIdList(constants.TypeRemoteRepository),
		RemoteRepositoryFields: RemoteRepositorySettableFields,
		StorageServices:        ObjNameIdList(constants.TypeStorageService),
		StorageServiceFields:   StorageServiceSettableFields,
	}
	for _, profile := range profiles {
		tagsForProfile, err := UserSettableTagsForProfile(profile.ID)
		if err == nil {
			opts.BagItProfileFields[profile.ID] = tagsForProfile
		} else {
			opts.BagItProfileFields[profile.ID] = []NameIDPair{{Name: err.Error(), ID: profile.ID}}
		}
	}
	return opts
}

// NewExportOptionsFromSettings returns an object containing
// data for use on the export questions page. Questions only pertain
// to settings in the export settings, so we want to limit the
// data here to those settings only. This differs from NewExportOptions(),
// which returns info about all settings in the entire system.
func NewExportOptionsFromSettings(exportSettings *ExportSettings) *ExportOptions {
	opts := &ExportOptions{
		AppSettings:            make([]NameIDPair, len(exportSettings.AppSettings)),
		AppSettingFields:       AppSettingSettableFields,
		BagItProfiles:          make([]NameIDPair, len(exportSettings.BagItProfiles)),
		BagItProfileFields:     make(map[string][]NameIDPair),
		RemoteRepositories:     make([]NameIDPair, len(exportSettings.RemoteRepositories)),
		RemoteRepositoryFields: RemoteRepositorySettableFields,
		StorageServices:        make([]NameIDPair, len(exportSettings.StorageServices)),
		StorageServiceFields:   StorageServiceSettableFields,
	}
	for i, setting := range exportSettings.AppSettings {
		opts.AppSettings[i] = NameIDPair{Name: setting.Name, ID: setting.ID}
	}
	for i, profile := range exportSettings.BagItProfiles {
		opts.BagItProfiles[i] = NameIDPair{Name: profile.Name, ID: profile.ID}
		tagsForProfile, err := UserSettableTagsForProfile(profile.ID)
		if err == nil {
			opts.BagItProfileFields[profile.ID] = tagsForProfile
		} else {
			opts.BagItProfileFields[profile.ID] = []NameIDPair{{Name: err.Error(), ID: profile.ID}}
		}
	}
	for i, repo := range exportSettings.RemoteRepositories {
		opts.RemoteRepositories[i] = NameIDPair{Name: repo.Name, ID: repo.ID}
	}
	for i, ss := range exportSettings.StorageServices {
		opts.StorageServices[i] = NameIDPair{Name: ss.Name, ID: ss.ID}
	}
	return opts
}

// UserSettableTagsForProfile returns a list of tags defined in the BagIt profile
// with the specified ID. The list inlcudes only those tags whose values users
// can set. Tags such as Payload-Oxum will be excluded because users can't
// set their value.
func UserSettableTagsForProfile(bagItProfileID string) ([]NameIDPair, error) {
	result := ObjFind(bagItProfileID)
	if result.Error != nil {
		return nil, result.Error
	}
	profile := result.BagItProfile()
	pairs := make([]NameIDPair, 0)
	for _, tag := range profile.Tags {
		if tag.SystemMustSet() {
			continue
		}
		pair := NameIDPair{
			Name: tag.FullyQualifiedName(),
			ID:   tag.ID,
		}
		pairs = append(pairs, pair)
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[j].Name > pairs[i].Name })
	return pairs, nil
}
