package core

import (
	"fmt"
	"strings"

	"github.com/APTrust/dart-runner/util"
)

type UploadOperation struct {
	Errors           map[string]string  `json:"errors"`
	PayloadSize      int64              `json:"payloadSize"`
	Results          []*OperationResult `json:"results"`
	SourceFiles      []string           `json:"sourceFiles"`
	StorageServiceID string             `json:"storageServiceId"`
}

func NewUploadOperation() *UploadOperation {
	return &UploadOperation{
		Errors:      make(map[string]string),
		Results:     make([]*OperationResult, 0),
		SourceFiles: make([]string, 0),
	}
}

func (u *UploadOperation) Validate() bool {
	u.Errors = make(map[string]string)
	if !util.LooksLikeUUID(u.StorageServiceID) {
		u.Errors["UploadOperation.StorageServiceID"] = "UploadOperation requires a StorageServiceID"
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
