package core_test

import (
	"os"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
)

func TestStorageService(t *testing.T) {
	ss := &core.StorageService{}
	assert.False(t, ss.Validate())
	assert.Equal(t, 5, len(ss.Errors))
	assert.Equal(t, "StorageService requires a protocol (s3, sftp, etc).", ss.Errors["StorageService.Protocol"])
	assert.Equal(t, "StorageService requires a hostname or IP address.", ss.Errors["StorageService.Host"])
	assert.Equal(t, "StorageService requires a bucket or folder name.", ss.Errors["StorageService.Bucket"])
	assert.Equal(t, "StorageService requires a login name or access key id.", ss.Errors["StorageService.Login"])
	assert.Equal(t, "StorageService requires a password or secret access key.", ss.Errors["StorageService.Password"])

	ss = &core.StorageService{
		Host:     "example.com",
		Bucket:   "uploads",
		Login:    "user@example.com",
		Password: "secret",
		Protocol: "s3",
	}
	assert.True(t, ss.Validate())
	assert.Empty(t, ss.Errors)
	assert.Equal(t, "s3://example.com/uploads/bag.tar", ss.URL("bag.tar"))
}

func TestStorageServiceSensitiveData(t *testing.T) {
	creds := map[string]string{
		"RUNNER_UNIT_TEST_SS_LOGIN": "user555@example.com",
		"RUNNER_UNIT_TEST_SS_PWD":   "Secret! Shh!!",
	}
	for key, value := range creds {
		os.Setenv(key, value)
	}
	defer func() {
		for key := range creds {
			os.Unsetenv(key)
		}
	}()

	ss := &core.StorageService{}
	ss.Login = "insecure@example.com"
	ss.Password = "not secret"

	assert.Equal(t, "insecure@example.com", ss.GetLogin())
	assert.Equal(t, "not secret", ss.GetPassword())

	ss.Login = "env:RUNNER_UNIT_TEST_SS_LOGIN"
	ss.Password = "env:RUNNER_UNIT_TEST_SS_PWD"

	assert.Equal(t, creds["RUNNER_UNIT_TEST_SS_LOGIN"], ss.GetLogin())
	assert.Equal(t, creds["RUNNER_UNIT_TEST_SS_PWD"], ss.GetPassword())
}

func TestStorageServiceURL(t *testing.T) {
	ss := &core.StorageService{}
	ss.Protocol = "s3"
	ss.Host = "example.com"
	ss.Bucket = "bucky"
	assert.Equal(t, "s3://example.com/bucky/document.pdf", ss.URL("document.pdf"))

	ss.Port = 9999
	assert.Equal(t, "s3://example.com:9999/bucky/document.pdf", ss.URL("document.pdf"))
}

func TestStorageServiceHostAndPort(t *testing.T) {
	ss := &core.StorageService{}
	ss.Host = "example.com"
	assert.Equal(t, "example.com", ss.HostAndPort())

	ss.Port = 9999
	assert.Equal(t, "example.com:9999", ss.HostAndPort())
}
