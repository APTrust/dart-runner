package bagit

// Checksum describes the algorithm and digest of a file at path.
type Checksum struct {
	Algorithm string `json:"algorithm"`
	Digest    string `json:"digest"`
	Path      string `json:"path"`
}
