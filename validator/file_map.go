package validator

import (
	"fmt"
)

// MaxErrors is the maximum number of errors the validator will
// collect before quitting and returning the error list. We don't
// quit at the first error because when developing a bagging
// process, multiple errors are common and we don't want to make
// depositors have to rebag constantly just to see the next error.
// We also try to be very specific with error messages so depositors
// know exactly what to fix.
const MaxErrors = 30

// FileMap contains a map of FileRecord objects and some methods to
// help validate those records.
type FileMap struct {
	Type  string
	Files map[string]*FileRecord
}

// NewFileMap returns a pointer to a new FileMap object.
func NewFileMap(fileType string) *FileMap {
	return &FileMap{
		Type:  fileType,
		Files: make(map[string]*FileRecord),
	}
}

// ValidateChecksums validates all checksums for all files in this
// FileMap. Param algs is a list of algorithms for manifests found
// in the bag.
func (fm *FileMap) ValidateChecksums(algs []string) []error {
	errors := make([]error, 0)
	for name, file := range fm.Files {
		err := file.Validate(fm.Type, algs)
		if err != nil {
			errors = append(errors, fmt.Errorf("File %s: %s", name, err.Error()))
			if len(errors) >= MaxErrors {
				break
			}
		}
	}
	return errors
}
