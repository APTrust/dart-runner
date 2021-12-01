package util_test

import (
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
)

func TestGetHashes(t *testing.T) {
	algs := constants.PreferredAlgsInOrder
	digests := util.GetHashes(algs)
	assert.Equal(t, len(algs), len(digests))
}
