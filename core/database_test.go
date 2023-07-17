package core_test

import (
	"fmt"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleArtifact() *core.Artifact {
	return &core.Artifact{
		ID:       uuid.NewString(),
		BagName:  "Tote Bag",
		ItemType: constants.ItemTypeFile,
		FileName: "manifest-md5.txt",
		RawData:  "raw data goes here",
	}
}

func TestInitSchema(t *testing.T) {
	defer core.ClearDartTable()
	defer core.ClearArtifactsTable()
	assert.NoError(t, core.InitSchema())
	appSetting := core.NewAppSetting("Name 1", "Value 1")
	assert.NoError(t, core.ObjSave(appSetting))
	artifact := sampleArtifact()
	assert.NoError(t, core.ArtifactSave(artifact))
}

func TestObjPersistenceOperations(t *testing.T) {
	defer core.ClearDartTable()

	// Test ObjCount() with no data
	count, err := core.ObjCount(constants.TypeAppSetting)
	require.Nil(t, err)
	assert.Equal(t, 0, count)

	// Test ObjSave()
	objIds := make([]string, 5)
	for i := 0; i < 5; i++ {
		setting := core.NewAppSetting(fmt.Sprintf("Name %d", i), fmt.Sprintf("Value %d", i))
		assert.Nil(t, core.ObjSave(setting))
		objIds[i] = setting.ID
	}

	// Test ObjCount() with data
	count, err = core.ObjCount(constants.TypeAppSetting)
	require.Nil(t, err)
	assert.Equal(t, 5, count)

	// Test ObjExists()
	for _, objId := range objIds {
		exists, err := core.ObjExists(objId)
		require.Nil(t, err)
		assert.True(t, exists)
	}

	// Random UUID should not exist in DB
	exists, err := core.ObjExists(uuid.NewString())
	require.Nil(t, err)
	assert.False(t, exists)

	// Test ObjList
	result := core.ObjList(constants.TypeAppSetting, "obj_name", 20, 0)
	require.Nil(t, result.Error)
	assert.Equal(t, 5, result.ObjCount)
	assert.Equal(t, 5, len(result.AppSettings))

	// Test delete
	setting := result.AppSettings[0]
	err = core.ObjDelete(setting)
	require.Nil(t, err)
	exists, err = core.ObjExists(setting.ID)
	require.Nil(t, err)
	assert.False(t, exists)
}

func TestArtifactPersistenceOperations(t *testing.T) {
	defer core.ClearArtifactsTable()

	// Test ArtifactSave()
	ids := make([]string, 5)
	for i := 0; i < 5; i++ {
		artifact := sampleArtifact()
		artifact.FileName = fmt.Sprintf("File %d", i)
		if i%2 == 0 {
			artifact.BagName = "Bag 1"
		} else {
			artifact.BagName = "Bag 2"
		}
		err := core.ArtifactSave(artifact)
		require.Nil(t, err)
		ids[i] = artifact.ID
	}

	// Test ArtifactList()
	artifacts, err := core.ArtifactList("Bag 1")
	require.Nil(t, err)
	assert.Equal(t, 3, len(artifacts))
	assert.Equal(t, "File 0", artifacts[0].FileName)
	assert.Equal(t, "File 2", artifacts[1].FileName)
	assert.Equal(t, "File 4", artifacts[2].FileName)

	artifacts, err = core.ArtifactList("Bag 2")
	require.Nil(t, err)
	assert.Equal(t, 2, len(artifacts))
	assert.Equal(t, "File 1", artifacts[0].FileName)
	assert.Equal(t, "File 3", artifacts[1].FileName)

	// Test ArtifactFind()
	for _, id := range ids {
		artifact, err := core.ArtifactFind(id)
		require.Nil(t, err)
		require.NotNil(t, artifact)
		assert.Equal(t, id, artifact.ID)
	}

	// Test ArtifactDelete()
	for _, id := range ids {
		err := core.ArtifactDelete(id)
		require.Nil(t, err)
	}
}

func TestFindConflictingUUID(t *testing.T) {
	defer core.ClearDartTable()
	appSetting := core.NewAppSetting("Name1", "Value1")
	require.Nil(t, core.ObjSave(appSetting))

	appSetting.ID = uuid.NewString()
	err := core.ObjSave(appSetting)
	assert.Equal(t, constants.ErrUniqueConstraint, err)
}
