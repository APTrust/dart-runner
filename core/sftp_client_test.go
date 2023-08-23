package core_test

import (
	"path"
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

	testUpload(t, sftpClient, ss)
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

	testUpload(t, sftpClient, ss)
}

func testUpload(t *testing.T, sftpClient *sftp.Client, ss *core.StorageService) {
	fileToUpload := path.Join(util.PathToTestData(), "bags", "example.edu.sample_good.tar")
	bytesCopied, err := core.SFTPUpload(ss, fileToUpload, nil)
	require.Nil(t, err)
	assert.Equal(t, int64(23552), bytesCopied)
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
