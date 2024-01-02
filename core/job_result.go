package core

import (
	"encoding/json"
)

// JobResult collects the results of an attempted job for
// reporting purposes.
type JobResult struct {
	JobID             string             `json:"jobId"`
	JobName           string             `json:"jobName"`
	PayloadByteCount  int64              `json:"payloadByteCount"`
	PayloadFileCount  int64              `json:"payloadFileCount"`
	Succeeded         bool               `json:"succeeded"`
	PackageResult     *OperationResult   `json:"packageResult"`
	ValidationResults []*OperationResult `json:"validationResults"`
	UploadResults     []*OperationResult `json:"uploadResults"`
	ValidationErrors  map[string]string  `json:"validationErrors"`
}

// NewJobResult creates an object containing the results of all
// actions attempted in a job. The WorkflowRunner prints a json
// representation of this object to stdout upon completion or
// termination of each job.
func NewJobResult(job *Job) *JobResult {
	jobResult := &JobResult{
		JobID:             job.ID,
		JobName:           job.Name(),
		PayloadByteCount:  job.ByteCount,
		PayloadFileCount:  job.PayloadFileCount,
		Succeeded:         len(job.Errors) == 0,
		ValidationResults: make([]*OperationResult, 0),
		UploadResults:     make([]*OperationResult, 0),
	}
	if job.PackageOp != nil && job.PackageOp.Result != nil {
		jobResult.PackageResult = job.PackageOp.Result
		if !job.PackageOp.Result.Succeeded() {
			jobResult.Succeeded = false
		}
	}
	if job.ValidationOp != nil && job.ValidationOp.Result != nil {
		jobResult.ValidationResults = append(jobResult.ValidationResults, job.ValidationOp.Result)
		if !job.ValidationOp.Result.Succeeded() {
			jobResult.Succeeded = false
		}
	}
	if job.UploadOps != nil {
		for _, op := range job.UploadOps {
			jobResult.UploadResults = append(jobResult.UploadResults, op.Result)
			if !op.Result.Succeeded() {
				jobResult.Succeeded = false
			}
		}
	}
	// Do we want this if statement here or not?
	//if len(job.Errors) > 0 && !job.PackageAttempted() && !job.ValidationAttempted() && !job.UploadAttempted() {
	jobResult.ValidationErrors = job.Errors
	//}
	return jobResult
}

// NewJobResultFromValidationJob creates a new JobResult containing info
// about the outcome of a ValidationJob.
func NewJobResultFromValidationJob(valJob *ValidationJob) *JobResult {
	jobResult := &JobResult{
		JobID:             valJob.ID,
		JobName:           valJob.Name,
		Succeeded:         true,
		ValidationResults: make([]*OperationResult, len(valJob.ValidationOps)),
	}
	for i, op := range valJob.ValidationOps {
		jobResult.ValidationResults[i] = op.Result
		jobResult.ValidationResults[i].Info = "Bag is valid."
		if !op.Result.Succeeded() {
			jobResult.Succeeded = false
			jobResult.ValidationResults[i].Info = op.PathToBag
		}
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
