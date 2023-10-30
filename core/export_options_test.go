package core_test

import (
	"fmt"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExportOptions(t *testing.T) {
	defer core.ClearDartTable()
	for i := 0; i < 5; i++ {
		setting := core.NewAppSetting(fmt.Sprintf("Setting %d", i), fmt.Sprintf("Value %d", i))
		require.NoError(t, core.ObjSaveWithoutValidation(setting))

		profile := core.NewBagItProfile()
		profile.Name = fmt.Sprintf("Profile %d", i)
		require.NoError(t, core.ObjSaveWithoutValidation(profile))

		repo := core.NewRemoteRepository()
		repo.Name = fmt.Sprintf("Repo %d", i)
		require.NoError(t, core.ObjSaveWithoutValidation(repo))

		ss := core.NewStorageService()
		ss.Name = fmt.Sprintf("Storage Service %d", i)
		require.NoError(t, core.ObjSaveWithoutValidation(ss))
	}

	opts := core.NewExportOptions()
	require.NotNil(t, opts)
	assert.Equal(t, 5, len(opts.AppSettings))
	assert.Equal(t, 5, len(opts.BagItProfiles))
	assert.Equal(t, 5, len(opts.RemoteRepositories))
	assert.Equal(t, 5, len(opts.StorageServices))
	for i := 0; i < 5; i++ {
		assert.Equal(t, fmt.Sprintf("Setting %d", i), opts.AppSettings[i].Name)
		assert.Equal(t, fmt.Sprintf("Profile %d", i), opts.BagItProfiles[i].Name)
		assert.Equal(t, fmt.Sprintf("Repo %d", i), opts.RemoteRepositories[i].Name)
		assert.Equal(t, fmt.Sprintf("Storage Service %d", i), opts.StorageServices[i].Name)
	}
	assert.Equal(t, 1, len(opts.AppSettingFields))
	assert.Equal(t, 5, len(opts.RemoteRepositoryFields))
	assert.Equal(t, 10, len(opts.StorageServiceFields))
}

func TestTagsForProfile(t *testing.T) {
	defer core.ClearDartTable()
	aptrustProfile := loadProfile(t, "aptrust-v2.2.json")
	btrProfile := loadProfile(t, "btr-v1.0-1.3.0.json")
	require.NoError(t, core.ObjSave(aptrustProfile))
	require.NoError(t, core.ObjSave(btrProfile))

	opts := core.NewExportOptions()
	require.NotNil(t, opts)
	require.Equal(t, 2, len(opts.BagItProfiles))

	testTagsMatchPairs(t, aptrustProfile)
	testTagsMatchPairs(t, btrProfile)
}

func testTagsMatchPairs(t *testing.T, profile *core.BagItProfile) {
	pairs, err := core.TagsForProfile(profile.ID)
	require.Nil(t, err)
	require.Equal(t, len(profile.Tags), len(pairs))
	for i, tag := range profile.Tags {
		assert.Equal(t, tag.ID, pairs[i].ID)
		assert.Equal(t, tag.TagName, pairs[i].Name)
	}
}
