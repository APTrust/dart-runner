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
		tagsForProfile, err := TagsForProfile(profile.ID)
		if err == nil {
			opts.BagItProfileFields[profile.ID] = tagsForProfile
		} else {
			opts.BagItProfileFields[profile.ID] = []NameIDPair{{Name: err.Error(), ID: profile.ID}}
		}
	}
	return opts
}

// TagsForProfile returns a list of tags defined in the BagIt profile
// with the specified ID.
func TagsForProfile(bagItProfileID string) ([]NameIDPair, error) {
	result := ObjFind(bagItProfileID)
	if result.Error != nil {
		return nil, result.Error
	}
	profile := result.BagItProfile()
	pairs := make([]NameIDPair, len(profile.Tags))
	for i, tag := range profile.Tags {
		pairs[i] = NameIDPair{
			Name: tag.FullyQualifiedName(),
			ID:   tag.ID,
		}
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[j].Name > pairs[i].Name })
	return pairs, nil
}
