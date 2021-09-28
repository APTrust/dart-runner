package util

// BagWriter describes the interface that bag writers must implement,
// whether writing to a file system, tar file, zip file, etc.
type BagWriter interface {
	// Open opens the writer.
	Open() error

	// AddFile adds a file to the bag.
	AddFile(*ExtendedFileInfo, string) error

	// Close closes the underlying writer, flushing remaining data
	// as necessary.
	Close() error
}
