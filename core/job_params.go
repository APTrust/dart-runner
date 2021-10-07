package core

import (
	"strings"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

// JobParams describes what we need to assemble to create a Job
// for the JobRunner. This structure allows to create a job object
// by describing what needs to be done (the workflow), what set of
// files we're operating on (the files) and where we'll store a
// local copy of the result (the output path).
type JobParams struct {
	Errors      map[string]string `json:"errors"`
	Files       []string          `json:"files"`
	PackageName string            `json:"packageName"`
	OutputPath  string            `json:"outputPath"`
	Tags        []*bagit.Tag      `json:"tags"`
	Workflow    *Workflow         `json:"workflow"`
}

// NewJobParams creates a new JobParams object.
//
// Param workflow is the workflow to run. packageName is the name of
// the output file (for now, a tarred bag, so something like "my_bag.tar").
// outputPath is the directory into which we should write the bag.
// files is a list of files to bag.
func NewJobParams(workflow *Workflow, packageName, outputPath string, files []string, tags []*bagit.Tag) *JobParams {
	return &JobParams{
		Errors:      make(map[string]string),
		Files:       files,
		OutputPath:  outputPath,
		PackageName: packageName,
		Tags:        tags,
		Workflow:    workflow,
	}
}

// ToJob converts a JobParams object to a Job object, which can be run
// directly by the JobRunner.
func (p *JobParams) ToJob() *Job {
	job := NewJob()
	job.BagItProfile = bagit.CloneProfile(p.Workflow.BagItProfile)
	job.WorkflowID = p.Workflow.ID
	p.makePackageOp(job)
	p.makeValidationOp(job)
	p.makeUploadOps(job)
	p.mergeTags(job)
	return job
}

// mergeTags merges tag values specific to this job into the list
// of profile tag values. The profile defines required tags and allowed
// values. The values we actually assign to those tags come from the
// WorkFlow's CSV file. For example, if Internal-Sender-Identifier is
// required, every line in the CSV file should have an entry for
// bag-info.txt/Internal-Sender-Identifier. We'll copy that value into
// the profile so we can write it into the bag.
//
// BTW, this is why we clone the profile. We don't want tag values from
// one bag leaking into the next.
//
// The CSV file may include tags not defined in the profile, and it may
// also include multiple copies of tags that are defined. (For example,
// multiple instances of bag-info.txt/Internal-Sender-Identifier.)
// In either case, we add the new/additional tags from the CSV into
// the profile so that our bagger will write them into the appropriate
// tag files.
//
// If the profile requires tags that are missing from the CSV file,
// the bagger will complain and quit.
func (p *JobParams) mergeTags(job *Job) {
	if p.Workflow.BagItProfile == nil {
		return
	}
	profile := job.BagItProfile
	for _, t := range p.Tags {
		profileTagDef := profile.GetTagDef(t.TagFile, t.TagName)
		if profileTagDef == nil {
			profileTagDef = &bagit.TagDefinition{
				TagFile: t.TagFile,
				TagName: t.TagName,
			}
			profile.Tags = append(profile.Tags, profileTagDef)
		}
		profileTagDef.UserValue = t.Value
	}
}

// makePackageOp creates the package operation for this job.
// For now, we support only BagIt package format and those must be
// serialized as tarballs. So, if we have a package operation, it's
// going to produce a tarred bag.
//
// It's possible to not have a package operation at all. This would
// be the case if you're only validating a bag, or just copying files
// to S3/SFTP.
func (p *JobParams) makePackageOp(job *Job) {
	if p.PackageName != "" {
		job.PackageOp = NewPackageOperation(p.PackageName, p.OutputPath, p.Files)
		job.PackageOp.PackageFormat = p.Workflow.PackageFormat
		p.setSerialization(job)
	}
}

// makeValidationOp creates a ValidationOperation if the job is packaging
// something and includes a BagIt profile.
func (p *JobParams) makeValidationOp(job *Job) {
	if p.PackageName != "" && p.Workflow.BagItProfile != nil {
		job.ValidationOp = NewValidationOperation(p.OutputPath)
	}
}

// makeUploadOps creates the upload operations for this job. Jobs may
// upload files to 0..N targets. A common case is to create a bag and
// send it off to an S3 bucket in AWS and a second bucket in Wasabi.
func (p *JobParams) makeUploadOps(job *Job) {
	if len(p.Workflow.StorageServices) == 0 {
		// No storage services specified, so no uploads to perform.
		return
	}
	var files []string
	if job.PackageOp != nil && job.PackageOp.OutputPath != "" {
		// We want to upload the result of the package operation.
		files = []string{job.PackageOp.OutputPath}
	} else {
		// No packaging step. We want to upload the files themselves.
		files = p.Files
	}
	for _, ss := range p.Workflow.StorageServices {
		job.UploadOps = append(job.UploadOps, NewUploadOperation(ss, files))
	}
}

// setSeriaization sets the serialization format for the bag
// that we'll produce in the package operation. Since we only
// support tar at the moment, this is always going to set the
// format to .tar.
func (p *JobParams) setSerialization(job *Job) {
	// We can't set this if there's no package operation,
	// as in upload-only or validation-only jobs.
	// We also can't set it if there's no bagit profile.
	if job.PackageOp == nil || job.BagItProfile == nil {
		return
	}
	profile := job.BagItProfile
	formats := profile.AcceptSerialization
	serializationOK := (profile.Serialization == constants.SerializationRequired || profile.Serialization == constants.SerializationOptional)
	supportsTar := util.StringListContains(formats, "application/tar") ||
		util.StringListContains(formats, "application/x-tar")
	if serializationOK && supportsTar {
		job.PackageOp.BagItSerialization = ".tar"
		if !strings.HasSuffix(job.PackageOp.OutputPath, ".tar") {
			job.PackageOp.OutputPath += ".tar"
		}
	}
}
