package controllers_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"

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

	// Note that because this is a pre-made test job,
	// we know for sure that it has a PackageOp, a
	// ValidationOp, and at least one UploadOp.
	outputPath := path.Join(os.TempDir(), "JobFilesControllerTest.tar")
	job.PackageOp.OutputPath = outputPath
	job.ValidationOp.PathToBag = outputPath
	job.UploadOps[0].SourceFiles = []string{outputPath}

	// Since paths will differ on each machine, we
	// need to set the source files dynamically so
	// we know they exist. This gives us three directories
	// from the current project with a variety of file types.
	job.PackageOp.SourceFiles = []string{
		path.Join(util.ProjectRoot(), "core"),
		path.Join(util.ProjectRoot(), "server", "assets"),
		path.Join(util.ProjectRoot(), "util"),
	}

	return job
}

func TestJobShowFiles(t *testing.T) {
	defer core.ClearDartTable()
	job := loadTestJob(t)
	assert.NoError(t, core.ObjSave(job))

	// Things we expect to see on this page:
	// Some headings and our list of source files.
	//
	// Also check for the buttons on the bottom
	// (Delete and Next). If there's a rendering problem
	// due to undefined variable references, these bottom
	// buttons won't render.
	expected := []string{
		"Files",
		"Desktop",
		"Documents",
		"Downloads",
		"File Path",
		"Directories",
		"Files",
		"Total Size",
		"Delete Job",
		"Next",
	}
	expected = append(expected, job.PackageOp.SourceFiles...)

	DoSimpleGetTest(t, fmt.Sprintf("/jobs/files/%s", job.ID), expected)

}

func TestJobAddFile(t *testing.T) {

}

func TestJobDeleteFile(t *testing.T) {

}
