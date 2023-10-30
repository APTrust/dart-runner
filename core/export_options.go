package core

import "github.com/APTrust/dart-runner/constants"

// ExportOptions returns options to populate lists on
// the export settings page.
type ExportOptions struct {
	AppSettings            []NameIDPair `json:"appSettings"`
	AppSettingFields       []string     `json:"appSettingFields"`
	BagItProfiles          []NameIDPair `json:"bagItProfiles"`
	RemoteRepositories     []NameIDPair `json:"remoteRepositories"`
	RemoteRepositoryFields []string     `json:"remoteRepositoryFields"`
	StorageServices        []NameIDPair `json:"storageServices"`
	StorageServiceFields   []string     `json:"storageServiceFields"`
}

// NewExportOptions returns an object containing the data
// we need to populate lists on the export settings page.
func NewExportOptions() *ExportOptions {
	return &ExportOptions{
		AppSettings:            ObjNameIdList(constants.TypeAppSetting),
		AppSettingFields:       []string{"Value"},
		BagItProfiles:          ObjNameIdList(constants.TypeBagItProfile),
		RemoteRepositories:     ObjNameIdList(constants.TypeRemoteRepository),
		RemoteRepositoryFields: []string{"APIToken", "LoginExtra", "Name", "Url", "UserID"},
		StorageServices:        ObjNameIdList(constants.TypeStorageService),
		StorageServiceFields:   []string{"AllowsDownload", "AllowsUpload", "Bucket", "Host", "Login", "LoginExtra", "Name", "Password", "Port", "Protocol"},
	}
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
			Name: tag.TagName,
			ID:   tag.ID,
		}
	}
	return pairs, nil
}
