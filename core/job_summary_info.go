package core

import (
	"os"

	"github.com/APTrust/dart-runner/util"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// JobSummaryInfo contains info about a job to be displayed
// in the front end. All of this info appears on the job_run
// page and the batch_run page. It may appear elsewhere as
// well. Note that it takes some logic to extract some of this
// information. We don't want to force that logic into HTML
// templates and front-end JavaScript, so we do it here, where
// it's easy to test.
type JobSummaryInfo struct {
	HasPackageOp            bool
	HasUploadOps            bool
	HasBagItProfile         bool
	PackageFormat           string
	PackageName             string
	BagItProfileName        string
	BagItProfileDescription string
	DirectoryCount          int64
	PayloadFileCount        int64
	ByteCount               int64
	ByteCountFormatted      string
	ByteCountHuman          string
	SourceFiles             []string
	OutputPath              string
	UploadTargets           []string
	PathSeparator           string
}

// NewJobSummaryInfo creates a new JobSummaryInfo object based on
// the given job.
func NewJobSummaryInfo(job *Job) *JobSummaryInfo {
	info := &JobSummaryInfo{
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
