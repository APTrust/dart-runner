package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidationJob(t *testing.T) {
	valJob := core.NewValidationJob()
	require.NotNil(t, valJob)
	assert.True(t, util.LooksLikeUUID(valJob.ID))
	assert.NotNil(t, valJob.Errors)
	assert.NotNil(t, valJob.PathsToValidate)
	assert.NotNil(t, valJob.ValidationOps)
	assert.NotEmpty(t, valJob.Name)

	assert.Equal(t, valJob.Name, valJob.ObjName())
	assert.Equal(t, valJob.Name, valJob.String())
	assert.Equal(t, constants.TypeValidationJob, valJob.ObjType())
	assert.True(t, valJob.IsDeletable())
}

func TestValidationJobValidate(t *testing.T) {
	valJob := core.NewValidationJob()
	isValid := valJob.Validate()
	assert.False(t, isValid)
	assert.Equal(t, "You must select at least one item to validate.", valJob.Errors["PathsToValidate"])
	assert.Equal(t, "Please choose a BagIt profile.", valJob.Errors["BagItProfileID"])

	valJob.BagItProfileID = constants.ProfileIDAPTrust
	valJob.PathsToValidate = []string{"/usr/file1.txt", "/var/lib/file2.png"}
	isValid = valJob.Validate()
	assert.True(t, isValid)
	assert.Empty(t, valJob.Errors)
}

func TestValidationJobToForm(t *testing.T) {
	valJob := core.NewValidationJob()
	valJob.BagItProfileID = constants.ProfileIDAPTrust
	valJob.PathsToValidate = []string{"/usr/file1.txt", "/var/lib/file2.png"}
	form := valJob.ToForm()
	require.NotNil(t, form)

	profileIDField := form.Fields["BagItProfileID"]
	require.NotNil(t, profileIDField)
	assert.Equal(t, valJob.BagItProfileID, profileIDField.Value)

	pathsField := form.Fields["PathsToValidate"]
	require.NotNil(t, pathsField)
	assert.Equal(t, valJob.PathsToValidate, pathsField.Values)
}

func TestValidationJobPersistence(t *testing.T) {
	defer core.ClearDartTable()
	ids := make([]string, 3)
	for i := 0; i < 3; i++ {
		valJob := core.NewValidationJob()
		valJob.BagItProfileID = constants.ProfileIDBTR
		valJob.PathsToValidate = []string{
			gofakeit.BeerName(),
			gofakeit.Adjective(),
		}
		ids[i] = valJob.ID
		require.NoError(t, core.ObjSave(valJob))
	}
	result := core.ObjList(constants.TypeValidationJob, "obj_name", 20, 0)
	require.NoError(t, result.Error)
	assert.Equal(t, 3, len(result.ValidationJobs))

	for _, id := range ids {
		result = core.ObjFind(id)
		require.NoError(t, result.Error)
		valJob := result.ValidationJob()
		require.NotNil(t, valJob)
		assert.Equal(t, id, valJob.ID)
		assert.NoError(t, core.ObjDelete(valJob))
	}
}
