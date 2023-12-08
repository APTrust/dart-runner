package core_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestWorkflow(t *testing.T) *core.Workflow {
	pathToProfile := filepath.Join(util.ProjectRoot(), "profiles", "aptrust-v2.2.json")
	profile, err := core.BagItProfileLoad(pathToProfile)
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
	ss := &core.StorageService{}
	ss.ID = uuid.NewString()
	ss.AllowsUpload = true
	ss.Bucket = "test-bucket"
	ss.Host = host
	ss.Login = "user@example.com"
	ss.Password = "secret"
	ss.Name = fmt.Sprintf("Test % service", protocol)
	ss.Protocol = protocol
	return ss
}

func getTestTags() []*core.Tag {
	tags := []*core.Tag{
		{TagFile: "bag-info.txt", TagName: "Source-Organization", Value: "The Liberry"},
		{TagFile: "aptrust-info.txt", TagName: "Title", Value: "Baggy Pants"},
		{TagFile: "aptrust-info.txt", TagName: "Description", Value: "Those are chock full of heady goodness."},
		{TagFile: "aptrust-info.txt", TagName: "Access", Value: "Institution"},
		{TagFile: "aptrust-info.txt", TagName: "Storage-Option", Value: "Glacier-Deep-OH"},
		{TagFile: "aptrust-info.txt", TagName: "Custom-Test-Tag", Value: "Kwik-E-Mart"},
		{TagFile: "bag-info.txt", TagName: "Repeated-Tag", Value: "1"},
		{TagFile: "bag-info.txt", TagName: "Repeated-Tag", Value: "2"},
		{TagFile: "bag-info.txt", TagName: "Repeated-Tag", Value: "3"},
	}
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("tag-%d", i+1)
		value := fmt.Sprintf("value-%d", i+1)
		tags = append(tags, core.NewTag("custom-file.txt", name, value))
	}
	return tags
}

func getTestFileList() []string {
	return []string{
		filepath.Join(util.ProjectRoot(), "profiles"),
		util.PathToTestData(),
	}
}

func TestJobParams(t *testing.T) {
	workflow := getTestWorkflow(t)
	files := getTestFileList()
	tags := getTestTags()
	params := core.NewJobParams(workflow, "bag.tar", "/user/homer/bag.tar", files, tags)
	assert.Equal(t, files, params.Files)
	assert.Equal(t, "bag.tar", params.PackageName)
	assert.Equal(t, "/user/homer/bag.tar", params.OutputPath)
	assert.Equal(t, tags, params.Tags)
	assert.Equal(t, workflow, params.Workflow)
	assert.NotNil(t, params.Errors)
	assert.Empty(t, params.Errors)

	job := params.ToJob()
	require.NotNil(t, job)
	require.Empty(t, params.Errors)

	// Job has the right profile
	assert.NotNil(t, job.BagItProfile)
	assert.Equal(t, "Copy of "+workflow.BagItProfile.Name, job.BagItProfile.Name)

	// Job has correctly initialized package op
	require.NotNil(t, job.PackageOp)
	assert.Equal(t, ".tar", job.PackageOp.BagItSerialization)
	assert.Equal(t, "bag.tar", job.PackageOp.PackageName)
	assert.Equal(t, "/user/homer/bag.tar", job.PackageOp.OutputPath)
	assert.EqualValues(t, files, job.PackageOp.SourceFiles)
	assert.NotNil(t, job.PackageOp.Result)
	assert.False(t, job.PackageOp.Result.WasAttempted())

	// TestRepeatedTags: https://github.com/APTrust/dart-runner/issues/7
	repeatedTags, err := job.BagItProfile.FindMatchingTags("TagName", "Repeated-Tag")
	require.Nil(t, err)
	require.Equal(t, 3, len(repeatedTags))
	assert.Equal(t, "bag-info.txt", repeatedTags[0].TagFile)
	assert.Equal(t, "Repeated-Tag", repeatedTags[0].TagName)
	assert.Equal(t, "1", repeatedTags[0].UserValue)

	assert.Equal(t, "bag-info.txt", repeatedTags[1].TagFile)
	assert.Equal(t, "Repeated-Tag", repeatedTags[1].TagName)
	assert.Equal(t, "2", repeatedTags[1].UserValue)

	assert.Equal(t, "bag-info.txt", repeatedTags[2].TagFile)
	assert.Equal(t, "Repeated-Tag", repeatedTags[2].TagName)
	assert.Equal(t, "3", repeatedTags[2].UserValue)

	// Has right Validation Op
	require.NotNil(t, job.ValidationOp)
	assert.Equal(t, "/user/homer/bag.tar", job.ValidationOp.PathToBag)

	// Job has correctly initialized upload operations
	uploadSrcFiles := []string{
		"/user/homer/bag.tar",
	}
	require.NotNil(t, job.UploadOps)
	require.Equal(t, 2, len(job.UploadOps))

	assert.NotNil(t, job.UploadOps[0].Result)
	assert.EqualValues(t, uploadSrcFiles, job.UploadOps[0].SourceFiles)
	assert.Equal(t, workflow.StorageServices[0].Host, job.UploadOps[0].StorageService.Host)

	assert.NotNil(t, job.UploadOps[1].Result)
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
		"bag-info.txt Repeated-Tag: 1",
		"bag-info.txt Repeated-Tag: 2",
		"bag-info.txt Repeated-Tag: 3",
		"custom-file.txt tag-1: value-1",
		"custom-file.txt tag-2: value-2",
		"custom-file.txt tag-3: value-3",
	}
	for i, tag := range job.BagItProfile.Tags {
		strValue := fmt.Sprintf("%s %s", tag.TagFile, tag.ToFormattedString())
		assert.Equal(t, expectedTags[i], strValue, i)
	}
}
