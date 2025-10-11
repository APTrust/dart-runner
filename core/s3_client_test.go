package core_test

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS3FileUpload(t *testing.T) {
	ss := getS3StorageService()
	s3Client, err := core.NewS3Client(ss, false, nil)
	require.Nil(t, err)
	filename := fileToUpload()
	key := filepath.Base(filename)
	err = s3Client.Upload(fileToUpload(), key)
	require.Nil(t, err)
}

func TestS3FileUploadWithProgress(t *testing.T) {
	fileToUpload := fileToUpload()
	payloadSize, err := core.GetUploadPayloadSize(fileToUpload)
	require.Nil(t, err)
	testSFTPUploadWithProgressBar(t, "file", payloadSize)
}

func TestS3DirectoryUpload(t *testing.T) {
	ss := getS3StorageService()
	s3Client, err := core.NewS3Client(ss, false, nil)
	require.Nil(t, err)
	dirname := dirToUpload()
	key := filepath.Base(dirname)
	err = s3Client.Upload(dirToUpload(), key)
	require.Nil(t, err)
}

func TestS3DirectoryUploadWithProgress(t *testing.T) {
	dirToUpload := dirToUpload()
	payloadSize, err := core.GetUploadPayloadSize(dirToUpload)
	require.Nil(t, err)
	testSFTPUploadWithProgressBar(t, "directory", payloadSize)
}

func getS3StorageService() *core.StorageService {
	return &core.StorageService{
		ID:           uuid.NewString(),
		Name:         "Local Minio S3 service",
		AllowsUpload: true,
		Bucket:       "aptrust.receiving.test.test.edu",
		Host:         "127.0.0.1",
		Login:        "minioadmin",
		Password:     "minioadmin",
		Port:         9899,
		Protocol:     constants.ProtocolS3,
	}
}

func testS3UploadFile(t *testing.T, s3Client *core.S3Client) {
	destPath := filepath.Base(fileToUpload())
	err := s3Client.Upload(fileToUpload(), destPath)
	require.Nil(t, err)
}

func testS3UploadDir(t *testing.T, s3Client *core.S3Client) {
	destPath := filepath.Base(dirToUpload())
	err := s3Client.Upload(dirToUpload(), destPath)
	require.Nil(t, err)
}

// This is a tricky test. We want to make sure the SFTP uploader
// pushes messages into the progress reader's message channel.
// See comments inline...
func testS3UploadWithProgressBar(t *testing.T, fileOrDir string, payloadSize int64) {
	iReadSomething := false
	messageChannel := make(chan *core.EventMessage)

	// When this function exits, make sure our reader read a message
	// from the messageChannel. The SFTP uploader is supposed to pump
	// messages into this channel with info about the upload progress.
	defer func() {
		close(messageChannel)
		assert.True(t, iReadSomething)
	}()

	// We have to set up a WaitGroup to prevent a data race.
	// The go routine below is fiddling with the iReadSomething
	// flag defined above. Unless we wait, our test function could
	// exit before the go routine below gets to do its work.
	// That can lead to two problems:
	// 1. iReadSomething never gets set to true.
	// 2. The defer function will run too soon, closing the message
	//    channel, and then the go routine will try to read from a
	//    closed channel.
	var wg sync.WaitGroup

	// Create a client with progress reader.
	ss := getS3StorageService()
	s3Client, err := core.NewS3Client(ss, false, messageChannel)
	require.Nil(t, err)
	require.NotNil(t, s3Client)

	// How many messages will the SFTP uploader send into our
	// messageChannel? If we're uploading a single file, it will
	// be just one. For a directory, we have to calculate.
	// We do this so our goroutine knows when to call wg.Done()
	expectedMessageCount := 1
	if fileOrDir == "directory" {
		expectedMessageCount, err = fileCount(dirToUpload())
		require.Nil(t, err)
	}
	go func() {
		messageCount := 0
		for msg := range messageChannel {
			iReadSomething = true
			assert.True(t, strings.HasPrefix(msg.Message, "Sent"))
			messageCount += 1
			if messageCount == expectedMessageCount {
				wg.Done()
			}
		}
	}()

	// Tell our WaitGroup we have a task pending.
	// Then we'll kick off the upload and wait for WaitGroup
	// to complete. When the go routine above cathches a message
	// from core.SFTPUpload, it will call wg.Done(). Then the
	// wg.Wait() at the end of this function will stop waiting.
	//
	// At that point, our defer func above will fire, closing
	// the message channel and testing to see whether the go
	// routine actually read a message.
	wg.Add(1)

	if fileOrDir == "file" {
		testS3UploadFile(t, s3Client)
	} else {
		testS3UploadDir(t, s3Client)
	}

	wg.Wait()
}
