package core

import (
	"fmt"
	"sync"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

// StreamProgress sends info from an upload/download stream into
// a message channel so we can report upload/download progress to
// the user. When running in GUI mode, the S3 and SFTP clients send
// EventMessages to the front-end via server-sent events, and the
// front end updates a progress bar.
type StreamProgress struct {
	Total          int64
	Current        int64
	Percent        int
	MessageChannel chan *EventMessage
	mutex          sync.RWMutex
}

// NewStreamProgress returns a new StreamProgress object. Param byteCount
// is the number of bytes we need to send or receive to complete
// the upload/download/copy operation.
func NewStreamProgress(byteCount int64, messageChannel chan *EventMessage) *StreamProgress {
	return &StreamProgress{
		Total:          byteCount,
		MessageChannel: messageChannel,
	}
}

// Read satisfies the progress interface for the Minio client,
// which passes progress information as it uploads. Each time
// Minio calls Read(), the length of b equals the total number
// of bytes read thus far.
//
// See https://github.com/search?q=repo%3Aminio%2Fminio-go%20progress&type=code
func (p *StreamProgress) Read(b []byte) (int, error) {
	bytesTransferredSoFar := int64(len(b))
	return p.SetTotalBytesCompleted(bytesTransferredSoFar)
}

// SetTotalBytesCompleted satisfies the progress interface for SFTP
// uploads, where the progress meter periodically tells us the number
// of total bytes uploaded thus far.
func (p *StreamProgress) SetTotalBytesCompleted(byteCount int64) (int, error) {
	p.mutex.Lock()
	p.Current += byteCount
	p.Percent = int(float64(p.Current) * 100 / float64(p.Total))
	total := util.ToHumanSize(p.Total, 1024)
	sent := util.ToHumanSize(p.Current, 1024)
	message := fmt.Sprintf("Sent %s of %s (%d%%)", sent, total, p.Percent)
	eventMessage := InfoEvent(constants.StageUpload, message)
	eventMessage.Total = p.Total
	eventMessage.Current = p.Current
	eventMessage.Percent = p.Percent
	p.mutex.Unlock()

	p.MessageChannel <- eventMessage
	return int(byteCount), nil
}
