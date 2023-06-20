package core_test

import (
	"database/sql"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteRepositoryPersistence(t *testing.T) {

	// Clean up when test completes.
	defer core.ClearDartTable()

	// Insert records for testing.
	rr1 := core.NewRemoteRepository()
	rr1.Name = "RR-1"
	rr1.Url = "https://example.com/rr-1"
	rr2 := core.NewRemoteRepository()
	rr2.Name = "RR-2"
	rr2.Url = "https://example.com/rr-2"
	rr3 := core.NewRemoteRepository()
	rr3.Name = "RR-3"
	rr3.Url = "https://example.com/rr-3"
	assert.Nil(t, core.ObjSave(rr1))
	assert.Nil(t, core.ObjSave(rr2))
	assert.Nil(t, core.ObjSave(rr3))

	// Make sure S1 was saved as expected.
	s1Reload, err := core.RemoteRepositoryFind(rr1.ID)
	require.Nil(t, err)
	require.NotNil(t, s1Reload)
	assert.Equal(t, rr1.ID, s1Reload.ID)
	assert.Equal(t, rr1.Name, s1Reload.Name)
	assert.Equal(t, rr1.Url, s1Reload.Url)

	// Make sure order, offset and limit work on list query.
	repos, err := core.RemoteRepositoryList("obj_name", 1, 0)
	require.Nil(t, err)
	require.Equal(t, 1, len(repos))
	assert.Equal(t, rr1.ID, repos[0].ID)

	// Make sure we can get all results.
	repos, err = core.RemoteRepositoryList("obj_name", 100, 0)
	require.Nil(t, err)
	require.Equal(t, 3, len(repos))
	assert.Equal(t, rr1.ID, repos[0].ID)
	assert.Equal(t, rr2.ID, repos[1].ID)
	assert.Equal(t, rr3.ID, repos[2].ID)

	// Make sure delete works. Should return no error.
	assert.Nil(t, core.ObjDelete(rr1))

	// Make sure the record was truly deleted.
	deletedRecord, err := core.RemoteRepositoryFind(rr1.ID)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.Nil(t, deletedRecord)
}

func TestRemoteRepositoryValidation(t *testing.T) {
	// Clean up after test
	defer core.ClearDartTable()

	rr1 := core.NewRemoteRepository()
	rr1.Name = "RR-1"
	rr1.Url = "https://example.com/rr-1"
	assert.True(t, rr1.Validate())
	assert.Nil(t, core.ObjSave(rr1))

	rr1.Url = "this-aint-no-url"
	assert.False(t, rr1.Validate())
	assert.Equal(t, "Repository URL must be a valid URL beginning with http:// or https://.", rr1.Errors["Url"])
	assert.Equal(t, constants.ErrObjecValidation, core.ObjSave(rr1))
}
