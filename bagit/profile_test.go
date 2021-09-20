package bagit_test

import (
	"path"
	"testing"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This also implicitly tests BagItProfileFromJson
func TestBagItProfileLoad(t *testing.T) {
	filename := path.Join(util.ProjectRoot(), "profiles", "aptrust-v2.2.json")
	profile, err := bagit.ProfileLoad(filename)
	assert.Nil(t, err)
	require.NotNil(t, profile)

	// Spot check
	assert.Equal(t, "support@aptrust.org", profile.BagItProfileInfo.ContactEmail)
	assert.Equal(t, 14, len(profile.Tags))
	assert.Equal(t, "BagIt-Version", profile.Tags[0].TagName)
	assert.Equal(t, "Storage-Option", profile.Tags[13].TagName)
	assert.Equal(t, 9, len(profile.Tags[13].Values))

	// Test with bad filename
	_, err = bagit.ProfileLoad("__file_does_not_exist__")
	assert.NotNil(t, err)

	// Test with non-JSON file. This is a tar file.
	filename = path.Join(util.PathToUnitTestBag("example.edu.tagsample_good.tar"))
	_, err = bagit.ProfileLoad(filename)
	assert.NotNil(t, err)
}

func TestGetTagDef(t *testing.T) {
	filename := path.Join(util.ProjectRoot(), "profiles", "aptrust-v2.2.json")
	profile, err := bagit.ProfileLoad(filename)
	assert.Nil(t, err)
	require.NotNil(t, profile)

	tagDef := profile.GetTagDef("aptrust-info.txt", "Access")
	require.NotNil(t, tagDef)
	assert.Equal(t, "aptrust-info.txt", tagDef.TagFile)
	assert.Equal(t, "Access", tagDef.TagName)

	tagDef = profile.GetTagDef("aptrust-info.txt", "Tag-Does-Not-Exist")
	assert.Nil(t, tagDef)
}
