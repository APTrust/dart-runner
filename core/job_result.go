package core

import (
	"encoding/json"
)

// TODO: Change Results to PackageResult, ValidationResult, UploadResults
//       Fill in FileSize, FilePath on package & validation results
//       Provider should indicate this was a DART-runner job

// JobResult collects the results of an attempted job for
// reporting purposes.
type JobResult struct {
	JobName          string             `json:"jobName"`
	PayloadByteCount int64              `json:"payloadByteCount"`
	PayloadFileCount int64              `json:"payloadFileCount"`
	Succeeded        bool               `json:"succeeded"`
	PackageResult    *OperationResult   `json:"packageResult"`
	ValidationResult *OperationResult   `json:"validationResult"`
	UploadResults    []*OperationResult `json:"uploadResults"`
	ValidationErrors map[string]string  `json:"validationErrors"`
}

// NewJobResult creates an object containing the results of all
// actions attempted in a job. The WorkflowRunner prints a json
// representation of this object to stdout upon completion or
// termination of each job.
func NewJobResult(job *Job) *JobResult {
	jobResult := &JobResult{
		JobName:          job.Name(),
		PayloadByteCount: job.ByteCount,
		PayloadFileCount: job.FileCount,
		Succeeded:        len(job.Errors) == 0,
		UploadResults:    make([]*OperationResult, 0),
	}
	if job.PackageOp != nil && job.PackageOp.Result != nil {
		jobResult.PackageResult = job.PackageOp.Result
	}
	if job.ValidationOp != nil && job.ValidationOp.Result != nil {
		jobResult.ValidationResult = job.ValidationOp.Result
	}
	if job.UploadOps != nil {
		for _, op := range job.UploadOps {
			jobResult.UploadResults = append(jobResult.UploadResults, op.Result)
		}
	}
	if len(job.Errors) > 0 && !job.PackageAttempted() && !job.ValidationAttempted() && !job.UploadAttempted() {
		jobResult.ValidationErrors = job.Errors
	}
	return jobResult
}

// ToJson returns a JSON string describing the results of this
// job's operations.
func (r *JobResult) ToJson() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", nil
	}
	return string(data), nil
}
