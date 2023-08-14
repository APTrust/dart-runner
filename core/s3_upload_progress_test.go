package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
)

func TestS3UploadProgress(t *testing.T) {
	messageChannel := make(chan *core.EventMessage)
	defer close(messageChannel)

	up := core.NewS3UploadProgress(200, messageChannel)
	go func() {
		up.Read([]byte("ten bytes!"))
	}()

	msg, ok := <-messageChannel
	assert.True(t, ok)
	assert.Equal(t, constants.EventTypeInfo, msg.EventType)
	assert.Equal(t, constants.StageUpload, msg.Stage)
	assert.Equal(t, int64(200), msg.Total)
	assert.Equal(t, int64(10), msg.Current)
	assert.Equal(t, 5, msg.Percent)
	assert.Equal(t, "Sent 10 B of 200 B (5%)", msg.Message)
}
