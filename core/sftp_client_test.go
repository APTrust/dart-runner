package core_test

import (
	"path"
	"sync"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sftpUserName = "pw_user"
	sftpPassword = "password"
	sftpHost     = "127.0.0.1"
	sftpPort     = 2222
	tarFileSize  = int64(23552)
)

// Note: SFTP tests require the SFTP server to be running.
// scripts/test.rb will start up the server in a docker container.
func TestSftpUploadWithPassword(t *testing.T) {
	ss := getSftpStorageService()

	sshClient, err := core.SSHConnect(ss)
	require.Nil(t, err)
	require.NotNil(t, sshClient)

	sftpClient, err := core.SFTPConnect(ss)
	require.Nil(t, err)
	require.NotNil(t, sftpClient)

	testUploadWithoutProgress(t, sftpClient, ss)
	testUploadWithProgress(t, sftpClient, ss)
}

func TestSftpUploadWithSSHKey(t *testing.T) {
	ss := getSftpStorageService()
	ss.Login = "key_user"
	ss.Password = "not-a-valid-password"
	ss.LoginExtra = path.Join(util.PathToTestData(), "sftp", "sftp_user_key")

	sshClient, err := core.SSHConnect(ss)
	require.Nil(t, err)
	require.NotNil(t, sshClient)

	sftpClient, err := core.SFTPConnect(ss)
	require.Nil(t, err)
	require.NotNil(t, sftpClient)

	testUploadWithoutProgress(t, sftpClient, ss)
	testUploadWithProgress(t, sftpClient, ss)
}

func testUploadWithoutProgress(t *testing.T, sftpClient *sftp.Client, ss *core.StorageService) {
	fileToUpload := path.Join(util.PathToTestData(), "bags", "example.edu.sample_good.tar")
	bytesCopied, err := core.SFTPUpload(ss, fileToUpload, nil)
	require.Nil(t, err)
	assert.Equal(t, tarFileSize, bytesCopied)
}

// This is a tricky test. We want to make sure the SFTP uploader
// pushes messages into the progress reader's message channel.
// See comments inline...
func testUploadWithProgress(t *testing.T, sftpClient *sftp.Client, ss *core.StorageService) {
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
	progress := core.NewStreamProgress(tarFileSize, messageChannel)
	go func() {
		if msg, ok := <-messageChannel; ok {
			assert.Equal(t, "Sent 23.0 kB of 23.0 kB (100%)", msg.Message)
			iReadSomething = true
			wg.Done()
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
	fileToUpload := path.Join(util.PathToTestData(), "bags", "example.edu.sample_good.tar")
	bytesCopied, err := core.SFTPUpload(ss, fileToUpload, progress)
	require.Nil(t, err)
	assert.Equal(t, tarFileSize, bytesCopied)
	wg.Wait()
}

func getSftpStorageService() *core.StorageService {
	return &core.StorageService{
		ID:           uuid.NewString(),
		Name:         "Local SFTP test service",
		AllowsUpload: true,
		Bucket:       "uploads",
		Host:         sftpHost,
		Login:        sftpUserName,
		Password:     sftpPassword,
		Port:         sftpPort,
		Protocol:     constants.ProtocolSFTP,
	}
}
