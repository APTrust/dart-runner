package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/google/uuid"
)

const (
	sftpUserName = "demo"
	sftpPassword = "password"
	sftpHost     = "127.0.0.1"
	sftpPort     = 2022
)

func TestSftpUpload(t *testing.T) {
	// ss := getSftpStorageService()

	// sshClient, err := core.SSHConnect(ss)
	// require.Nil(t, err)
	// require.NotNil(t, sshClient)

	// sftpClient, err := core.SFTPConnect(ss)
	// require.Nil(t, err)
	// require.NotNil(t, sftpClient)

	// fileToUpload := path.Join(util.PathToTestData(), "bags", "example.edu.sample_good.tar")
	// bytesCopied, err := core.SFTPUpload(ss, fileToUpload, nil)
	// require.Nil(t, err)
	// assert.Equal(t, -1, bytesCopied)

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
