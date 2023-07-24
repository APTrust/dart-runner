package core_test

import (
	"database/sql"
	"os"
	"path"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestJob(t *testing.T) *core.Job {
	workflow := getTestWorkflow(t)
	files := getTestFileList()
	tags := getTestTags()
	params := core.NewJobParams(workflow, "job_unit_test.tar", os.TempDir(), files, tags)
	return params.ToJob()
}

func TestJobValidate(t *testing.T) {
	job := core.NewJob()
	assert.False(t, job.Validate())
	require.Equal(t, 1, len(job.Errors))
	assert.Equal(t, "Job has nothing to package, validate, or upload.", job.Errors["Job"])

	job = getTestJob(t)
	assert.Equal(t, 0, len(job.Errors))

	// BagItJob without Profile should cause error
	job.BagItProfile = nil
	assert.False(t, job.Validate())

	assert.Equal(t, 2, len(job.Errors))
	assert.Equal(t, "BagIt packaging requires a BagItProfile.", job.Errors["Job.Package.BagItProfile"])
	assert.Equal(t, "Validation requires a BagItProfile.", job.Errors["Job.Validate.BagItProfile"])

	// Let's cause some more problems, shall we?
	for _, uploadOp := range job.UploadOps {
		uploadOp.StorageService.Login = ""
	}
	assert.False(t, job.Validate())
	assert.Equal(t, 3, len(job.Errors))
	assert.Equal(t, "StorageService requires a login name or access key id.", job.Errors["Job.StorageService.Login"])

	for _, uploadOp := range job.UploadOps {
		uploadOp.StorageService = nil
	}
	assert.False(t, job.Validate())
	assert.Equal(t, 4, len(job.Errors))
	assert.Equal(t, "UploadOperation requires a StorageService", job.Errors["UploadOp[0].StorageService"])
	assert.Equal(t, "UploadOperation requires a StorageService", job.Errors["UploadOp[1].StorageService"])
}

func TestJobFromJson(t *testing.T) {
	pathToJobFile := path.Join(util.PathToTestData(), "files", "sample_job.json")
	job, err := core.JobFromJson(pathToJobFile)
	require.Nil(t, err)
	require.NotNil(t, job)
}

func TestJobPersistence(t *testing.T) {
	// Clean up when test completes.
	defer core.ClearDartTable()

	// Insert three records for testing.
	job1 := getTestJob(t)
	job1.ID = uuid.NewString()
	job1.PackageOp.PackageName = "Job1.tar"
	job2 := getTestJob(t)
	job2.ID = uuid.NewString()
	job2.PackageOp.PackageName = "Job2.tar"
	job3 := getTestJob(t)
	job3.ID = uuid.NewString()
	job3.PackageOp.PackageName = "Job3.tar"
	assert.Nil(t, core.ObjSave(job1))
	assert.Nil(t, core.ObjSave(job2))
	assert.Nil(t, core.ObjSave(job3))

	// Make sure S1 was saved as expected.
	result := core.ObjFind(job1.ID)
	require.Nil(t, result.Error)
	job1Reload := result.Job()
	require.NotNil(t, job1Reload)
	assert.Equal(t, job1.ID, job1Reload.ID)
	assert.Equal(t, job1.Name(), job1Reload.Name())
	assert.Equal(t, job1.BagItProfile.BagItProfileInfo.BagItProfileIdentifier, job1Reload.BagItProfile.BagItProfileInfo.BagItProfileIdentifier)

	// Make sure order, offset and limit work on list query.
	result = core.ObjList(constants.TypeJob, "obj_name", 1, 0)
	require.Nil(t, result.Error)
	require.Equal(t, 1, len(result.Jobs))
	assert.Equal(t, job1.ID, result.Jobs[0].ID)

	// Make sure we can get all results.
	result = core.ObjList(constants.TypeJob, "obj_name", 100, 0)
	require.Nil(t, result.Error)
	require.Equal(t, 3, len(result.Jobs))
	assert.Equal(t, job1.ID, result.Jobs[0].ID)
	assert.Equal(t, job2.ID, result.Jobs[1].ID)
	assert.Equal(t, job3.ID, result.Jobs[2].ID)

	// Make sure delete works. Should return no error.
	assert.NoError(t, core.ObjDelete(job1))

	// Make sure the record was truly deleted.
	result = core.ObjFind(job1.ID)
	assert.Equal(t, sql.ErrNoRows, result.Error)
	assert.Nil(t, result.Job())
}
