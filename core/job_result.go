package core

import (
	"encoding/json"
)

// JobResult collects the results of an attempted job for
// reporting purposes.
type JobResult struct {
	JobName          string             `json:"jobName"`
	PayloadByteCount int64              `json:"payloadByteCount"`
	PayloadFileCount int64              `json:"payloadFileCount"`
	Succeeded        bool               `json:"succeeded"`
	Results          []*OperationResult `json:"operationResults"`
	ValidationErrors map[string]string  `json:"validationErrors"`
}

// NewJobResult creates an object containing the results of all
// actions attempted in a job. The WorkflowRunner prints a json
// representation of this object to stdout upon completion or
// termination of each job.
func NewJobResult(job *Job) *JobResult {
	results := make([]*OperationResult, 0)
	if job.PackageOp != nil && job.PackageOp.Result != nil {
		results = append(results, job.PackageOp.Result)
	}
	if job.ValidationOp != nil && job.ValidationOp.Result != nil {
		results = append(results, job.ValidationOp.Result)
	}
	if job.UploadOps != nil {
		for _, op := range job.UploadOps {
			results = append(results, op.Result)
		}
	}
	validationErrors := make(map[string]string)
	if len(job.Errors) > 0 && !job.PackageAttempted() && !job.ValidationAttempted() && !job.UploadAttempted() {
		validationErrors = job.Errors
	}
	return &JobResult{
		JobName:          job.Name(),
		PayloadByteCount: job.ByteCount,
		PayloadFileCount: job.FileCount,
		Succeeded:        len(job.Errors) == 0,
		ValidationErrors: validationErrors,
		Results:          results,
	}
}

// ToJson returns a pretty-printed JSON string describing the
// results of this job's operations.
func (r *JobResult) ToJson() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", nil
	}
	return string(data), nil
}
