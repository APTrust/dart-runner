package core

import (
	"encoding/json"
	"fmt"

	"github.com/APTrust/dart-runner/constants"
)

// EventMessage contains info to send back to the front end UI
// about the progress of a job, so the user can see that things
// are moving along. We'll queue info events when we add files
// to a package, validate checksums, etc.
//
// The JobResult property will be set only for the finish event.
// JobResult contains final information about the success or
// failure of the job.
type EventMessage struct {
	// EventType is the type of event.
	EventType string `json:"eventType"`
	// Stage describes which stage of the job this
	// event pertains to (packaging, validation, upload).
	Stage string `json:"stage"`
	// Status describes the status of this stage
	// (running, success, failure).
	Status string `json:"status"`
	// Message is a human-friendly message to display to
	// the user.
	Message string `json:"message"`
	// Total is the total number of files to package or bytes
	// to upload. This is used to calculate percent complete for
	// progress bars on the front end.
	Total int64 `json:"total"`
	// Current is the current number of files packaged or bytes
	// uploaded. This is used to calculate percent complete for
	// progress bars on the front end.
	Current int64 `json:"current"`
	// Percent is the percent complete of the current packaging,
	// validation or upload operation. This is used to control
	// progress bars on the front end.
	Percent int `json:"percent"`
	// JobResult describes the final outcome of a job and all
	// its component operations (packaging, validation, upload).
	// This object will only be present in the finish event.
	// For all other events, it's null.
	JobResult *JobResult `json:"jobResult,omitempty"`
	// JobInitSettings contains information to help the front
	// end set up the job progress display. This object is present
	// only in the init event. For all other events, it's null.
	JobInitSettings *JobInitSettings `json:"jobInitSettings,omitempty"`
}

// JobInitSettings contains information for the front-end
// HTML/JavaScript page to help display the status of a
// running job.
type JobInitSettings struct {
	// RunningJobId is the ID of the currently running job.
	// This will change on the workflow batch page as we
	// run a series of jobs.
	RunningJobId string `json:"runningJobId"`
	// HasPackageOp describes whether this job includes a
	// package operation so the front end knows whether to
	// display the packaging status info and progress bar.
	HasPackageOp bool `json:"hasPackageOp"`
	// HasUploadOp describes whether this job includes any
	// upload operations so the front end knows whether to
	// display the upload status info and progress bars.
	HasUploadOp bool `json:"hasUploadOp"`
	// PathSeparator is the operating system's path separator,
	// which is a backslash on Windows and forward slash on
	// other operating systems.
	PathSeparator string `json:"pathSeparator"`
	// PackageFormat will be either BagIt or none, depending
	// on if/how this job packages files.
	PackageFormat string `json:"packageFormat"`
	// EventSourceUrl is the URL that the front end should
	// attach to to receive server-sent events. For individual
	// jobs on the run-job page, this will "/jobs/run/:id". For
	// workflow batches, it will be "/workflows/runbatch"
	EventSourceUrl string `json:"eventSourceUrl"`
}

// InitEvent creates a new initialization event message. This message
// contains info that the front end needs to set up the job progress
// display.
func InitEvent(jobInitSettings *JobInitSettings) *EventMessage {
	message := fmt.Sprintf("Starting job %s", jobInitSettings.RunningJobId)
	return &EventMessage{
		Stage:           constants.StagePreRun,
		EventType:       constants.EventTypeInit,
		Message:         message,
		Status:          constants.StatusStarting,
		JobInitSettings: jobInitSettings,
	}
}

// StartEvent creates a new EventMessage with EventType start.
// This event has no JobResult.
func StartEvent(stage, message string) *EventMessage {
	return &EventMessage{
		Stage:     stage,
		EventType: constants.EventTypeStart,
		Message:   message,
		Status:    constants.StatusRunning,
	}
}

// InfoEvent creates a new EventMessage with EventType info.
// This event has no JobResult.
func InfoEvent(stage, message string) *EventMessage {
	return &EventMessage{
		Stage:     stage,
		EventType: constants.EventTypeInfo,
		Message:   message,
		Status:    constants.StatusRunning,
	}
}

// WarningEvent creates a new EventMessage with EventType warning.
// This event has no JobResult.
func WarningEvent(stage, message string) *EventMessage {
	return &EventMessage{
		Stage:     stage,
		EventType: constants.EventTypeWarning,
		Message:   message,
		Status:    constants.StatusRunning,
	}
}

// ErrorEvent creates a new EventMessage with EventType error.
// This event has no JobResult.
func ErrorEvent(stage, message string) *EventMessage {
	return &EventMessage{
		Stage:     stage,
		EventType: constants.EventTypeError,
		Message:   message,
		Status:    constants.StatusRunning, // caller can change to StatusFailed if this is fatal
	}
}

// Finish event creates an EventMessage with a JobResult
// describing how the job turned out. This sets the EventType
// to "info" if the job succeeded or to "error" if it failed.
func FinishEvent(jobResult *JobResult) *EventMessage {
	message := "Job failed"
	status := constants.StatusFailed
	if jobResult.Succeeded {
		message = "Job succeeded"
		status = constants.StatusSuccess
	}
	return &EventMessage{
		EventType: constants.EventTypeFinish,
		Stage:     constants.StageFinish,
		Message:   message,
		JobResult: jobResult,
		Status:    status,
	}
}

// ToJson converts this EventMessage to JSON, so we can send it
// back to the UI for display.
func (e *EventMessage) ToJson() string {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf("Error serializing job result: %v", err)
	}
	return string(data)
}
