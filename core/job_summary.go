package core

import (
	"os"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// JobSummary contains info about a job to be displayed
// in the front end. All of this info appears on the job_run
// page and the batch_run page. It may appear elsewhere as
// well. Note that it takes some logic to extract some of this
// information. We don't want to force that logic into HTML
// templates and front-end JavaScript, so we do it here, where
// it's easy to test.
type JobSummary struct {
	ID                      string   `json:"id"`
	JobType                 string   `json:"jobType"`
	Name                    string   `json:"name"`
	HasPackageOp            bool     `json:"hasPackageOp"`
	HasUploadOps            bool     `json:"hasUploadOps"`
	HasBagItProfile         bool     `json:"hasBagItProfile"`
	PackageFormat           string   `json:"packageFormat"`
	PackageName             string   `json:"packageName"`
	BagItProfileName        string   `json:"bagItProfileName"`
	BagItProfileDescription string   `json:"bagItProfileDescription"`
	DirectoryCount          int64    `json:"directoryCount"`
	PayloadFileCount        int64    `json:"payloadFileCount"`
	ByteCount               int64    `json:"byteCount"`
	ByteCountFormatted      string   `json:"byteCountFormatted"`
	ByteCountHuman          string   `json:"byteCountHuman"`
	SourceFiles             []string `json:"sourceFiles"`
	OutputPath              string   `json:"outputPath"`
	UploadTargets           []string `json:"uploadTargets"`
	PathSeparator           string   `json:"pathSeparator"`
}

// NewJobSummary creates a new JobSummaryInfo object based on
// the given job.
//
// NOTE: To get accurate info about the job's payload, call
// job.UpdatePayloadStats() before calling this constructor.
func NewJobSummary(job *Job) *JobSummary {
	info := &JobSummary{
		ID:               job.ID,
		JobType:          constants.TypeJob,
		Name:             job.Name(),
		HasPackageOp:     job.HasPackageOp(),
		HasUploadOps:     job.HasUploadOps(),
		HasBagItProfile:  job.BagItProfile != nil,
		DirectoryCount:   job.DirCount,
		PayloadFileCount: job.PayloadFileCount,
		ByteCount:        job.ByteCount,
		ByteCountHuman:   util.HumanSize(job.ByteCount),
		PathSeparator:    string(os.PathSeparator),
	}
	// Byte count with commas, for more readable display
	p := message.NewPrinter(language.English)
	info.ByteCountFormatted = p.Sprintf("%d", job.ByteCount)

	// Packaging info
	if job.HasPackageOp() {
		info.OutputPath = job.PackageOp.OutputPath
		info.PackageFormat = job.PackageOp.PackageFormat
		info.PackageName = job.PackageOp.PackageName
		info.SourceFiles = job.PackageOp.SourceFiles
	}

	// BagIt profile info
	if job.BagItProfile != nil {
		info.BagItProfileName = job.BagItProfile.Name
		info.BagItProfileDescription = job.BagItProfile.Description
	}

	// Upload info
	info.UploadTargets = make([]string, len(job.UploadOps))
	for i, op := range job.UploadOps {
		info.UploadTargets[i] = op.StorageService.Name
	}

	return info
}

func NewValidationJobSummary(valJob *ValidationJob, profile *BagItProfile) *JobSummary {
	return &JobSummary{
		ID:                      valJob.ID,
		JobType:                 constants.TypeValidationJob,
		BagItProfileDescription: profile.Description,
		BagItProfileName:        profile.Name,
		ByteCount:               0,
		ByteCountFormatted:      "N/A",
		ByteCountHuman:          "0",
		DirectoryCount:          0,
		HasBagItProfile:         true,
		HasPackageOp:            false,
		HasUploadOps:            false,
		Name:                    "Validate Bags",
		PathSeparator:           string(os.PathSeparator),
		PayloadFileCount:        0,
		SourceFiles:             valJob.PathsToValidate,
	}
}
