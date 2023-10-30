package core_test

import (
	"fmt"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExportQuestion(t *testing.T) {
	q := core.NewExportQuestion()
	require.NotNil(t, q) // TODO: List object fields or, for BagIt profiles, list tag names

	assert.True(t, util.LooksLikeUUID(q.ID))
}

func TestExportQuestionToForm(t *testing.T) {
	defer core.ClearDartTable()
	for i := 0; i < 5; i++ {
		ss := core.NewStorageService()
		ss.Name = fmt.Sprintf("Storage Service %d", i)
		require.NoError(t, core.ObjSaveWithoutValidation(ss))
	}

	q := core.NewExportQuestion()
	q.Prompt = "Whassup, Chuck?"
	q.ObjType = constants.TypeStorageService
	q.ObjID = constants.EmptyUUID
	q.Field = "UserID"

	form := q.ToForm()
	assert.Equal(t, q.ID, form.Fields["ID"].Value)
	assert.Equal(t, q.Prompt, form.Fields["Prompt"].Value)
	assert.Equal(t, q.ObjType, form.Fields["ObjType"].Value)
	assert.Equal(t, 5, len(form.Fields["ObjType"].Choices))
	assert.Equal(t, q.ObjID, form.Fields["ObjID"].Value)
	assert.Equal(t, 5, len(form.Fields["ObjID"].Choices))
	assert.Equal(t, q.Field, form.Fields["Field"].Value)
}
