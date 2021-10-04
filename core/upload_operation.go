package core

import (
	"fmt"
	"strings"

	"github.com/APTrust/dart-runner/util"
)

type UploadOperation struct {
	Errors         map[string]string  `json:"errors"`
	PayloadSize    int64              `json:"payloadSize"`
	Results        []*OperationResult `json:"results"`
	SourceFiles    []string           `json:"sourceFiles"`
	StorageService *StorageService    `json:"storageService"`
}

func NewUploadOperation(ss *StorageService, files []string) *UploadOperation {
	return &UploadOperation{
		Errors:         make(map[string]string),
		Results:        make([]*OperationResult, 0),
		SourceFiles:    files,
		StorageService: ss,
	}
}

func (u *UploadOperation) Validate() bool {
	u.Errors = make(map[string]string)
	if u.StorageService == nil {
		u.Errors["UploadOperation.StorageService"] = "UploadOperation requires a StorageService"
	} else if u.StorageService.Validate() == false {
		u.Errors = u.StorageService.Errors
	}
	missingFiles := make([]string, 0)
	for _, file := range u.SourceFiles {
		if !util.FileExists(file) {
			missingFiles = append(missingFiles, file)
		}
	}
	if len(missingFiles) > 0 {
		u.Errors["UploadOperation.SourceFiles"] = fmt.Sprintf("UploadOperation source files are missing: %s", strings.Join(missingFiles, ";"))
	}
	return len(u.Errors) == 0
}
