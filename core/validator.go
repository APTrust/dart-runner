package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

type Validator struct {
	PathToBag          string
	Profile            *BagItProfile
	PayloadFiles       *FileMap
	PayloadManifests   *FileMap
	TagFiles           *FileMap
	TagManifests       *FileMap
	Tags               []*Tag
	UnparsableTagFiles []string
	Errors             map[string]string
	mapForType         map[string]*FileMap
	IgnoreOxumMismatch bool
}

func NewValidator(pathToBag string, profile *BagItProfile) (*Validator, error) {
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
		Tags:               make([]*Tag, 0),
		UnparsableTagFiles: make([]string, 0),
		Errors:             make(map[string]string),
		IgnoreOxumMismatch: false,
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
// That collection should contain an exhaustive list of errors,
// though it maxes out at 30 checksum validation errors.
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

	v.checkRequiredManifests()
	v.checkRequiredTagManifests()
	v.checkForbiddenManifests()
	v.checkForbiddenTagManifests()
	v.checkRequiredTagFiles()
	v.checkForbiddenTagFiles()
	v.validateTags()

	// Do this at the validation stage whether user says to
	// ignore mismatch or not. Ignoring only allows us to do
	// a full scan during the ScanBag() stage.
	v.AssertOxumsMatch()

	// Validate payload checksums
	algs, _ := v.PayloadManifestAlgs()
	errors := v.PayloadFiles.ValidateChecksums(algs)
	for key, value := range errors {
		v.Errors[key] = value
	}

	// Validate tag file checksums
	algs, _ = v.TagManifestAlgs()
	errors = v.TagFiles.ValidateChecksums(algs)
	for key, value := range errors {
		v.Errors[key] = value
	}

	return len(v.Errors) == 0
}

// ScanBag scans the bag's metadata and payload, recording file names,
// tag values, checksums, and errors. This will not run checksums on
// the payload if Payload-Oxum doesn't match because that's expensive.
// You can force checksum calculation here by setting
// Validator.IgnoreOxumMismatch to true. We will still flag the Oxum
// mismatch when you call Validate(), but you'll get to see which
// extra or missing files may be triggering the Oxum mismatch.
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
	// But user may want a full scan anyway. We can ignore this
	// now, but we'll flag it in Validate().
	if !v.IgnoreOxumMismatch {
		err = v.AssertOxumsMatch()
		if err != nil {
			return err
		}
	}

	return reader.ScanPayload()
}

func (v *Validator) AssertOxumsMatch() error {
	tags := v.GetTags("bag-info.txt", "Payload-Oxum")
	if len(tags) > 0 && v.PayloadFiles.Oxum() != tags[0].Value {
		err := fmt.Errorf("Payload-Oxum does not match payload")
		v.Errors["Payload-Oxum"] = err.Error()
		return err
	}
	return nil
}

func (v *Validator) GetTags(tagFile, tagName string) []*Tag {
	tags := make([]*Tag, 0)
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

func (v *Validator) checkRequiredManifests() bool {
	valid := true
	for _, alg := range v.Profile.ManifestsRequired {
		filename := fmt.Sprintf("manifest-%s.txt", alg)
		if _, ok := v.PayloadManifests.Files[filename]; !ok {
			v.Errors[filename] = "Required manifest is missing."
			valid = false
		}
	}
	return valid
}

func (v *Validator) checkRequiredTagManifests() bool {
	valid := true
	for _, alg := range v.Profile.TagManifestsRequired {
		filename := fmt.Sprintf("tagmanifest-%s.txt", alg)
		if _, ok := v.TagManifests.Files[filename]; !ok {
			v.Errors[filename] = "Required tag manifest is missing."
			valid = false
		}
	}
	return valid
}

func (v *Validator) checkForbiddenManifests() bool {
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

func (v *Validator) checkForbiddenTagManifests() bool {
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

func (v *Validator) checkRequiredTagFiles() bool {
	valid := true
	for _, filename := range v.Profile.TagFilesRequired {
		if _, ok := v.TagFiles.Files[filename]; !ok {
			v.Errors[filename] = "Required tag file is missing."
			valid = false
		}
	}
	return valid
}

func (v *Validator) checkForbiddenTagFiles() bool {
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
