package core

import (
	"path"

	"github.com/APTrust/dart-runner/bagit"
)

type JobParams struct {
	Errors     map[string]string `json:"errors"`
	Files      []string          `json:"files"`
	OutputPath string            `json:"packageName"`
	Tags       []*bagit.Tag      `json:"tags"`
	Workflow   *Workflow         `json:"workflow"`
}

func NewJobParams(workflow *Workflow, outputPath string, files []string, tags []*bagit.Tag) *JobParams {
	return &JobParams{
		Errors:     make(map[string]string),
		Files:      files,
		OutputPath: outputPath,
		Tags:       tags,
		Workflow:   workflow,
	}
}

func (p *JobParams) ToJob() *Job {
	job := NewJob()
	job.BagItProfile = bagit.CloneProfile(p.Workflow.BagItProfile)
	job.WorkflowID = p.Workflow.ID
	p.makePackageOp(job)
	p.makeUploadOps(job)
	p.mergeTags(job)
	return job
}

func (p *JobParams) mergeTags(job *Job) {

}

func (p *JobParams) makePackageOp(job *Job) {

}

func (p *JobParams) makeUploadOps(job *Job) {

}

func (p *JobParams) PackageName() string {
	return path.Base(p.OutputPath)
}

func (p *JobParams) setSerialization() {

}
