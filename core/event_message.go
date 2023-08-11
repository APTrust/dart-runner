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
	EventType string     `json:"eventType"`
	Stage     string     `json:"stage"`
	Status    string     `json:"status"`
	Message   string     `json:"message"`
	Total     int64      `json:"total"`
	Current   int64      `json:"current"`
	Percent   int        `json:"percent"`
	JobResult *JobResult `json:"jobResult,omitempty"`
}

// InfoMessage creates a new EventMessage with EventType info.
// This event has no JobResult.
func InfoEvent(stage, message string) *EventMessage {
	return &EventMessage{
		Stage:     stage,
		EventType: constants.EventTypeInfo,
		Message:   message,
		Status:    constants.StatusRunning,
	}
}

// WarningMessage creates a new EventMessage with EventType warning.
// This event has no JobResult.
func WarningEvent(stage, message string) *EventMessage {
	return &EventMessage{
		Stage:     stage,
		EventType: constants.EventTypeWarning,
		Message:   message,
		Status:    constants.StatusRunning,
	}
}

// ErrorMessage creates a new EventMessage with EventType error.
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
