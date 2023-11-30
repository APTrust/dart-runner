package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/APTrust/dart-runner/constants"
	"github.com/google/uuid"
)

type Artifact struct {
	ID        string
	JobID     string
	BagName   string
	ItemType  string // File or JobResult
	FileName  string // name of manifest or tag file
	FileType  string // manifest or tag file
	RawData   string // file content or work result json
	UpdatedAt time.Time
}

// NewArtifact creates a new empty Artifact with a unique id and timestamp.
func NewArtifact() *Artifact {
	return &Artifact{
		ID:        uuid.NewString(),
		UpdatedAt: time.Now(),
	}
}

// NewJobResultArtifact creates a new Artifact to store a JobResult.
func NewJobResultArtifact(bagName string, jobResult *JobResult) *Artifact {
	resultJson, _ := json.MarshalIndent(jobResult, "", "  ")
	return &Artifact{
		ID:        uuid.NewString(),
		JobID:     jobResult.JobID,
		BagName:   bagName,
		ItemType:  constants.ItemTypeJobResult,
		FileName:  fmt.Sprintf("Job Result %s", jobResult.JobName),
		FileType:  constants.FileTypeJsonData,
		RawData:   string(resultJson),
		UpdatedAt: time.Now(),
	}
}

// NewManifestArtifact creates a new Artifact to store a bag's payload manifest.
func NewManifestArtifact(bagName, jobID, manifestName, manifestContent string) *Artifact {
	return &Artifact{
		ID:        uuid.NewString(),
		JobID:     jobID,
		BagName:   bagName,
		ItemType:  constants.ItemTypeManifest,
		FileName:  manifestName,
		FileType:  constants.FileTypeManifest,
		RawData:   manifestContent,
		UpdatedAt: time.Now(),
	}
}

// NewTagFileArtifact creates a new Artifact to store a bag's tag file.
func NewTagFileArtifact(bagName, jobID, tagFileName, tagFileContent string) *Artifact {
	return &Artifact{
		ID:        uuid.NewString(),
		JobID:     jobID,
		BagName:   bagName,
		ItemType:  constants.ItemTypeTagFile,
		FileName:  tagFileName,
		FileType:  constants.FileTypeTag,
		RawData:   tagFileContent,
		UpdatedAt: time.Now(),
	}
}
