package validator

import (
	"fmt"
)

var ErrFileMissingFromBag = fmt.Errorf("file is missing from bag")

type FileRecord struct {
	Checksums []*Checksum
}

func NewFileRecord() *FileRecord {
	return &FileRecord{
		Checksums: make([]*Checksum, 0),
	}
}

func (fr *FileRecord) AddChecksum(checksum *Checksum) {
	fr.Checksums = append(fr.Checksums, checksum)
}

func (fr *FileRecord) Validate(algs []DigestAlgorithm) error {
	if !fr.FileExists() {
		return ErrFileMissingFromBag
	}
	for _, alg := range algs {
		if !fr.ManifestEntryExists(alg) {
			return fmt.Errorf("missing %s digest entry", algNames[alg])
		}
	}
	_, err := fr.ChecksumsMatch()
	return err
}

func (fr *FileRecord) ChecksumsMatch() (bool, error) {
	digest := ""
	source := ""
	for _, c := range fr.Checksums {
		if digest == "" {
			digest = c.Digest
			source = c.SourceName()
			continue
		}
		if c.Digest != digest {
			return false, fmt.Errorf("Digest %s in %s does not match digest %s in %s", c.Digest, c.SourceName(), digest, source)
		}
	}
	return true, nil
}

func (fr *FileRecord) FileExists() bool {
	for _, c := range fr.Checksums {
		if c.Source == SourcePayloadFile || c.Source == SourceTagFile {
			return true
		}
	}
	return false
}

func (fr *FileRecord) ManifestEntryExists(alg DigestAlgorithm) bool {
	for _, c := range fr.Checksums {
		if c.Algorithm == alg {
			return true
		}
	}
	return false
}
