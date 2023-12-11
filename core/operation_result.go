package core

import (
	"time"
)

type OperationResult struct {
	Attempt          int               `json:"attempt"`
	Completed        time.Time         `json:"completed"`
	Errors           map[string]string `json:"errors"`
	FileMTime        time.Time         `json:"fileMtime"`
	FilePath         string            `json:"filepath"`
	FileSize         int64             `json:"filesize"`
	Info             string            `json:"info"`
	Operation        string            `json:"operation"`
	Provider         string            `json:"provider"`
	RemoteTargetName string            `json:"remoteTargetName"`
	RemoteChecksum   string            `json:"remoteChecksum"`
	RemoteURL        string            `json:"remoteURL"`
	Started          time.Time         `json:"started"`
	Warning          string            `json:"warning"`
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
	return !r.Started.IsZero() && !r.Completed.IsZero() && !r.Completed.Before(r.Started)
}

func (r *OperationResult) Succeeded() bool {
	return r.WasAttempted() && r.WasCompleted() && !r.HasErrors()
}

// // UnmarshalJson will unmarshal both current and legacy (DART 2.x)
// // OperationResult structs into a current OperationResult object.
// func (r *OperationResult) UnmarshalJSON(data []byte) error {
// 	// In current struct, Errors is type map[string]string.
// 	var currentFormat struct {
// 		Attempt        int               `json:"attempt"`
// 		Completed      time.Time         `json:"completed"`
// 		Errors         map[string]string `json:"errors"`
// 		FileMTime      time.Time         `json:"fileMtime"`
// 		FilePath       string            `json:"filepath"`
// 		FileSize       int64             `json:"filesize"`
// 		Info           string            `json:"info"`
// 		Operation      string            `json:"operation"`
// 		Provider       string            `json:"provider"`
// 		RemoteChecksum string            `json:"remoteChecksum"`
// 		RemoteURL      string            `json:"remoteURL"`
// 		Started        time.Time         `json:"started"`
// 		Warning        string            `json:"warning"`
// 	}

// 	// In legacy struct, Errors is type []string.
// 	var legacyFormat struct {
// 		Attempt        int       `json:"attempt"`
// 		Completed      time.Time `json:"completed"`
// 		Errors         []string  `json:"errors"`
// 		FileMTime      time.Time `json:"fileMtime"`
// 		FilePath       string    `json:"filepath"`
// 		FileSize       int64     `json:"filesize"`
// 		Info           string    `json:"info"`
// 		Operation      string    `json:"operation"`
// 		Provider       string    `json:"provider"`
// 		RemoteChecksum string    `json:"remoteChecksum"`
// 		RemoteURL      string    `json:"remoteURL"`
// 		Started        time.Time `json:"started"`
// 		Warning        string    `json:"warning"`
// 	}

// 	// Try to unmarshal the struct in its current format.
// 	// If there's no error, copy values from currentStruct
// 	// into self and return.
// 	err := json.Unmarshal(data, &currentFormat)
// 	if err == nil {
// 		r.Attempt = currentFormat.Attempt
// 		r.Completed = currentFormat.Completed
// 		r.Errors = currentFormat.Errors
// 		r.FileMTime = currentFormat.FileMTime
// 		r.FilePath = currentFormat.FilePath
// 		r.FileSize = currentFormat.FileSize
// 		r.Info = currentFormat.Info
// 		r.Operation = currentFormat.Operation
// 		r.Provider = currentFormat.Provider
// 		r.RemoteChecksum = currentFormat.RemoteChecksum
// 		r.RemoteURL = currentFormat.RemoteURL
// 		r.Started = currentFormat.Started
// 		r.Warning = currentFormat.Warning
// 		return nil
// 	}

// 	// If we reach this point, the attempt to unmarshal into
// 	// type currentStruct failed. Let's try to unmarshal into
// 	// legacyStruct.
// 	err = json.Unmarshal(data, &legacyFormat)
// 	if err != nil {
// 		// Dang!
// 		return err
// 	}

// 	// OK, we were able to get a legacyStruct. Copy the values
// 	// from there into self, and note that we convert Errors
// 	// below from []string to map[string]string.
// 	r.Attempt = legacyFormat.Attempt
// 	r.Completed = legacyFormat.Completed
// 	r.FileMTime = legacyFormat.FileMTime
// 	r.FilePath = legacyFormat.FilePath
// 	r.FileSize = legacyFormat.FileSize
// 	r.Info = legacyFormat.Info
// 	r.Operation = legacyFormat.Operation
// 	r.Provider = legacyFormat.Provider
// 	r.RemoteChecksum = legacyFormat.RemoteChecksum
// 	r.RemoteURL = legacyFormat.RemoteURL
// 	r.Started = legacyFormat.Started
// 	r.Warning = legacyFormat.Warning
// 	r.Errors = make(map[string]string)
// 	for i, errMsg := range legacyFormat.Errors {
// 		key := fmt.Sprintf("Error #%d", i+1)
// 		r.Errors[key] = errMsg
// 	}
// 	return nil
// }
