package util_test

import (
	"fmt"
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDict(t *testing.T) {
	expected := map[string]interface{}{
		"key1": 1,
		"key2": "two",
	}
	dict, err := util.Dict("key1", 1, "key2", "two")
	require.Nil(t, err)
	assert.Equal(t, expected, dict)

	_, err = util.Dict("key1", 1, "key2")
	assert.Error(t, fmt.Errorf("invalid parameter length: should be an even number"), err)

	_, err = util.Dict(1, "key2")
	assert.Error(t, fmt.Errorf("wrong data type: key '1' should be a string"), err)
}
