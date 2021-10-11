package core

import (
	"encoding/json"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/APTrust/dart-runner/util"
)

type Workflow struct {
	ID              string            `json:"id"`
	BagItProfile    *bagit.Profile    `json:"bagItProfile"`
	Description     string            `json:"description"`
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
