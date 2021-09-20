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
