package validator

import (
	"fmt"
)

type ChecksumSource int
type DigestAlgorithm int

const (
	SourceManifest ChecksumSource = iota
	SourceTagManifest
	SourcePayloadFile
	SourceTagFile
)

var sourceNames = []string{
	"manifest",
	"tagmanifest",
	"payload file",
	"tag file",
}

const (
	AlgMd5 DigestAlgorithm = iota
	AlgSha1
	AlgSha256
	AlgSha512
)

// Match algorithm enum to human-readable name.
var algNames = []string{
	"md5",
	"sha1",
	"sha256",
	"sha512",
}

func AlgToEnum(name string) (DigestAlgorithm, error) {
	for i, alg := range algNames {
		if alg == name {
			return DigestAlgorithm(i), nil
		}
	}
	return 0, fmt.Errorf("unsupported digest algorithm: %s", name)
}

type Checksum struct {
	Source    ChecksumSource
	Algorithm DigestAlgorithm
	Digest    string
}

func NewChecksum(source ChecksumSource, alg DigestAlgorithm, digest string) *Checksum {
	return &Checksum{
		Source:    source,
		Algorithm: alg,
		Digest:    digest,
	}
}

func (c *Checksum) SourceName() string {
	if c.Source == SourcePayloadFile || c.Source == SourceTagFile {
		return sourceNames[c.Source]
	}
	return fmt.Sprintf("%s-%s.txt", sourceNames[c.Source], algNames[c.Algorithm])
}
