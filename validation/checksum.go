package validation

import (
	"fmt"

	"github.com/APTrust/dart-runner/constants"
)

// Checksum records information about a file's checksum.
type Checksum struct {
	Source    string
	Algorithm string
	Digest    string
}

func NewChecksum(source, alg, digest string) *Checksum {
	return &Checksum{
		Source:    source,
		Algorithm: alg,
		Digest:    digest,
	}
}

func (cs *Checksum) SourceName() string {
	switch cs.Source {
	case constants.FileTypeManifest:
		return fmt.Sprintf("manifest-%s.txt", cs.Algorithm)
	case constants.FileTypeTagManifest:
		return fmt.Sprintf("tagmanifest-%s.txt", cs.Algorithm)
	default:
		return cs.Source
	}
}
