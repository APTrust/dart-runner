package core_test

import (
	"database/sql"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
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
	result := core.ObjFind(rr1.ID)
	require.Nil(t, result.Error)
	s1Reload := result.RemoteRepository()
	require.NotNil(t, s1Reload)
	assert.Equal(t, rr1.ID, s1Reload.ID)
	assert.Equal(t, rr1.Name, s1Reload.Name)
	assert.Equal(t, rr1.Url, s1Reload.Url)

	// Make sure order, offset and limit work on list query.
	result = core.ObjList(constants.TypeRemoteRepository, "obj_name", 1, 0)
	require.Nil(t, result.Error)
	require.Equal(t, 1, len(result.RemoteRepositories))
	assert.Equal(t, rr1.ID, result.RemoteRepositories[0].ID)

	// Make sure we can get all results.
	result = core.ObjList(constants.TypeRemoteRepository, "obj_name", 100, 0)
	require.Nil(t, result.Error)
	require.Equal(t, 3, len(result.RemoteRepositories))
	assert.Equal(t, rr1.ID, result.RemoteRepositories[0].ID)
	assert.Equal(t, rr2.ID, result.RemoteRepositories[1].ID)
	assert.Equal(t, rr3.ID, result.RemoteRepositories[2].ID)

	// Make sure delete works. Should return no error.
	assert.Nil(t, core.ObjDelete(rr1))

	// Make sure the record was truly deleted.
	result = core.ObjFind(rr1.ID)
	assert.Equal(t, sql.ErrNoRows, result.Error)
	assert.Nil(t, result.RemoteRepository())
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

func TestRemoteRepositoryToForm(t *testing.T) {
	repo := core.NewRemoteRepository()
	repo.APIToken = "1234"
	repo.LoginExtra = "login-xtra"
	repo.Name = "test repo"
	repo.PluginID = constants.PluginIdAPTrustClient
	repo.Url = "https://repo.example.com"
	repo.UserID = "spongebob"

	form := repo.ToForm()
	assert.Equal(t, 7, len(form.Fields))
	assert.True(t, form.UserCanDelete)
	assert.Equal(t, repo.ID, form.Fields["ID"].Value)
	assert.Equal(t, repo.Name, form.Fields["Name"].Value)
	assert.Equal(t, repo.APIToken, form.Fields["APIToken"].Value)
	assert.Equal(t, repo.LoginExtra, form.Fields["LoginExtra"].Value)
	assert.Equal(t, repo.PluginID, form.Fields["PluginID"].Value)
	assert.Equal(t, repo.Url, form.Fields["Url"].Value)
	assert.Equal(t, repo.UserID, form.Fields["UserID"].Value)

	assert.True(t, form.Fields["ID"].Required)
	assert.True(t, form.Fields["Name"].Required)
	assert.True(t, form.Fields["Url"].Required)
	assert.False(t, form.Fields["APIToken"].Required)
	assert.False(t, form.Fields["LoginExtra"].Required)
	assert.False(t, form.Fields["PluginID"].Required)
	assert.False(t, form.Fields["UserID"].Required)
}

func TestRemoteRepositoryPersistentObject(t *testing.T) {
	repo := core.NewRemoteRepository()
	repo.Name = "test repo"

	assert.Equal(t, constants.TypeRemoteRepository, repo.ObjType())
	assert.Equal(t, "RemoteRepository", repo.ObjType())
	assert.Equal(t, repo.ID, repo.ObjID())
	assert.True(t, util.LooksLikeUUID(repo.ObjID()))
	assert.True(t, repo.IsDeletable())
	assert.Equal(t, "test repo", repo.ObjName())
	assert.Equal(t, "RemoteRepository: 'test repo'", repo.String())
	assert.Empty(t, repo.GetErrors())

	repo.Errors = map[string]string{
		"Error 1": "Message 1",
		"Error 2": "Message 2",
	}

	assert.Equal(t, 2, len(repo.GetErrors()))
	assert.Equal(t, "Message 1", repo.GetErrors()["Error 1"])
	assert.Equal(t, "Message 2", repo.GetErrors()["Error 2"])
}

func TestHasPlaintextAPIToken(t *testing.T) {
	repo := &core.RemoteRepository{}

	// False because password is empty
	assert.False(t, repo.HasPlaintextAPIToken())

	// False because password is empty
	repo.APIToken = "    "
	assert.False(t, repo.HasPlaintextAPIToken())

	// False because password uses env variable
	repo.APIToken = "env:REPO_TOKEN"
	assert.False(t, repo.HasPlaintextAPIToken())

	// True because password is non-empty, and
	// not an ENV variable.
	repo.APIToken = "this-here-is-secret"
	assert.True(t, repo.HasPlaintextAPIToken())
}
