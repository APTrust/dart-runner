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
	Message   string     `json:"message"`
	JobResult *JobResult `json:"jobResult,omitempty"`
}

// InfoMessage creates a new EventMessage with EventType info.
// This event has no JobResult.
func InfoEvent(message string) *EventMessage {
	return &EventMessage{
		EventType: constants.EventTypeInfo,
		Message:   message,
	}
}

// WarningMessage creates a new EventMessage with EventType warning.
// This event has no JobResult.
func WarningEvent(message string) *EventMessage {
	return &EventMessage{
		EventType: constants.EventTypeWarning,
		Message:   message,
	}
}

// ErrorMessage creates a new EventMessage with EventType error.
// This event has no JobResult.
func ErrorEvent(message string) *EventMessage {
	return &EventMessage{
		EventType: constants.EventTypeError,
		Message:   message,
	}
}

// Finish event creates an EventMessage with a JobResult
// describing how the job turned out. This sets the EventType
// to "info" if the job succeeded or to "error" if it failed.
func FinishEvent(jobResult *JobResult) *EventMessage {
	message := "Job failed"
	if jobResult.Succeeded {
		message = "Job succeeded"
	}
	return &EventMessage{
		EventType: constants.EventTypeFinish,
		Message:   message,
		JobResult: jobResult,
	}
}

// ToJson converts this EventMessage to JSON, so we can send it
// back to the UI for display.
func (e *EventMessage) ToJson() string {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf("Error serializin job result: %v", err)
	}
	return string(data)
}
