package core

import (
	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

// BagWriter describes the interface that bag writers must implement,
// whether writing to a file system, tar file, zip file, etc.
type BagWriter interface {
	// Open opens the writer.
	Open() error

	// AddFile adds a file to the bag. The returned map has digest
	// alg names for keys and digests for values. For example,
	// checksums["md5"] = "0987654321".
	AddFile(*util.ExtendedFileInfo, string) (map[string]string, error)

	// DigestAlgs returns a list of digest algoritms that the
	// writer calculates as it writes. E.g. ["md5", "sha256"].
	// These are defined in the contants package.
	DigestAlgs() []string

	// Close closes the underlying writer, flushing remaining data
	// as necessary.
	Close() error
}

// GetBagWriter returns a bag writer of the specified type.
// Valid types include constants.BagWriterTypeFileSystem and
// constants.BagWriterTypeTar.
func GetBagWriter(writerType, outputPath string, digestAlgs []string) (BagWriter, error) {
	switch writerType {
	case constants.BagWriterTypeFileSystem:
		return NewFileSystemBagWriter(outputPath, digestAlgs), nil
	case constants.BagWriterTypeTar:
		return NewTarredBagWriter(outputPath, digestAlgs), nil
	default:
		return nil, constants.ErrUnknownType
	}
}
