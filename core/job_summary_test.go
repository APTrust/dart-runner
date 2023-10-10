package core_test

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadTestJob(t *testing.T) *core.Job {
	filename := path.Join(util.PathToTestData(), "files", "aptrust_unit_test_job.json")
	data, err := os.ReadFile(filename)
	require.Nil(t, err)
	job := &core.Job{}
	err = json.Unmarshal(data, job)
	require.Nil(t, err)
	require.NotNil(t, job)
	return job
}

func TestNewJobSummary(t *testing.T) {
	job := loadTestJob(t)
	info := core.NewJobSummary(job)
	assert.Equal(t, job.ID, info.ID)
	assert.Equal(t, job.Name(), info.Name)
	assert.Equal(t, "APTrust Profile for Wasabi VA ingest", info.BagItProfileDescription)
	assert.Equal(t, "APTrust - Wasabi VA", info.BagItProfileName)
	assert.Equal(t, int64(0), info.ByteCount)
	assert.Equal(t, "0", info.ByteCountFormatted)
	assert.Equal(t, "0 B", info.ByteCountHuman)
	assert.Equal(t, int64(0), info.DirectoryCount)
	assert.Equal(t, int64(0), info.PayloadFileCount)
	assert.True(t, info.HasBagItProfile)
	assert.True(t, info.HasPackageOp)
	assert.True(t, info.HasUploadOps)
	assert.Equal(t, "to-be-set-by-unit-test.tar", info.OutputPath)
	assert.Equal(t, constants.PackageFormatBagIt, info.PackageFormat)
	assert.Equal(t, "APTrust-S3-Bag-01.tar", info.PackageName)

	// Get some actual numbers in there
	job.ByteCount = 999888777666
	job.DirCount = 33
	job.PayloadFileCount = 1271

	info = core.NewJobSummary(job)
	assert.Equal(t, int64(999888777666), info.ByteCount)
	assert.Equal(t, "999,888,777,666", info.ByteCountFormatted)
	assert.Equal(t, "931.2 GB", info.ByteCountHuman)
	assert.Equal(t, int64(33), info.DirectoryCount)
	assert.Equal(t, int64(1271), info.PayloadFileCount)

	// If these items aren't present, make sure
	// the info object reflects that.
	job.BagItProfile = nil
	job.PackageOp = nil
	job.UploadOps = nil

	info = core.NewJobSummary(job)
	assert.False(t, info.HasBagItProfile)
	assert.False(t, info.HasPackageOp)
	assert.False(t, info.HasUploadOps)

	assert.Empty(t, info.BagItProfileDescription)
	assert.Empty(t, info.BagItProfileName)

	assert.Empty(t, info.OutputPath)
	assert.Empty(t, info.PackageFormat)
	assert.Empty(t, info.PackageName)
}
