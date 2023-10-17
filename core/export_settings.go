package core

type ExportSettings struct {
	ID                 string              `json:"id"`
	AppSettings        []*AppSetting       `json:"appSettings"`
	BagItProfiles      []*BagItProfile     `json:"bagItProfiles"`
	Questions          []*ExportQuestion   `json:"questions"`
	RemoteRepositories []*RemoteRepository `json:"remoteRepositories"`
	StorageServices    []*StorageService   `json:"storageServices"`
}
