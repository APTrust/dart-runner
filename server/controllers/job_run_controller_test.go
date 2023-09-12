package controllers_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
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
	// This requires use of StreamRecorder to capture server-sent events.
	// This also requires a local running Minio server to receive the
	// upload portion of the job. If you run tests using
	// `./scripts/run.rb tests`, the script will start the server for you.
	// If you're running this test by itself or inside VS Code,
	// be sure the start the Minio server first. See scripts/run.rb.
	defer core.ClearDartTable()
	job := loadTestJob(t)
	require.NoError(t, core.ObjSave(job))

	recorder := NewStreamRecorder()
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/jobs/run/%s", job.ID), nil)
	go dartServer.ServeHTTP(recorder, req)

	for recorder.LastEvent == nil {
		time.Sleep(250 * time.Millisecond)
	}

	assert.Equal(t, http.StatusOK, recorder.Code)
	//html := recorder.Body.String()
	//fmt.Println(html)
	assert.True(t, recorder.EventCount > 100)
	assert.Equal(t, "Job completed with exit code 0", recorder.LastEvent.Message)

	jobResult := recorder.ResultEvent.JobResult
	require.NotNil(t, jobResult)

	// Check some basic details...
	assert.Equal(t, "APTrust-S3-Bag-01.tar", jobResult.JobName)
	assert.True(t, jobResult.PayloadByteCount > 15000000, jobResult.PayloadByteCount)
	assert.True(t, jobResult.PayloadFileCount > int64(1000), jobResult.PayloadFileCount)

	// Make sure job definition was valid.
	assert.Empty(t, jobResult.ValidationErrors)

	// Make sure the packaging step (bagging) was attempted and did succeed.
	assert.True(t, jobResult.PackageResult.WasAttempted())
	assert.True(t, jobResult.PackageResult.Succeeded())
	assert.Empty(t, jobResult.PackageResult.Errors)

	// Make sure the job validated the bag and found no errors.
	assert.True(t, jobResult.ValidationResult.WasAttempted())
	assert.True(t, jobResult.ValidationResult.Succeeded())
	assert.Empty(t, jobResult.ValidationResult.Errors)

	// Make sure all upload operations were attempted and succeeded.
	for _, uploadResult := range jobResult.UploadResults {
		assert.True(t, uploadResult.WasAttempted())
		assert.True(t, uploadResult.Succeeded())
		assert.Empty(t, uploadResult.Errors)
	}
}
