package validation

import (
	"os"
	"strings"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

type Validator struct {
	PathToBag          string
	Profile            *bagit.Profile
	PayloadFiles       *FileMap
	PayloadManifests   *FileMap
	TagFiles           *FileMap
	TagManifests       *FileMap
	Tags               []*bagit.Tag
	UnparsableTagFiles []string
	Errors             []error
	mapForType         map[string]*FileMap
}

func NewValidator(pathToBag string) (*Validator, error) {
	if !util.FileExists(pathToBag) {
		return nil, os.ErrNotExist
	}
	validator := &Validator{
		PathToBag:          pathToBag,
		PayloadFiles:       NewFileMap(constants.FileTypePayload),
		PayloadManifests:   NewFileMap(constants.FileTypeManifest),
		TagFiles:           NewFileMap(constants.FileTypeTag),
		TagManifests:       NewFileMap(constants.FileTypeTagManifest),
		Tags:               make([]*bagit.Tag, 0),
		UnparsableTagFiles: make([]string, 0),
		Errors:             make([]error, 0),
	}
	validator.mapForType = map[string]*FileMap{
		constants.FileTypePayload:     validator.PayloadFiles,
		constants.FileTypeManifest:    validator.PayloadManifests,
		constants.FileTypeTag:         validator.TagFiles,
		constants.FileTypeTagManifest: validator.TagManifests,
	}
	return validator, nil
}

func (v *Validator) MapForPath(pathInBag string) *FileMap {
	return v.mapForType[util.BagFileType(pathInBag)]
}

func (v *Validator) PayloadManifestAlgs() ([]string, error) {
	return v.manifestAlgs(v.PayloadManifests)
}

func (v *Validator) TagManifestAlgs() ([]string, error) {
	return v.manifestAlgs(v.TagManifests)
}

func (v *Validator) manifestAlgs(fileMap *FileMap) ([]string, error) {
	algs := make([]string, 0)
	for name, _ := range fileMap.Files {
		alg, err := util.AlgorithmFromManifestName(name)
		if err != nil {
			return nil, err
		}
		algs = append(algs, alg)
	}
	return algs, nil
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

func (v *Validator) CompareOxums() error {

	return nil
}

func (v *Validator) GetTags(tagFile, tagName string) []*bagit.Tag {
	tags := make([]*bagit.Tag, 0)
	for _, tag := range v.Tags {
		if tag.TagFile == tagFile && tag.TagName == tagName {
			tags = append(tags, tag)
		} else if tag.TagFile == tagFile && strings.EqualFold(tag.TagName, tagName) {
			tags = append(tags, tag)
		}
	}
	return tags
}
