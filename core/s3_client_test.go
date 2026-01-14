package core_test

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: S3 tests require a local Minio server to be running.
// See core/sftp_client_test.go for instructions on starting
// the Minio server using scripts/test.rb or scripts/run.rb.

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

	assert.Equal(t, int64(1), s3Client.FilesUploaded())
	assert.Equal(t, int64(23552), s3Client.BytesUploaded())
	assert.Equal(t, 1, len(s3Client.EtagMap()))
}

func testS3UploadDir(t *testing.T, s3Client *core.S3Client) {
	destPath := filepath.Base(dirToUpload())
	err := s3Client.Upload(dirToUpload(), destPath)
	require.Nil(t, err)

	fileCount, err := fileCount(dirToUpload())
	require.Nil(t, err)
	assert.Equal(t, int64(fileCount), s3Client.FilesUploaded())
	assert.True(t, s3Client.BytesUploaded() > int64(25000))
	assert.Equal(t, fileCount, len(s3Client.EtagMap()))
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

func TestListBuckets(t *testing.T) {
	ss := getS3StorageService()
	s3Client, err := core.NewS3Client(ss, false, nil)
	require.Nil(t, err)
	require.NotNil(t, s3Client)

	buckets, err := s3Client.ListBuckets()
	require.Nil(t, err)
	require.NotNil(t, buckets)

	// The local Minio instance should have at least one bucket
	// (the test bucket defined in getS3StorageService)
	assert.True(t, len(buckets) > 0, "Expected at least one bucket")

	// Check that we can find our test bucket
	foundTestBucket := false
	for _, bucket := range buckets {
		if bucket.Name == ss.Bucket {
			foundTestBucket = true
			break
		}
	}
	assert.True(t, foundTestBucket, "Expected to find test bucket: %s", ss.Bucket)
}

func TestListObjects(t *testing.T) {
	ss := getS3StorageService()
	s3Client, err := core.NewS3Client(ss, false, nil)
	require.Nil(t, err)
	require.NotNil(t, s3Client)

	// First, upload a test file so we have something to list
	filename := fileToUpload()
	key := filepath.Base(filename)
	err = s3Client.Upload(filename, key)
	require.Nil(t, err)

	// Now list objects in the bucket
	opts := minio.ListObjectsOptions{
		Recursive: true,
	}
	objects := s3Client.ListObjects(ss.Bucket, "", opts)
	require.NotNil(t, objects)

	// We should have at least one object (the one we just uploaded)
	assert.True(t, len(objects) > 0, "Expected at least one object in bucket")

	// Find our uploaded file
	foundUploadedFile := false
	for _, obj := range objects {
		if obj.Key == key {
			foundUploadedFile = true
			assert.Equal(t, key, obj.Key)
			assert.True(t, obj.Size > 0, "Expected object size to be greater than 0")
			assert.NotEmpty(t, obj.ETag)
			assert.NotEmpty(t, obj.LastModified)
			break
		}
	}
	assert.True(t, foundUploadedFile, "Expected to find uploaded file: %s", key)
}

func TestListObjectsWithPrefix(t *testing.T) {
	ss := getS3StorageService()
	s3Client, err := core.NewS3Client(ss, false, nil)
	require.Nil(t, err)
	require.NotNil(t, s3Client)

	// Upload a test directory with multiple files
	dirname := dirToUpload()
	key := filepath.Base(dirname)
	err = s3Client.Upload(dirname, key)
	require.Nil(t, err)

	// List objects with the directory prefix
	opts := minio.ListObjectsOptions{
		Recursive: true,
		Prefix:    key,
	}
	objects := s3Client.ListObjects(ss.Bucket, key, opts)
	require.NotNil(t, objects)

	// We should have multiple objects from the uploaded directory
	assert.True(t, len(objects) > 1, "Expected multiple objects with prefix")

	// All objects should start with our prefix
	for _, obj := range objects {
		assert.True(t, strings.HasPrefix(obj.Key, key), "Expected object key to have prefix: %s", key)
	}
}

func TestListObjectsWithMaxKeys(t *testing.T) {
	ss := getS3StorageService()
	s3Client, err := core.NewS3Client(ss, false, nil)
	require.Nil(t, err)
	require.NotNil(t, s3Client)

	// Upload a test directory with multiple files to ensure we have enough objects
	dirname := dirToUpload()
	key := filepath.Base(dirname)
	err = s3Client.Upload(dirname, key)
	require.Nil(t, err)

	// List objects with MaxKeys set to limit the number of results
	maxKeys := 2
	opts := minio.ListObjectsOptions{
		Recursive: true,
		Prefix:    key,
		MaxKeys:   maxKeys,
	}
	objects := s3Client.ListObjects(ss.Bucket, key, opts)
	require.NotNil(t, objects)

	// Verify that we only get the specified maximum number of keys
	assert.Equal(t, maxKeys, len(objects), "Expected exactly %d objects due to MaxKeys setting", maxKeys)

	// Verify that the objects we got have the correct prefix
	for _, obj := range objects {
		assert.True(t, strings.HasPrefix(obj.Key, key), "Expected object key to have prefix: %s", key)
	}
}

func TestGetObject(t *testing.T) {
	ss := getS3StorageService()
	s3Client, err := core.NewS3Client(ss, false, nil)
	require.Nil(t, err)
	require.NotNil(t, s3Client)

	// Upload a test file first
	filename := fileToUpload()
	key := filepath.Base(filename)
	err = s3Client.Upload(filename, key)
	require.Nil(t, err)

	// Now retrieve the object
	opts := minio.GetObjectOptions{}
	obj, err := s3Client.GetObject(ss.Bucket, key, opts)
	require.Nil(t, err)
	require.NotNil(t, obj)

	// Verify we can read the object's metadata
	stat, err := obj.Stat()
	require.Nil(t, err)
	assert.Equal(t, key, stat.Key)
	assert.True(t, stat.Size > 0, "Expected object size to be greater than 0")
	assert.NotEmpty(t, stat.ETag)

	// Close the object when we're done
	err = obj.Close()
	require.Nil(t, err)
}

func TestGetObjectNonExistent(t *testing.T) {
	ss := getS3StorageService()
	s3Client, err := core.NewS3Client(ss, false, nil)
	require.Nil(t, err)
	require.NotNil(t, s3Client)

	// Try to get an object that doesn't exist
	nonExistentKey := "non-existent-file-" + uuid.NewString() + ".txt"
	opts := minio.GetObjectOptions{}
	obj, err := s3Client.GetObject(ss.Bucket, nonExistentKey, opts)
	require.Nil(t, err)
	require.NotNil(t, obj)

	// The error won't occur until we try to read from or stat the object
	_, err = obj.Stat()
	require.NotNil(t, err, "Expected error when getting non-existent object")
}
