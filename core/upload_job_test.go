package core_test

import (
	"path/filepath"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUploadJob(t *testing.T) {
	uploadJob := core.NewUploadJob()
	require.NotNil(t, uploadJob)
	assert.True(t, util.LooksLikeUUID(uploadJob.ID))
	assert.NotNil(t, uploadJob.Errors)
	assert.NotNil(t, uploadJob.PathsToUpload)
	assert.NotNil(t, uploadJob.StorageServiceIDs)
	assert.NotNil(t, uploadJob.UploadOps)
	assert.NotEmpty(t, uploadJob.Name)

	assert.Equal(t, uploadJob.Name, uploadJob.ObjName())
	assert.Equal(t, uploadJob.Name, uploadJob.String())
	assert.Equal(t, constants.TypeUploadJob, uploadJob.ObjType())
	assert.True(t, uploadJob.IsDeletable())
}

func TestUploadJobValidate(t *testing.T) {
	uploadJob := core.NewUploadJob()
	isValid := uploadJob.Validate()
	assert.False(t, isValid)

	require.Equal(t, 2, len(uploadJob.Errors))
	errors := uploadJob.GetErrors()
	assert.Equal(t, "You must select at least one item to upload.", errors["PathsToUpload"])
	assert.Equal(t, "Please choose at least one upload target.", errors["StorageServiceIDs"])

	uploadJob.StorageServiceIDs = []string{constants.EmptyUUID}
	uploadJob.PathsToUpload = []string{"/home/linus/file1.txt", "/home/linus/file2.pdf"}
	isValid = uploadJob.Validate()
	assert.True(t, isValid)
	assert.Empty(t, uploadJob.Errors)
}

func TestUploadJobToForm(t *testing.T) {
	uploadJob := core.NewUploadJob()
	uploadJob.StorageServiceIDs = []string{uuid.NewString(), uuid.NewString()}
	uploadJob.PathsToUpload = []string{"/home/linus/file1.txt", "/home/linus/file2.pdf"}

	form := uploadJob.ToForm()
	require.NotNil(t, form)

	pathsField := form.Fields["PathsToUpload"]
	require.NotNil(t, pathsField)
	assert.Equal(t, uploadJob.PathsToUpload, pathsField.Values)

	ssidField := form.Fields["StorageServiceIDs"]
	require.NotNil(t, ssidField)
	assert.Equal(t, uploadJob.StorageServiceIDs, ssidField.Values)
}

func TestUploadJobPersistence(t *testing.T) {
	defer core.ClearDartTable()
	ids := make([]string, 3)
	for i := 0; i < 3; i++ {
		uploadJob := core.NewUploadJob()
		uploadJob.StorageServiceIDs = []string{constants.EmptyUUID}
		uploadJob.PathsToUpload = []string{
			gofakeit.FarmAnimal(),
			gofakeit.Adjective(),
		}
		ids[i] = uploadJob.ID
		require.NoError(t, core.ObjSave(uploadJob))
	}
	result := core.ObjList(constants.TypeUploadJob, "obj_name", 20, 0)
	require.NoError(t, result.Error)
	assert.Equal(t, 3, len(result.UploadJobs))

	for _, id := range ids {
		result = core.ObjFind(id)
		require.NoError(t, result.Error)
		uploadJob := result.UploadJob()
		require.NotNil(t, uploadJob)
		assert.Equal(t, id, uploadJob.ID)
		assert.NoError(t, core.ObjDelete(uploadJob))
	}
}

// Note that for this to work, our local Minio and SFTP
// containers have to be running. They will be if you run
// tests via `./scripts/run.rb tests`.
//
// To debug this test, you must first start the Minio
// and SFTP containers using `./scripts/run/rb services`.
func TestUploadRun(t *testing.T) {
	defer core.ClearDartTable()
	localMinioService, err := core.LoadStorageServiceFixture("storage_service_local_minio.json")
	require.Nil(t, err)
	require.NotNil(t, localMinioService)

	localSFTPService, err := core.LoadStorageServiceFixture("storage_service_local_sftp.json")
	require.Nil(t, err)
	require.NotNil(t, localSFTPService)

	require.NoError(t, core.ObjSave(localMinioService))
	require.NoError(t, core.ObjSave(localSFTPService))

	uploadJob := core.NewUploadJob()
	uploadJob.StorageServiceIDs = []string{
		localMinioService.ID,
		localSFTPService.ID,
	}
	uploadJob.PathsToUpload = []string{
		filepath.Join(util.PathToTestData(), "files", "postbuild_test_workflow.json"),
		filepath.Join(util.PathToTestData(), "files", "aptrust_unit_test_job.json"),
		filepath.Join(util.PathToTestData(), "files", "test_batch.csv"),
	}

	result := uploadJob.Run(nil)
	assert.Equal(t, constants.ExitOK, result)
	assert.Equal(t, 2, len(uploadJob.UploadOps))
	for _, op := range uploadJob.UploadOps {
		assert.True(t, op.Result.Succeeded(), op.StorageService.Name)
	}

	// Test a failure case: Good file, but bad storage service.
	badStorageService := core.NewStorageService()
	badStorageService.AllowsDownload = true
	badStorageService.AllowsDownload = true
	badStorageService.Bucket = "bucket-one"
	badStorageService.Host = "127.0.0.1"
	badStorageService.Name = "Bad Storage Service"
	badStorageService.Port = 54321
	badStorageService.Protocol = constants.ProtocolS3
	badStorageService.Login = "Bad-login"
	badStorageService.Password = "Bad-password"
	require.Nil(t, core.ObjSave(badStorageService))

	uploadJob = core.NewUploadJob()
	uploadJob.StorageServiceIDs = []string{
		badStorageService.ID,
	}
	uploadJob.PathsToUpload = []string{
		filepath.Join(util.PathToTestData(), "files", "postbuild_test_workflow.json"),
	}
	result = uploadJob.Run(nil)
	assert.Equal(t, constants.ExitRuntimeErr, result)
	assert.Equal(t, 1, len(uploadJob.UploadOps))
	assert.False(t, uploadJob.UploadOps[0].Result.Succeeded(), uploadJob.UploadOps[0].StorageService.Name)

	// Test another failure case: good storage service but bad file.
	uploadJob = core.NewUploadJob()
	uploadJob.StorageServiceIDs = []string{
		localMinioService.ID,
	}
	uploadJob.PathsToUpload = []string{
		filepath.Join(util.PathToTestData(), "files", "postbuild_test_workflow.json"),
		filepath.Join(util.PathToTestData(), "files", "this-file-does-not-exist.txt"),
	}
	result = uploadJob.Run(nil)
	assert.Equal(t, constants.ExitRuntimeErr, result)
	assert.Equal(t, 1, len(uploadJob.UploadOps))
	assert.False(t, uploadJob.UploadOps[0].Result.Succeeded(), uploadJob.UploadOps[0].StorageService.Name)

}
