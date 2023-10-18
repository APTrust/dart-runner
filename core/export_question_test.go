package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExportQuestion(t *testing.T) {
	q := core.NewExportQuestion()
	require.NotNil(t, q)
	assert.True(t, util.LooksLikeUUID(q.ID))
}
