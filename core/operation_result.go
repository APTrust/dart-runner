package core

import (
	"time"
)

type OperationResult struct {
	Attempt        int               `json:"attempt"`
	Completed      time.Time         `json:"completed"`
	Errors         map[string]string `json:"errors"`
	FileMTime      time.Time         `json:"fileMtime"`
	FilePath       string            `json:"filepath"`
	FileSize       int64             `json:"filesize"`
	Info           string            `json:"info"`
	Operation      string            `json:"operation"`
	Provider       string            `json:"provider"`
	RemoteChecksum string            `json:"remoteChecksum"`
	RemoteURL      string            `json:"remoteURL"`
	Started        time.Time         `json:"started"`
	Warning        string            `json:"warning"`
}

func NewOperationResult(operation, provider string) *OperationResult {
	return &OperationResult{
		Errors:    make(map[string]string),
		Operation: operation,
		Provider:  provider,
	}
}

func (r *OperationResult) Reset() {
	r.Started = time.Time{}
	r.Completed = time.Time{}
	r.FileSize = 0
	r.FileMTime = time.Time{}
	r.RemoteChecksum = ""
	r.RemoteURL = ""
	r.Info = ""
	r.Warning = ""
	r.Errors = make(map[string]string)
}

func (r *OperationResult) Start() {
	r.Reset()
	r.Started = time.Now()
	r.Attempt += 1
}

func (r *OperationResult) Finish(errors map[string]string) {
	r.Completed = time.Now()
	for key, value := range errors {
		r.Errors[key] = value
	}
}

func (r *OperationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

func (r *OperationResult) WasAttempted() bool {
	return r.Attempt > 0
}

func (r *OperationResult) WasCompleted() bool {
	return !r.Started.IsZero() && !r.Completed.IsZero() && r.Completed.After(r.Started)
}

func (r *OperationResult) Succeeded() bool {
	return r.WasAttempted() && r.WasCompleted() && !r.HasErrors()
}
