package util_test

import (
	"testing"

	"github.com/APTrust/dart-runner/constants"
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

type TaggedStruct struct {
	UserID   string `json:"userId"`
	Port     int    `json:"port"`
	Allows   bool   `json:"allowsUpload"`
	Ratio    float64
	Tags     []string `json:"tags"`
	Skipped  string   `json:"-"`
	internal string
}

func TestSetValueByJsonOrCaseInsensitiveName(t *testing.T) {
	s := &TaggedStruct{}

	// Go name, JSON tag name, and case-insensitive name should all work.
	require.Nil(t, util.SetStringValue(s, "UserID", "one"))
	assert.Equal(t, "one", s.UserID)
	require.Nil(t, util.SetStringValue(s, "userId", "two"))
	assert.Equal(t, "two", s.UserID)
	require.Nil(t, util.SetStringValue(s, "userid", "three"))
	assert.Equal(t, "three", s.UserID)

	require.Nil(t, util.SetIntValue(s, "port", 22))
	assert.Equal(t, 22, s.Port)
	require.Nil(t, util.SetBoolValue(s, "allowsUpload", true))
	assert.True(t, s.Allows)
	require.Nil(t, util.SetFloatValue(s, "ratio", 0.5))
	assert.Equal(t, 0.5, s.Ratio)
}

func TestSetValueErrors(t *testing.T) {
	s := &TaggedStruct{}

	err := util.SetStringValue(s, "NoSuchField", "x")
	assert.ErrorIs(t, err, constants.ErrNoSuchField)
	err = util.SetStringValue(s, "", "x")
	assert.ErrorIs(t, err, constants.ErrNoSuchField)

	// json:"-" is not a field name.
	err = util.SetStringValue(s, "-", "x")
	assert.ErrorIs(t, err, constants.ErrNoSuchField)
	assert.Empty(t, s.Skipped)

	// Unexported fields are not settable.
	err = util.SetStringValue(s, "internal", "x")
	assert.ErrorIs(t, err, constants.ErrNoSuchField)

	// Wrong types. These used to panic or return a vague error.
	err = util.SetStringValue(s, "Port", "x")
	assert.ErrorIs(t, err, constants.ErrInvalidOperation)
	err = util.SetIntValue(s, "UserID", 65)
	assert.ErrorIs(t, err, constants.ErrInvalidOperation)
	assert.Empty(t, s.UserID)

	err = util.SetStringValue(*s, "UserID", "x")
	assert.ErrorIs(t, err, constants.ErrNotAPointer)
}

func TestSetValueFromString(t *testing.T) {
	s := &TaggedStruct{}
	require.Nil(t, util.SetValueFromString(s, "userId", "homer"))
	assert.Equal(t, "homer", s.UserID)
	require.Nil(t, util.SetValueFromString(s, "port", " 2222 "))
	assert.Equal(t, 2222, s.Port)
	require.Nil(t, util.SetValueFromString(s, "allowsUpload", "yes"))
	assert.True(t, s.Allows)
	require.Nil(t, util.SetValueFromString(s, "Ratio", "1.25"))
	assert.Equal(t, 1.25, s.Ratio)

	assert.NotNil(t, util.SetValueFromString(s, "port", "not a number"))
	assert.NotNil(t, util.SetValueFromString(s, "allowsUpload", "maybe"))
	assert.ErrorIs(t, util.SetValueFromString(s, "tags", "a,b"), constants.ErrInvalidOperation)
	assert.ErrorIs(t, util.SetValueFromString(s, "bogus", "a"), constants.ErrNoSuchField)
}
