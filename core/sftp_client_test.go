package core_test

import (
	"path/filepath"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/require"
)

// Note: SFTP tests require the SFTP server to be running.
// scripts/test.rb will start up the server in a docker container.
func TestSftpUpload_Password(t *testing.T) {
	ss := getSftpStorageService()

	//sshClient, err := core.SSHConnect(ss)
	sftp, err := core.NewSFTPClient(ss)

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

	sftp, err := core.NewSFTPClient(ss)

	require.Nil(t, err)
	require.NotNil(t, sftp)

	testUploadFile(t, sftp)
	testUploadDir(t, sftp)
}

func testUploadFile(t *testing.T, sftp *core.SFTPClient) {
	fileToUpload := filepath.Join(util.PathToTestData(), "bags", "example.edu.sample_good.tar")
	destPath := filepath.Base(fileToUpload)
	err := sftp.Upload(fileToUpload, filepath.Join("uploads", destPath))
	require.Nil(t, err)
}

func testUploadDir(t *testing.T, sftp *core.SFTPClient) {
	fileToUpload := filepath.Join(util.PathToTestData(), "profiles")
	destPath := filepath.Base(fileToUpload)
	err := sftp.Upload(fileToUpload, filepath.Join("uploads", destPath))
	require.Nil(t, err)
}
