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

func TestObjCountAndObjExists(t *testing.T) {
	defer core.ClearDartTable()
	count, err := core.ObjCount(constants.TypeAppSetting)
	require.Nil(t, err)
	assert.Equal(t, 0, count)

	objIds := make([]string, 5)
	for i := 0; i < 5; i++ {
		setting := core.NewAppSetting(fmt.Sprintf("Name %d", i), fmt.Sprintf("Value %d", i))
		assert.Nil(t, setting.Save())
		objIds[i] = setting.ID
	}

	count, err = core.ObjCount(constants.TypeAppSetting)
	require.Nil(t, err)
	assert.Equal(t, 5, count)

	for _, objId := range objIds {
		exists, err := core.ObjExists(objId)
		require.Nil(t, err)
		assert.True(t, exists)
	}

	// Random UUID should not exist in DB
	exists, err := core.ObjExists(uuid.NewString())
	require.Nil(t, err)
	assert.False(t, exists)
}
