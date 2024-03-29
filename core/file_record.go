package core

import (
	"fmt"

	"github.com/APTrust/dart-runner/constants"
)

// ErrFileMissingFromBag indicates that a file present in a payload
// manifest is not present in the bag's data directory.
var ErrFileMissingFromBag = fmt.Errorf("file is missing from bag")

// FileRecord contains a collection of checksums for a single file.
type FileRecord struct {
	Size      int64
	Checksums []*Checksum
}

// NewFileRecord returns a pointer to a new FileRecord object.
func NewFileRecord() *FileRecord {
	return &FileRecord{
		Checksums: make([]*Checksum, 0),
	}
}

// AddChecksum adds a checksum to this FileRecord.
func (fr *FileRecord) AddChecksum(source, alg, digest string) {
	fr.Checksums = append(fr.Checksums, NewChecksum(source, alg, digest))
}

// Validate validates the following about the current file:
//
//   - file is present in the payload directory
//   - file is listed in the payload manifests matching the
//     specified algorithms.
//   - the checksums that the validator calculated on the file itself
//     match the checksums in the manifests.
//
// Validate returns true if the checksums we calculated for the
// file match the checksums in the manifests.
func (fr *FileRecord) Validate(fileType string, algs []string) error {
	// existingAlgorithms := fr.DigestAlgorithms()
	srcFile := constants.FileTypePayload
	srcManifest := constants.FileTypeManifest
	if fileType == constants.FileTypeTag {
		srcFile = constants.FileTypeTag
		srcManifest = constants.FileTypeTagManifest
	}
	for _, alg := range algs {
		fileChecksum := fr.GetChecksum(alg, srcFile)
		if fileChecksum == nil {
			return ErrFileMissingFromBag
		}
		manifestChecksum := fr.GetChecksum(alg, srcManifest)
		if srcManifest == constants.FileTypeTagManifest && manifestChecksum == nil {
			continue // tag files don't have to appear in tag manifests
		}
		if manifestChecksum == nil {
			manifestName := "manifest"
			if srcFile == constants.FileTypeTag {
				manifestName = "tagmanifest"
			}
			return fmt.Errorf("file is missing from %s-%s.txt", manifestName, alg)
		}
		if fileChecksum.Digest != manifestChecksum.Digest {
			return fmt.Errorf("Digest %s in %s does not match digest %s in %s", manifestChecksum.Digest, manifestChecksum.SourceName(), fileChecksum.Digest, fileChecksum.SourceName())
		}
	}
	return nil
}

// GetChecksum returns the checksum with the specified algorithm
// and source.
func (fr *FileRecord) GetChecksum(alg, source string) *Checksum {
	for _, cs := range fr.Checksums {
		if cs.Algorithm == alg && cs.Source == source {
			return cs
		}
	}
	return nil
}

// DigestAlgorithms returns a list of digest algorithms calculated for
// this file.
func (fr *FileRecord) DigestAlgorithms() []string {
	alreadyAdded := make(map[string]bool)
	algs := make([]string, 0)
	for _, cs := range fr.Checksums {
		if !alreadyAdded[cs.Algorithm] {
			algs = append(algs, cs.Algorithm)
			alreadyAdded[cs.Algorithm] = true
		}
	}
	return algs
}
