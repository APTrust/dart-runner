package core_test

import (
	"encoding/json"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
)

func TestInfoEvent(t *testing.T) {
	message := "Hey diddly ho, Homer!"
	e := core.InfoEvent(message)
	assert.Equal(t, constants.EventTypeInfo, e.EventType)
	assert.Equal(t, message, e.Message)
	assert.Equal(t, `{"eventType":"info","message":"Hey diddly ho, Homer!"}`, e.ToJson())
}

func TestWarningEvent(t *testing.T) {
	message := "Rock stars! Is there anything they don't know?"
	e := core.WarningEvent(message)
	assert.Equal(t, constants.EventTypeWarning, e.EventType)
	assert.Equal(t, message, e.Message)
	assert.Equal(t, `{"eventType":"warning","message":"Rock stars! Is there anything they don't know?"}`, e.ToJson())
}

func TestErrorEvent(t *testing.T) {
	message := "The internet? Pfft! Is that thing still around?"
	e := core.ErrorEvent(message)
	assert.Equal(t, constants.EventTypeError, e.EventType)
	assert.Equal(t, message, e.Message)
	assert.Equal(t, `{"eventType":"error","message":"The internet? Pfft! Is that thing still around?"}`, e.ToJson())
}

func TestFinishEvent(t *testing.T) {
	job := getJobForJobResult()
	jobResult := core.NewJobResult(job)
	e := core.FinishEvent(jobResult)
	assert.Equal(t, constants.EventTypeFinish, e.EventType)
	assert.Equal(t, "Job succeeded", e.Message)
	assert.Equal(t, jobResult, e.JobResult)
	resultJson, _ := json.Marshal(jobResult)
	eventJson := e.ToJson()
	assert.Contains(t, eventJson, "finish")
	assert.Contains(t, eventJson, string(resultJson))
}
