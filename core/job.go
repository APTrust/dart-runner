package core

import (
	"github.com/APTrust/dart-runner/bagit"
)

// TitleTags is a list of BagItProfile tags to check to try to find a
// meaningful title for this job. DART checks them in order and
// returns the first one that has a non-empty user-defined value.
var TitleTags = []string{
	"Title",
	"Internal-Sender-Identifier",
	"External-Identifier",
	"Internal-Sender-Description",
	"External-Description",
	"Description",
}

type Job struct {
	BagItProfile *bagit.Profile       `json:"bagItProfile"`
	ByteCount    int64                `json:"dirCount"`
	DirCount     int                  `json:"dirCount"`
	Errors       map[string]string    `json:"errors"`
	FileCount    int                  `json:"fileCount"`
	PackageOp    *PackageOperation    `json:"packageOp"`
	UploadOps    []*UploadOperation   `json:"uploadOps"`
	ValidationOp *ValidationOperation `json:"validationOp"`
	WorkflowID   string               `json:"workflowId"`
}
