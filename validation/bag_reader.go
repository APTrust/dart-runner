package validation

// BagReader describes the interface that bag readers must implement,
// whether reading from a file system, tar file, zip file, etc.
type BagReader interface {
	// ScanMetadata scans and parses the tag files and manifests.
	// We do this first so we know which checksums to calculate when we
	// scan the payload later. Many BagIt profiles allow five or six
	// different manifest algorithms, but the bag may contain only one
	// or two from the allowed list. We won't know what they are until
	// we look.
	//
	// If you want to do a quick validation, you can scan the metadata
	// and then compare the Validator's oxum against the one in bag-info.txt.
	// If that's bad, you can skip the expensive payload scan.
	ScanMetadata() error

	// ScanPayload calculates checksums on the payload and tag files.
	ScanPayload() error

	// Close closes the underlying reader.
	Close()
}
