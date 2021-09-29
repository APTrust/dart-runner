package util

// BagWriter describes the interface that bag writers must implement,
// whether writing to a file system, tar file, zip file, etc.
type BagWriter interface {
	// Open opens the writer.
	Open() error

	// AddFile adds a file to the bag. The returned map has digest
	// alg names for keys and digests for values. For example,
	// checksums["md5"] = "0987654321".
	AddFile(*ExtendedFileInfo, string) (map[string]string, error)

	// Close closes the underlying writer, flushing remaining data
	// as necessary.
	Close() error
}
