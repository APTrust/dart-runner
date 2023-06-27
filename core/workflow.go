package core

import (
	"encoding/json"
	"fmt"

	"github.com/APTrust/dart-runner/util"
)

type Workflow struct {
	ID              string            `json:"id"`
	BagItProfile    *BagItProfile     `json:"bagItProfile"`
	Description     string            `json:"description"`
	Errors          map[string]string `json:"-"`
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
	return &Workflow{
		ID:              w.ID,
		BagItProfile:    BagItProfileClone(w.BagItProfile),
		Description:     w.Description,
		Errors:          w.Errors,
		Name:            w.Name,
		PackageFormat:   w.PackageFormat,
		StorageServices: ssCopy,
	}

}
