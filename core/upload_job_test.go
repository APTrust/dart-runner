package core_test

import (
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
