package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAppSettings(t *testing.T) {
	defer core.ClearDartTable()
	settings, err := core.CreateAppSettings(4)
	require.NoError(t, err)
	assert.Equal(t, 4, len(settings))
}

func TestCreateBagItProfiles(t *testing.T) {
	defer core.ClearDartTable()
	profiles, err := core.CreateBagItProfiles(4)
	require.NoError(t, err)
	assert.Equal(t, 4, len(profiles))
}

func TestCreateRemoteRepos(t *testing.T) {
	defer core.ClearDartTable()
	repos, err := core.CreateRemoteRepos(4)
	require.NoError(t, err)
	assert.Equal(t, 4, len(repos))
}

func TestCreateStorageServices(t *testing.T) {
	defer core.ClearDartTable()
	services, err := core.CreateStorageServices(4)
	require.NoError(t, err)
	assert.Equal(t, 4, len(services))
}
