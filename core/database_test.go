package core_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleArtifact() *core.Artifact {
	return &core.Artifact{
		ID:       uuid.NewString(),
		JobID:    constants.EmptyUUID,
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
		// We need a delay here, or timestamps won't
		// sort correctly in the queries below.
		time.Sleep(50 * time.Millisecond)
		require.Nil(t, err)
		ids[i] = artifact.ID
	}

	// Note that when we retrieve artifacts, they're ordered
	// by updated_at desc, name asc.

	// Test ArtifactListByJobName()
	artifacts, err := core.ArtifactListByJobName("Bag 1")
	require.Nil(t, err)
	assert.Equal(t, 3, len(artifacts))
	assert.Equal(t, "File 4", artifacts[0].FileName)
	assert.Equal(t, "File 2", artifacts[1].FileName)
	assert.Equal(t, "File 0", artifacts[2].FileName)

	artifacts, err = core.ArtifactListByJobName("Bag 2")
	require.Nil(t, err)
	assert.Equal(t, 2, len(artifacts))
	assert.Equal(t, "File 3", artifacts[0].FileName)
	assert.Equal(t, "File 1", artifacts[1].FileName)

	// Test ArtifactListByJobId()
	artifacts, err = core.ArtifactListByJobID(constants.EmptyUUID)
	require.Nil(t, err)
	assert.Equal(t, 5, len(artifacts))
	assert.Equal(t, "File 4", artifacts[0].FileName)
	assert.Equal(t, "File 3", artifacts[1].FileName)
	assert.Equal(t, "File 2", artifacts[2].FileName)
	assert.Equal(t, "File 1", artifacts[3].FileName)
	assert.Equal(t, "File 0", artifacts[4].FileName)

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

func TestConflictingObjectName(t *testing.T) {
	defer core.ClearDartTable()
	appSetting := core.NewAppSetting("Name1", "Value1")
	require.Nil(t, core.ObjSave(appSetting))

	appSetting2 := core.NewAppSetting("Name1", "Value2")
	err := core.ObjSave(appSetting2)
	assert.Equal(t, constants.ErrUniqueConstraint, err)
}

func createAppSettings(t *testing.T, count int) []string {
	objIds := make([]string, count)
	for i := 0; i < count; i++ {
		setting := core.NewAppSetting(fmt.Sprintf("Setting #%d", i), fmt.Sprintf("Value %d", i))
		assert.Nil(t, core.ObjSave(setting))
		objIds[i] = setting.ID
	}
	return objIds
}

func TestObjNameIDAndChoiceList(t *testing.T) {
	defer core.ClearDartTable()
	ids := createAppSettings(t, 10)

	nameIdList := core.ObjNameIdList(constants.TypeAppSetting)
	assert.Equal(t, 10, len(nameIdList))
	for i, id := range ids {
		name := fmt.Sprintf("Setting #%d", i)
		assert.Equal(t, id, nameIdList[i].ID)
		assert.Equal(t, name, nameIdList[i].Name)
	}

	// This should return nothing
	nameIdList = core.ObjNameIdList(constants.TypeRemoteRepository)
	assert.Empty(t, nameIdList)

	// Choice list should contain the same data as name-id list,
	// in the same order. The following two choices should be marked
	// as selected.
	selected := []string{
		ids[2],
		ids[3],
	}
	choiceList := core.ObjChoiceList(constants.TypeAppSetting, selected)
	for i, id := range ids {
		name := fmt.Sprintf("Setting #%d", i)
		assert.Equal(t, id, choiceList[i].Value)
		assert.Equal(t, name, choiceList[i].Label)
		assert.Equal(t, util.StringListContains(selected, id), choiceList[i].Selected)
	}

	// These should return nothing
	nameIdList = core.ObjNameIdList(constants.TypeRemoteRepository)
	assert.Empty(t, nameIdList)
	choiceList = core.ObjChoiceList(constants.TypeRemoteRepository, selected)
	assert.Empty(t, choiceList)

}

func TestGetAppSetting(t *testing.T) {
	defer core.ClearDartTable()
	value, err := core.GetAppSetting("Setting does not exist")
	assert.Error(t, err)
	assert.Empty(t, value)

	setting := core.NewAppSetting("Test Setting", "Test Value")
	require.NoError(t, core.ObjSave(setting))

	value, err = core.GetAppSetting("Test Setting")
	assert.NoError(t, err)
	assert.Equal(t, "Test Value", value)
}

func TestArtifactNameIDList(t *testing.T) {
	defer core.ClearArtifactsTable()
	for i := 0; i < 5; i++ {
		artifact := core.NewManifestArtifact(
			"test-bag.tar",
			constants.EmptyUUID,
			fmt.Sprintf("manifest-%d.txt", i),
			fmt.Sprintf("Contents of manifest %d ...", i))
		require.NoError(t, core.ArtifactSave(artifact))
	}
	nameIDList, err := core.ArtifactNameIDList(constants.EmptyUUID)
	require.Nil(t, err)
	assert.Equal(t, 5, len(nameIDList))

	// If no artifacts, we should get empty list.
	nameIDList, err = core.ArtifactNameIDList(uuid.NewString())
	require.Nil(t, err)
	assert.Equal(t, 0, len(nameIDList))

	testArtifactsDeleteByJobID(t, constants.EmptyUUID)
}

func testArtifactsDeleteByJobID(t *testing.T, jobID string) {
	require.NoError(t, core.ArtifactsDeleteByJobID(jobID))
	nameIDList, err := core.ArtifactNameIDList(jobID)
	require.Nil(t, err)
	assert.Empty(t, nameIDList)
}
