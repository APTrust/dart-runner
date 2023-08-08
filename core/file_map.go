package core

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/APTrust/dart-runner/constants"
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
	Type           string
	Files          map[string]*FileRecord
	MessageChannel chan *EventMessage
}

// NewFileMap returns a pointer to a new FileMap object.
//
// TODO: Deprecate this. New version should always use channel.
func NewFileMap(fileType string) *FileMap {
	return &FileMap{
		Type:  fileType,
		Files: make(map[string]*FileRecord),
	}
}

// NewFileMapWithChannel returns a pointer to a new FileMap object.
// This object has a message channel to pass progress information back to
// the front end.
func NewFileMapWithChannel(fileType string, messageChannel chan *EventMessage) *FileMap {
	return &FileMap{
		Type:           fileType,
		Files:          make(map[string]*FileRecord),
		MessageChannel: messageChannel,
	}
}

// ValidateChecksums validates all checksums for all files in this
// FileMap. Param algs is a list of algorithms for manifests found
// in the bag.
func (fm *FileMap) ValidateChecksums(algs []string) map[string]string {
	errors := make(map[string]string)
	for name, file := range fm.Files {
		if fm.MessageChannel != nil {
			fm.MessageChannel <- InfoEvent(constants.StageValidate, fmt.Sprintf("Validating %s", name))
		}
		err := file.Validate(fm.Type, algs)
		if err != nil {
			errors[name] = err.Error()
			if len(errors) >= MaxErrors {
				break
			}
		}
	}
	return errors
}

func (fm *FileMap) FileCount() int64 {
	return int64(len(fm.Files))
}

func (fm *FileMap) TotalBytes() int64 {
	sum := int64(0)
	for _, fr := range fm.Files {
		sum += fr.Size
	}
	return sum
}

func (fm *FileMap) Oxum() string {
	return fmt.Sprintf("%d.%d", fm.TotalBytes(), fm.FileCount())
}

// WriteManifest is used during bagging to write a manifest.
// Param fileType should be either constants.FileTypePayload
// or constants.FileTypeTag, depending on whether you're writing
// a payload manifest or a tag manifest. Param alg is the digest
// algorithm. Those are defined in constants. ("md5", "sha256", etc)
//
// Param trimFromPath is an optional prefix to remove from file paths
// when writing the manifest. For example, tarred bags include an additional
// top-level directory that matches the bag name, but this directory should
// not be included in the manifest paths.
//
// What appears in the tar file as bag_name/data/file.txt should appear in the
// manifest as data/file.txt. To make that happen, pass "bag_name/" as the
// last param, and it will be trimmed. Pass an empty string if you don't
// need to trim leading paths.
func (fm *FileMap) WriteManifest(writer io.Writer, fileType, alg, trimFromPath string) error {
	sortedNames := make([]string, len(fm.Files))
	i := 0
	for filename, _ := range fm.Files {
		sortedNames[i] = filename
		i++
	}
	sort.Strings(sortedNames)
	for _, filename := range sortedNames {
		cs := fm.Files[filename].GetChecksum(alg, fileType)
		if cs == nil {
			return fmt.Errorf("Missing %s digest for %s [%s]", alg, filename, fileType)
		}
		trimmedFileName := strings.Replace(filename, trimFromPath, "", 1)
		entry := fmt.Sprintf("%s  %s\n", cs.Digest, trimmedFileName)
		n, err := writer.Write([]byte(entry))
		if err != nil {
			return err
		}
		if n != len(entry) {
			return fmt.Errorf("Wrote %d of %d bytes for %s entry", n, len(entry), filename)
		}
	}
	return nil
}
