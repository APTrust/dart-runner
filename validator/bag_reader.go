package validator

// BagReader describes the interface that bag readers must implement,
// whether reading from a file system, tar file, zip file, etc.
type BagReader interface {
	// ScanMetadata scans and parses the tag files and manifests.
	// We do this first so we know which checksums to calculate when we
	// scan the payload later. Many BagIt profiles allow five or six
	// different manifest algorithms, but the bag may contain only one
	// or two from the allowed list. We won't know what they are until
	// we look.
	ScanMetadata(v *Validator) error

	// ScanPayload calculates checksums on the payload and tag files.
	ScanPayload(v *Validator) error
}
