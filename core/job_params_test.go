package core_test

import (
	"fmt"
	"path"
	"testing"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestWorkflow(t *testing.T) *core.Workflow {
	pathToProfile := path.Join(util.ProjectRoot(), "profiles", "aptrust-v2.2.json")
	profile, err := bagit.ProfileLoad(pathToProfile)
	require.Nil(t, err)
	require.NotNil(t, profile)

	storageServices := []*core.StorageService{
		getTestStorageService("s3", "s3.example.com"),
		getTestStorageService("sftp", "sftp.example.com"),
	}
	return &core.Workflow{
		ID:              constants.EmptyUUID,
		BagItProfile:    profile,
		Description:     "Workflow for unit tests",
		Name:            "Unit test workflow",
		PackageFormat:   constants.PackageFormatBagIt,
		StorageServices: storageServices,
	}
}

func getTestStorageService(protocol, host string) *core.StorageService {
	ss := core.NewStorageService()
	ss.AllowsUpload = true
	ss.Bucket = "test-bucket"
	ss.Host = host
	ss.Login = "user@example.com"
	ss.Password = "secret"
	ss.Name = fmt.Sprintf("Test % service", protocol)
	ss.Protocol = protocol
	return ss
}

func getTestTags() []*bagit.Tag {
	tags := []*bagit.Tag{
		{"bag-info.txt", "Source-Organization", "The Liberry"},
		{"aptrust-info.txt", "Title", "Baggy Pants"},
		{"aptrust-info.txt", "Description", "Those are chock full of heady goodness."},
		{"aptrust-info.txt", "Access", "Institution"},
		{"aptrust-info.txt", "Storage-Option", "Glacier-Deep-OH"},
		{"aptrust-info.txt", "Custom-Test-Tag", "Kwik-E-Mart"},
	}
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("tag-%d", i+1)
		value := fmt.Sprintf("value-%d", i+1)
		tags = append(tags, bagit.NewTag("custom-file.txt", name, value))
	}
	return tags
}

func getTestFileList() []string {
	return []string{
		path.Join(util.ProjectRoot(), "profiles"),
		util.PathToTestData(),
	}
}

func TestJobParams(t *testing.T) {
	workflow := getTestWorkflow(t)
	files := getTestFileList()
	tags := getTestTags()
	params := core.NewJobParams(workflow, "bag.tar", "/user/homer/bags", files, tags)
	assert.Equal(t, files, params.Files)
	assert.Equal(t, "bag.tar", params.PackageName)
	assert.Equal(t, "/user/homer/bags", params.OutputPath)
	assert.Equal(t, tags, params.Tags)
	assert.Equal(t, workflow, params.Workflow)
	assert.NotNil(t, params.Errors)
	assert.Empty(t, params.Errors)

	job := params.ToJob()
	require.NotNil(t, job)
	require.Empty(t, params.Errors)

	// Job has the right profile
	assert.NotNil(t, job.BagItProfile)
	assert.Equal(t, workflow.BagItProfile.Name, job.BagItProfile.Name)

	// Job has correctly initialized package op
	require.NotNil(t, job.PackageOp)
	assert.Equal(t, ".tar", job.PackageOp.BagItSerialization)
	assert.Equal(t, "bag.tar", job.PackageOp.PackageName)
	assert.Equal(t, "/user/homer/bags.tar", job.PackageOp.OutputPath)
	assert.EqualValues(t, files, job.PackageOp.SourceFiles)
	assert.NotNil(t, job.PackageOp.Result)
	assert.False(t, job.PackageOp.Result.WasAttempted())

	// Has right Validation Op
	require.NotNil(t, job.ValidationOp)
	assert.Equal(t, "/user/homer/bags/bag.tar", job.ValidationOp.PathToBag)

	// Job has correctly initialized upload operations
	uploadSrcFiles := []string{
		"/user/homer/bags.tar",
	}
	require.NotNil(t, job.UploadOps)
	require.Equal(t, 2, len(job.UploadOps))

	assert.NotNil(t, job.UploadOps[0].Results)
	assert.EqualValues(t, uploadSrcFiles, job.UploadOps[0].SourceFiles)
	assert.Equal(t, workflow.StorageServices[0].Host, job.UploadOps[0].StorageService.Host)

	assert.NotNil(t, job.UploadOps[1].Results)
	assert.EqualValues(t, uploadSrcFiles, job.UploadOps[1].SourceFiles)
	assert.Equal(t, workflow.StorageServices[1].Host, job.UploadOps[1].StorageService.Host)

	// Job has merged tags
	expectedTags := []string{
		"bagit.txt BagIt-Version: 0.97",
		"bagit.txt Tag-File-Character-Encoding: UTF-8",
		"bag-info.txt Source-Organization: The Liberry",
		"bag-info.txt Bag-Count: ",
		"bag-info.txt Bagging-Date: ",
		"bag-info.txt Bagging-Software: ",
		"bag-info.txt Bag-Group-Identifier: ",
		"bag-info.txt Internal-Sender-Description: ",
		"bag-info.txt Internal-Sender-Identifier: ",
		"bag-info.txt Payload-Oxum: ",
		"aptrust-info.txt Title: Baggy Pants",
		"aptrust-info.txt Access: Institution",
		"aptrust-info.txt Description: Those are chock full of heady goodness.",
		"aptrust-info.txt Storage-Option: Glacier-Deep-OH",
		"aptrust-info.txt Custom-Test-Tag: Kwik-E-Mart",
		"custom-file.txt tag-1: value-1",
		"custom-file.txt tag-2: value-2",
		"custom-file.txt tag-3: value-3",
	}
	for i, tag := range job.BagItProfile.Tags {
		strValue := fmt.Sprintf("%s %s", tag.TagFile, tag.ToFormattedString())
		assert.Equal(t, expectedTags[i], strValue)
	}
}
