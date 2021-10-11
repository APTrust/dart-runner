package core

import (
	"encoding/json"
	"fmt"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/APTrust/dart-runner/util"
)

type Workflow struct {
	ID              string            `json:"id"`
	BagItProfile    *bagit.Profile    `json:"bagItProfile"`
	Description     string            `json:"description"`
	Errors          map[string]string `json:-`
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

func (w *Workflow) Validate() bool {
	w.Errors = make(map[string]string)
	if w.BagItProfile != nil && !w.BagItProfile.IsValid() {
		for key, value := range w.BagItProfile.Errors {
			w.Errors[key] = value
		}
	}
	if w.StorageServices != nil {
		for _, ss := range w.StorageServices {
			if !ss.Validate() {
				for key, value := range ss.Errors {
					uniqueKey := fmt.Sprintf("%s.%s", ss.Name, key)
					w.Errors[uniqueKey] = value
				}
			}
		}
	}
	return len(w.Errors) == 0
}
