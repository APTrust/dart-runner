package util_test

import (
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	StringField string
	IntField    int64
	FloatField  float64
	BoolField   bool
}

func TestSetStringValue(t *testing.T) {
	testStruct := &TestStruct{}
	assert.Empty(t, testStruct.StringField)
	err := util.SetStringValue(testStruct, "StringField", "Homer Simpson")
	require.Nil(t, err)
	assert.Equal(t, "Homer Simpson", testStruct.StringField)
}

func TestSetIntValue(t *testing.T) {
	testStruct := &TestStruct{}
	assert.Empty(t, testStruct.IntField)
	err := util.SetIntValue(testStruct, "IntField", 333)
	require.Nil(t, err)
	assert.Equal(t, int64(333), testStruct.IntField)
}

func TestSetFloatValue(t *testing.T) {
	testStruct := &TestStruct{}
	assert.Empty(t, testStruct.FloatField)
	err := util.SetFloatValue(testStruct, "FloatField", 333.97)
	require.Nil(t, err)
	assert.Equal(t, float64(333.97), testStruct.FloatField)
}

func TestSetBoolValue(t *testing.T) {
	testStruct := &TestStruct{}
	assert.False(t, testStruct.BoolField)
	err := util.SetBoolValue(testStruct, "BoolField", true)
	require.Nil(t, err)
	assert.True(t, testStruct.BoolField)
}
