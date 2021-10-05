package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadOperation(t *testing.T) {
	op := core.NewUploadOperation(nil, nil)
	require.NotNil(t, op)
	assert.False(t, op.Validate())
	assert.Equal(t, 2, len(op.Errors))
	assert.Equal(t, "UploadOperation requires a StorageService", op.Errors["UploadOperation.StorageService"])
	assert.Equal(t, "UploadOperation requires one or more files to upload", op.Errors["UploadOperation.SourceFiles"])

	ss := &core.StorageService{}
	files := []string{
		"file-does-not-exist",
	}
	op = core.NewUploadOperation(ss, files)
	require.NotNil(t, op)
	assert.False(t, op.Validate())
	assert.Equal(t, 6, len(op.Errors))
	assert.Equal(t, "StorageService requires a protocol (s3, sftp, etc).", op.Errors["StorageService.Protocol"])
	assert.Equal(t, "StorageService requires a hostname or IP address.", op.Errors["StorageService.Host"])
	assert.Equal(t, "StorageService requires a bucket or folder name.", op.Errors["StorageService.Bucket"])
	assert.Equal(t, "StorageService requires a login name or access key id.", op.Errors["StorageService.Login"])
	assert.Equal(t, "StorageService requires a password or secret access key.", op.Errors["StorageService.Password"])
	assert.Equal(t, "UploadOperation source files are missing: file-does-not-exist", op.Errors["UploadOperation.SourceFiles"])

	ss = &core.StorageService{
		Host:     "example.com",
		Bucket:   "uploads",
		Login:    "user@example.com",
		Password: "secret",
		Protocol: "sftp",
	}
	files = []string{
		util.PathToUnitTestBag("test.edu.btr_good_sha256.tar"),
		util.PathToUnitTestBag("test.edu.btr_good_sha512.tar"),
	}
	op = core.NewUploadOperation(ss, files)
	require.NotNil(t, op)
	assert.True(t, op.Validate())
	assert.Empty(t, op.Errors)
}
