package core_test

import (
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
