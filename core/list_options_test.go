package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
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
