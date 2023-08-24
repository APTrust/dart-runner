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

// TODO: Refactor this. Message channel should be private.
// Name should change to StreamProgress.
func NewS3UploadProgress(bytesToSend int64, messageChannel chan *EventMessage) *S3UploadProgress {
	return &S3UploadProgress{
		Total:          bytesToSend,
		MessageChannel: messageChannel,
	}
}

// Read satisfies the progress interface for the Minio client,
// which passes each chunk of bytes through this function as it
// uploads.
func (p *S3UploadProgress) Read(b []byte) (int, error) {

	byteCount := int64(len(b))
	return p.SetTotalBytesCompleted(p.Current + byteCount)

	// byteCount := int64(len(b))
	// p.Current += byteCount
	// p.Percent = int(float64(p.Current) * 100 / float64(p.Total))

	// total := util.ToHumanSize(p.Total, 1024)
	// sent := util.ToHumanSize(p.Current, 1024)

	// message := fmt.Sprintf("Sent %s of %s (%d%%)", sent, total, p.Percent)
	// eventMessage := InfoEvent(constants.StageUpload, message)
	// eventMessage.Total = p.Total
	// eventMessage.Current = p.Current
	// eventMessage.Percent = p.Percent
	// p.MessageChannel <- eventMessage
	// return int(byteCount), nil
}

// SetTotalBytesCompleted satisfies the progress interface for SFTP
// uploads, where the progress meter periodically tells us the number
// of total bytes uploaded thus far.
func (p *S3UploadProgress) SetTotalBytesCompleted(byteCount int64) (int, error) {
	p.Current += byteCount
	p.Percent = int(float64(p.Current) * 100 / float64(p.Total))

	total := util.ToHumanSize(p.Total, 1024)
	sent := util.ToHumanSize(p.Current, 1024)

	message := fmt.Sprintf("Sent %s of %s (%d%%)", sent, total, p.Percent)
	eventMessage := InfoEvent(constants.StageUpload, message)
	eventMessage.Total = p.Total
	eventMessage.Current = p.Current
	eventMessage.Percent = p.Percent
	p.MessageChannel <- eventMessage
	return int(byteCount), nil
}
