package core

// START HERE

// TODO: Define bag reader interface, similar to bag writer interface.
//       Then implement this so it's interchangeable with TarredBagReader.

type FileSystemBagReader struct {
	validator *Validator
}

func NewFileSystemBagReader(validator *Validator) (*FileSystemBagReader, error) {
	return &FileSystemBagReader{
		validator: validator,
	}, nil
}

func (reader *FileSystemBagReader) ScanMetadata() error {
	return nil
}

func (reader *FileSystemBagReader) ScanPayload() error {
	return nil
}

func (reader *FileSystemBagReader) Close() {
	// No op?
}
