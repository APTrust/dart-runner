package core

import (
	"fmt"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

type S3UploadProgress struct {
	Total          int64
	Current        int64
	Percent        int
	MessageChannel chan *EventMessage
}

func NewS3UploadProgress(bytesToSend int64, messageChannel chan *EventMessage) *S3UploadProgress {
	return &S3UploadProgress{
		Total:          bytesToSend,
		MessageChannel: messageChannel,
	}
}

func (p *S3UploadProgress) Read(b []byte) (int, error) {
	byteCount := int64(len(b))
	p.Current += byteCount
	p.Percent = int(float64(p.Current) * 100 / float64(p.Total))

	total := util.ToHumanSize(p.Total, 1024)
	sent := util.ToHumanSize(p.Current, 1024)

	message := fmt.Sprintf("Sent %s of %s (%d%%)", sent, total, p.Percent)
	p.MessageChannel <- InfoEvent(constants.StageUpload, message)
	return int(byteCount), nil
}
