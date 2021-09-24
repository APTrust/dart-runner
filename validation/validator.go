package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
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
	Errors             map[string]string
	mapForType         map[string]*FileMap
}

func NewValidator(pathToBag string, profile *bagit.Profile) (*Validator, error) {
	if !util.FileExists(pathToBag) {
		return nil, os.ErrNotExist
	}
	validator := &Validator{
		PathToBag:          pathToBag,
		PayloadFiles:       NewFileMap(constants.FileTypePayload),
		PayloadManifests:   NewFileMap(constants.FileTypeManifest),
		Profile:            profile,
		TagFiles:           NewFileMap(constants.FileTypeTag),
		TagManifests:       NewFileMap(constants.FileTypeTagManifest),
		Tags:               make([]*bagit.Tag, 0),
		UnparsableTagFiles: make([]string, 0),
		Errors:             make(map[string]string),
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

// Validate validates the bag and returns true if it's valid.
// If this returns false, check the errors in Validator.Errors.
// The validator quits after 30 errors.
func (v *Validator) Validate() bool {
	// Make sure BagItProfile is present and valid.
	if !v.Profile.IsValid() {
		v.Errors = v.Profile.Errors
		return false
	}
	// Make sure bag has valid serialization format, per profile.
	if !v.validateSerialization() {
		return false
	}

	// Scan the bag.

	// Make sure required manifests are present.
	if !v.hasRequiredManifests() {
		return false
	}
	// Make sure required tag manifests are present.
	if !v.hasRequiredTagManifests() {
		return false
	}

	// Make sure all existing manifests are allowed.
	if v.hasForbiddenManifests() {
		return false
	}

	// Make sure all existing tag manifests are allowed.
	if v.hasForbiddenTagManifests() {
		return false
	}

	// Make sure we have all required tag files
	if !v.hasRequiredTagFiles() {
		return false
	}

	// Make sure existing tag files are allowed.
	if v.hasForbiddenTagFiles() {
		return false
	}

	// Validate tags
	if !v.validateTags() {
		return false
	}

	// Validate payload checksums
	algs, _ := v.PayloadManifestAlgs()
	errors := v.PayloadFiles.ValidateChecksums(algs)
	if len(errors) > 0 {
		v.Errors = errors
		return false
	}

	// Validate tag file checksums
	algs, _ = v.TagManifestAlgs()
	errors = v.TagFiles.ValidateChecksums(algs)
	if len(errors) > 0 {
		v.Errors = errors
		return false
	}

	return true
}

// ScanBag scans the bag's metadata and payload, recording file names,
// tag values, checksums, and errors.
func (v *Validator) ScanBag() error {
	reader, err := v.getReader()
	if err != nil {
		return err
	}
	defer reader.Close()
	err = reader.ScanMetadata()
	if err != nil {
		return err
	}

	// If Payload-Oxum doesn't match, there's no sense in doing
	// the heavy work of calculating checksums on the payload.
	ok, err := v.OxumsMatch()
	if !ok && err == nil {
		return fmt.Errorf("Payload-Oxum does not match payload")
	}
	return reader.ScanPayload()
}

func (v *Validator) OxumsMatch() (bool, error) {
	tags := v.GetTags("bag-info.txt", "Payload-Oxum")
	if len(tags) == 0 {
		return false, ErrTagNotFound
	}
	return v.PayloadFiles.Oxum() == tags[0].Value, nil
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

// In future, this will return a reader based on the file type.
// For now, we only support tar files, so this always returns a
// tar reader.
func (v *Validator) getReader() (BagReader, error) {
	return NewTarredBagReader(v)
}

func (v *Validator) validateSerialization() bool {
	bagIsDirectory := util.IsDirectory(v.PathToBag)
	if v.Profile.Serialization == constants.SerializationRequired && bagIsDirectory {
		v.Errors["Serialization"] = "Profile says bag must be serialized, but it is a directory."
		return false
	} else if v.Profile.Serialization == constants.SerializationForbidden && !bagIsDirectory {
		v.Errors["Serialization"] = "Profile says bag must not be serialized, but bag is not a directory."
		return false
	}
	if !bagIsDirectory {
		var err error
		var ok bool
		for _, mimeType := range v.Profile.AcceptSerialization {
			ok, err = util.HasValidExtensionForMimeType(v.PathToBag, mimeType)
			if ok {
				break
			}
		}
		if err != nil {
			v.Errors["Serialization"] = err.Error()
			return false
		} else if !ok {
			ext := path.Ext(v.PathToBag)
			v.Errors["Serialization"] = fmt.Sprintf("Bag has extension %s, but profile says it must be serialized as of one of the following types: %s.", ext, strings.Join(v.Profile.AcceptSerialization, ","))
			return false
		}
	}
	return true
}

func (v *Validator) hasRequiredManifests() bool {
	valid := true
	for _, filename := range v.Profile.ManifestsRequired {
		if _, ok := v.PayloadManifests.Files[filename]; !ok {
			v.Errors[filename] = "Required manifest is missing."
			valid = false
		}
	}
	return valid
}

func (v *Validator) hasRequiredTagManifests() bool {
	valid := true
	for _, filename := range v.Profile.TagManifestsRequired {
		if _, ok := v.TagManifests.Files[filename]; !ok {
			v.Errors[filename] = "Required tag manifest is missing."
			valid = false
		}
	}
	return valid
}

func (v *Validator) hasForbiddenManifests() bool {
	hasForbidden := false
	// We can ignore error here. We would have hit it earlier,
	// while scanning the bag.
	algs, _ := v.PayloadManifestAlgs()
	for _, alg := range algs {
		if !util.StringListContains(v.Profile.ManifestsAllowed, alg) {
			filename := fmt.Sprintf("manifest-%s.txt", alg)
			v.Errors[filename] = "Payload manifest is forbidden by profile."
			hasForbidden = true
		}
	}
	return hasForbidden
}

func (v *Validator) hasForbiddenTagManifests() bool {
	hasForbidden := false
	// We can ignore error here. We would have hit it earlier,
	// while scanning the bag.
	algs, _ := v.TagManifestAlgs()
	for _, alg := range algs {
		if !util.StringListContains(v.Profile.TagManifestsAllowed, alg) {
			filename := fmt.Sprintf("tagmanifest-%s.txt", alg)
			v.Errors[filename] = "Tag manifest is forbidden by profile."
			hasForbidden = true
		}
	}
	return hasForbidden
}

func (v *Validator) hasRequiredTagFiles() bool {
	valid := true
	for _, filename := range v.Profile.TagFilesRequired {
		if _, ok := v.TagFiles.Files[filename]; !ok {
			v.Errors[filename] = "Required tag file is missing."
			valid = false
		}
	}
	return valid
}

func (v *Validator) hasForbiddenTagFiles() bool {
	for _, pattern := range v.Profile.TagFilesAllowed {
		if strings.TrimSpace(pattern) == "*" {
			return false
		}
	}
	hasForbidden := false
	for filename, _ := range v.TagFiles.Files {
		var err error
		fileMatches := false
		fileWasTested := false
		for _, pattern := range v.Profile.TagFilesAllowed {
			if strings.TrimSpace(pattern) == "" {
				continue
			}
			// Should probably replace * with .* in pattern
			// to get a valid regex match.
			rePattern := strings.ReplaceAll(pattern, "*", ".*")
			fileMatches, err = regexp.MatchString(rePattern, filename)
			if err != nil {
				v.Errors[pattern] = "Cannot match tag file names against this pattern."
				return false // no use continuing if we can't do our job
			}
			if fileMatches {
				break
			}
		}
		if fileWasTested && !fileMatches {
			v.Errors[filename] = fmt.Sprintf("Tag file %s is not in the list of allowed tag files.", filename)
			hasForbidden = true
		}
	}
	return hasForbidden
}

func (v *Validator) validateTags() bool {
	valid := true
	for _, tagDef := range v.Profile.Tags {
		key := fmt.Sprintf("%s/%s", tagDef.TagFile, tagDef.TagName)
		tags := v.GetTags(tagDef.TagFile, tagDef.TagName)
		if len(tags) == 0 && tagDef.Required {
			v.Errors[key] = "Required tag is missing."
			valid = false
			continue
		}
		hasValue := false
		for _, tag := range tags {
			if tag.Value != "" {
				hasValue = true
			}
			if !tagDef.IsLegalValue(tag.Value) {
				v.Errors[key] = fmt.Sprintf("Tag has illegal value '%s'. Allowed values are: %s", tag.Value, strings.Join(tagDef.Values, ","))
				valid = false
			}
		}
		if tagDef.Required && !tagDef.EmptyOK && !hasValue {
			v.Errors[key] = "Required tag is present but has no value."
			valid = false
		}
	}
	return valid
}

func (v *Validator) ErrorString() string {
	errs := make([]string, len(v.Errors))
	i := 0
	for k, v := range v.Errors {
		errs[i] = fmt.Sprintf("%s -> %s", k, v)
		i++
	}
	return strings.Join(errs, "\n")
}

func (v *Validator) ErrorJSON() string {
	data, _ := json.Marshal(v.Errors)
	return string(data)
}
