package util_test

import (
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNameValuePairList(t *testing.T) {
	list := util.NewNameValuePairList()
	require.NotNil(t, list.Items)
	assert.Empty(t, list.Items)

	list.Add("One", "First")
	list.Add("Two", "Second")
	list.Add("MultiItem", "One")
	list.Add("MultiItem", "Two")
	list.Add("MultiItem", "Three")

	assert.Equal(t, 5, len(list.Items))

	pair, found := list.FirstMatching("Does Not Exist")
	assert.False(t, found)
	assert.Empty(t, pair)

	pairs := list.AllMatching("Does Not Exist")
	assert.Empty(t, pairs)

	one, found := list.FirstMatching("One")
	assert.True(t, found)
	assert.Equal(t, "One", one.Name)
	assert.Equal(t, "First", one.Value)

	matches := list.AllMatching("One")
	require.Equal(t, 1, len(matches))
	assert.Equal(t, "One", matches[0].Name)
	assert.Equal(t, "First", matches[0].Value)

	matches = list.AllMatching("MultiItem")
	require.Equal(t, 3, len(matches))
	assert.Equal(t, "One", matches[0].Value)
	assert.Equal(t, "Two", matches[1].Value)
	assert.Equal(t, "Three", matches[2].Value)
}
