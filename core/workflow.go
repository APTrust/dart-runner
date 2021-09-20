package core

type Workflow struct {
	BagItProfileID    string   `json:"bagItProfileId"`
	Description       string   `json:"description"`
	Name              string   `json:"name"`
	PackageFormat     string   `json:"packageFormat"`
	PackagePluginID   string   `json:"packagePluginId"`
	StorageServiceIDs []string `json:"storageServiceIds"`
}
