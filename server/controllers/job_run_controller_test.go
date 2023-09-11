package controllers_test

import (
	"fmt"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/require"
)

func TestJobRunShow(t *testing.T) {
	defer core.ClearDartTable()
	job := loadTestJob(t)
	require.NoError(t, core.ObjSave(job))
	require.NoError(t, core.ObjSave(job.BagItProfile))
	expectedContent := []string{
		"Package Name",
		"BagIt Profile",
		"Payload Summary",
		"Files to Package",
		"Upload To",
		job.BagItProfile.Name,
		job.Name(),
		job.ID,
		job.PackageOp.PackageName,
		job.PackageOp.OutputPath,
		"Local Minio",     // upload target
		"Payload Summary", // file count, dir count and bytes will change as project changes
		"Directories",
		"Files",
		"MB",
		"Back",
		"Run Job",
		"Create Workflow",
	}
	expectedContent = append(expectedContent, job.PackageOp.SourceFiles...)
	pageUrl := fmt.Sprintf("/jobs/summary/%s", job.ID)
	DoSimpleGetTest(t, pageUrl, expectedContent)
}

func TestJobRunExecute(t *testing.T) {

	// TODO: This requires use of TestResponseRecorder

}
