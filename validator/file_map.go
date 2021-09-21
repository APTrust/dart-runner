package validator

import (
	"fmt"
)

type FileType int

const (
	FileTypePayload FileType = iota
	FileTypeManifest
	FileTypeTagFile
	FileTypeTagManifest
)

var prefixFor = []string{
	"",
	"",
	"",
	"tag-",
}

const MaxErrors = 30

type FileMap struct {
	Type  FileType
	Files map[string]*FileRecord
}

func NewFileMap(fileType FileType) *FileMap {
	return &FileMap{
		Type:  fileType,
		Files: make(map[string]*FileRecord),
	}
}

func (fm *FileMap) ValidateChecksums(algs []DigestAlgorithm) []error {
	errors := make([]error, 0)
	for name, file := range fm.Files {
		err := file.Validate(algs)
		if err != nil {
			errors = append(errors, fmt.Errorf("File %s: %s", name, err.Error()))
			if len(errors) >= MaxErrors {
				break
			}
		}
	}
	return errors
}
