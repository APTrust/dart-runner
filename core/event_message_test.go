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
	e := core.InfoEvent(constants.StagePackage, message)
	assert.Equal(t, constants.StagePackage, e.Stage)
	assert.Equal(t, constants.EventTypeInfo, e.EventType)
	assert.Equal(t, message, e.Message)
	assert.Equal(t, `{"eventType":"info","stage":"package","status":"running","message":"Hey diddly ho, Homer!","total":0,"current":0,"percent":0}`, e.ToJson())
}

func TestWarningEvent(t *testing.T) {
	message := "Rock stars! Is there anything they don't know?"
	e := core.WarningEvent(constants.StagePreRun, message)
	assert.Equal(t, constants.StagePreRun, e.Stage)
	assert.Equal(t, constants.EventTypeWarning, e.EventType)
	assert.Equal(t, message, e.Message)
	assert.Equal(t, `{"eventType":"warning","stage":"pre-run","status":"running","message":"Rock stars! Is there anything they don't know?","total":0,"current":0,"percent":0}`, e.ToJson())
}

func TestErrorEvent(t *testing.T) {
	message := "The internet? Pfft! Is that thing still around?"
	e := core.ErrorEvent(constants.StageValidate, message)
	assert.Equal(t, constants.StageValidate, e.Stage)
	assert.Equal(t, constants.EventTypeError, e.EventType)
	assert.Equal(t, message, e.Message)
	assert.Equal(t, `{"eventType":"error","stage":"validate","status":"running","message":"The internet? Pfft! Is that thing still around?","total":0,"current":0,"percent":0}`, e.ToJson())
}

func TestFinishEvent(t *testing.T) {
	job := getJobForJobResult()
	jobResult := core.NewJobResult(job)
	e := core.FinishEvent(jobResult)
	assert.Equal(t, constants.EventTypeFinish, e.EventType)
	assert.Equal(t, constants.StageFinish, e.Stage)
	assert.Equal(t, "Job succeeded", e.Message)
	assert.Equal(t, jobResult, e.JobResult)
	resultJson, _ := json.Marshal(jobResult)
	eventJson := e.ToJson()
	assert.Contains(t, eventJson, "finish")
	assert.Contains(t, eventJson, string(resultJson))
}