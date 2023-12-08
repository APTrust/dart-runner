package core_test

import (
	"path/filepath"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/google/uuid"
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
	assert.Equal(t, 7, len(op.Errors))
	assert.Equal(t, "StorageService requires a valid ID.", op.Errors["StorageService.ID"])
	assert.Equal(t, "StorageService requires a protocol (s3, sftp, etc).", op.Errors["StorageService.Protocol"])
	assert.Equal(t, "StorageService requires a hostname or IP address.", op.Errors["StorageService.Host"])
	assert.Equal(t, "StorageService requires a login name or access key id.", op.Errors["StorageService.Login"])
	assert.Equal(t, "StorageService requires a password or secret access key, or the path to your SSH private key in the Login Extra field.", op.Errors["StorageService.Password"])
	assert.Equal(t, "UploadOperation source files are missing: file-does-not-exist", op.Errors["UploadOperation.SourceFiles"])

	ss = &core.StorageService{
		ID:       uuid.NewString(),
		Name:     "Patrick",
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

	assert.EqualValues(t, 0, op.PayloadSize)
	assert.NoError(t, op.CalculatePayloadSize())
	assert.EqualValues(t, 62464, op.PayloadSize)

	// Test with directory in source files.
	// This should calculate the size of all files
	// under that directory.
	files = []string{
		filepath.Join(util.ProjectRoot(), "core"),
	}
	op = core.NewUploadOperation(ss, files)
	require.NotNil(t, op)
	assert.True(t, op.Validate())
	assert.Empty(t, op.Errors)

	assert.EqualValues(t, 0, op.PayloadSize)
	assert.NoError(t, op.CalculatePayloadSize())
	assert.True(t, op.PayloadSize > 300000)

}
