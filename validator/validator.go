package validator

import (
	"os"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

type Validator struct {
	PathToBag        string
	Profile          *bagit.Profile
	PayloadFiles     *FileMap
	PayloadManifests *FileMap
	TagFiles         *FileMap
	TagManifests     *FileMap
	Errors           []error
}

func NewValidator(pathToBag string) (*Validator, error) {
	if !util.FileExists(pathToBag) {
		return nil, os.ErrNotExist
	}
	return &Validator{
		PathToBag:        pathToBag,
		PayloadFiles:     NewFileMap(constants.FileTypePayload),
		PayloadManifests: NewFileMap(constants.FileTypeManifest),
		TagFiles:         NewFileMap(constants.FileTypeTag),
		TagManifests:     NewFileMap(constants.FileTypeTagManifest),
		Errors:           make([]error, 0),
	}, nil
}
