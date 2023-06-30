package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYesNoChoices(t *testing.T) {
	choices := core.YesNoChoices(true)
	assert.Equal(t, 3, len(choices))
	assert.Equal(t, "Yes", choices[1].Label)
	assert.Equal(t, "true", choices[1].Value)
	assert.True(t, choices[1].Selected)
	assert.Equal(t, "No", choices[2].Label)
	assert.Equal(t, "false", choices[2].Value)
	assert.False(t, choices[2].Selected)

	choices = core.YesNoChoices(false)
	assert.False(t, choices[1].Selected)
	assert.True(t, choices[2].Selected)
}

func TestMakeChoiceList(t *testing.T) {
	choices := core.MakeChoiceList(constants.PreferredAlgsInOrder, constants.AlgSha256)
	require.Equal(t, len(constants.PreferredAlgsInOrder)+1, len(choices))
	assert.Equal(t, "", choices[0].Label)
	assert.Equal(t, "", choices[0].Value)
	for i, alg := range constants.PreferredAlgsInOrder {
		choice := choices[i+1]
		assert.Equal(t, constants.PreferredAlgsInOrder[i], choice.Label)
		assert.Equal(t, constants.PreferredAlgsInOrder[i], choice.Value)
		if alg == constants.AlgSha256 {
			assert.True(t, choice.Selected)
		} else {
			assert.False(t, choice.Selected)
		}
	}
}

func TestMakeMultiChoiceList(t *testing.T) {
	selected := []string{constants.AlgSha256, constants.AlgSha1}
	choices := core.MakeMultiChoiceList(constants.PreferredAlgsInOrder, selected)
	require.Equal(t, len(constants.PreferredAlgsInOrder)+1, len(choices))
	assert.Equal(t, "", choices[0].Label)
	assert.Equal(t, "", choices[0].Value)
	for i, alg := range constants.PreferredAlgsInOrder {
		choice := choices[i+1]
		assert.Equal(t, constants.PreferredAlgsInOrder[i], choice.Label)
		assert.Equal(t, constants.PreferredAlgsInOrder[i], choice.Value)
		if alg == constants.AlgSha256 || alg == constants.AlgSha1 {
			assert.True(t, choice.Selected)
		} else {
			assert.False(t, choice.Selected)
		}
	}
}
