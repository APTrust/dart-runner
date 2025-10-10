package core_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: SFTP tests require the SFTP server to be running.
// scripts/test.rb will start up the SFTP server in a docker
// container.
//
// If you want to run these tests manually, you can start the
// SFTP and Minio servers using the run.rb script from the
// DART repo: `./scripts/run.rb services`
//
// Also note that if you're manually checking files on the SFTP
// server, the SFTP password user and the SFTP key user are
// considered different users by the server and they have different
// home directories. That means their uploads will appear in different
// directories on the server.

// TODO: Verify that all files are present on the SFTP server
// and that permissions are correct.

func TestSftpUpload_Password(t *testing.T) {
	ss := getSftpStorageService()

	//sshClient, err := core.SSHConnect(ss)
	sftp, err := core.NewSFTPClient(ss, nil)

	require.Nil(t, err)
	require.NotNil(t, sftp)

	testUploadFile(t, sftp)
	testUploadDir(t, sftp)
}

func TestSftpUpload_SSHKey(t *testing.T) {
	ss := getSftpStorageService()
	ss.Login = "key_user"
	ss.Password = "not-a-valid-password"
	ss.LoginExtra = filepath.Join(util.PathToTestData(), "sftp", "sftp_user_key")

	sftp, err := core.NewSFTPClient(ss, nil)

	require.Nil(t, err)
	require.NotNil(t, sftp)

	testUploadFile(t, sftp)
	testUploadDir(t, sftp)
}

func fileToUpload() string {
	return filepath.Join(util.PathToTestData(), "bags", "example.edu.sample_good.tar")
}

func dirToUpload() string {
	return filepath.Join(util.PathToTestData(), "profiles")
}

func fileCount(dir string) (int, error) {
	fileCount := 0
	err := filepath.Walk(dir, func(filePath string, f os.FileInfo, err error) error {
		if f.Mode().IsRegular() {
			fileCount += 1
		}
		return nil
	})
	return fileCount, err

}

func testUploadFile(t *testing.T, sftp *core.SFTPClient) {
	destPath := filepath.Base(fileToUpload())
	err := sftp.Upload(fileToUpload(), filepath.Join("uploads", destPath))
	require.Nil(t, err)
}

func testUploadDir(t *testing.T, sftp *core.SFTPClient) {
	destPath := filepath.Base(dirToUpload())
	err := sftp.Upload(dirToUpload(), filepath.Join("uploads", destPath))
	require.Nil(t, err)
}

func TestUploadFileWithProgressBar(t *testing.T) {
	fileToUpload := filepath.Join(util.PathToTestData(), "bags", "example.edu.sample_good.tar")
	payloadSize, err := core.GetUploadPayloadSize(fileToUpload)
	require.Nil(t, err)
	testUploadWithProgressBar(t, "file", payloadSize)
}

func TestUploadDirectoryWithProgressBar(t *testing.T) {
	fileToUpload := filepath.Join(util.PathToTestData(), "profiles")
	payloadSize, err := core.GetUploadPayloadSize(fileToUpload)
	require.Nil(t, err)
	testUploadWithProgressBar(t, "directory", payloadSize)
}

// This is a tricky test. We want to make sure the SFTP uploader
// pushes messages into the progress reader's message channel.
// See comments inline...
func testUploadWithProgressBar(t *testing.T, fileOrDir string, payloadSize int64) {
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

	// Define our message channel and a go routine to listen to it.
	// The go routine will do nothing until later, after we call
	// core.SFTPUpload below, because until then, there will be no
	// traffic on the message channel.
	progress := core.NewStreamProgress(payloadSize, messageChannel)

	// Create a client with progress reader.
	ss := getSftpStorageService()
	sftp, err := core.NewSFTPClient(ss, progress)
	require.Nil(t, err)
	require.NotNil(t, sftp)

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
		testUploadFile(t, sftp)
	} else {
		testUploadDir(t, sftp)
	}

	wg.Wait()
}
