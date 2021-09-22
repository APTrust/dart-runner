package validation

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
	Tags             []*bagit.Tag
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
		Tags:             make([]*bagit.Tag, 0),
		Errors:           make([]error, 0),
	}, nil
}

// Validate validates the bag and returns true if it's valid.
// If this returns false, check the errors in Validator.Errors.
// The validator quits after 30 errors.
func (v *Validator) Validate() bool {
	// Make sure BagItProfile is present and valid.
	// Make sure bag has valid serialization format, per profile.

	// Scan the bag.

	// Make sure required manifests are present.
	// Make sure required tag manifests are present.
	// Make sure all existing manifests are allowed.
	// Make sure all existing tag manifests are allowed.
	// Make sure existing tag files are allowed.
	// Validate payload checksums
	// Validate tag file checksums
	// Validate payload oxum
	// Validate tags

	return true
}

// ScanBag scans the bag's metadata and payload, recording file names,
// tag values, checksums, and errors.
func (v *Validator) ScanBag() error {

	return nil
}