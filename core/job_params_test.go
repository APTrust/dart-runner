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
		{"aptrust-info.txt", "Access", "Institution"},
		{"aptrust-info.txt", "Storage-Option", "Glacier-Deep-OH"},
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

	// Make sure job has package op, upload ops, merged tags
}
