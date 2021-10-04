package core

import (
	"github.com/APTrust/dart-runner/bagit"
)

type Workflow struct {
	ID              string            `json:"id"`
	BagItProfile    *bagit.Profile    `json:"bagItProfile"`
	Description     string            `json:"description"`
	Name            string            `json:"name"`
	PackageFormat   string            `json:"packageFormat"`
	StorageServices []*StorageService `json:"storageServices"`
}
