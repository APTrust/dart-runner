package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
)

func TestNewField(t *testing.T) {
	f := core.NewField("field_id", "Field1", "Label1", "Value1", false)
	assert.Equal(t, "field_id", f.ID)
	assert.Equal(t, "Field1", f.Name)
	assert.Equal(t, "Label1", f.Label)
	assert.Equal(t, "Value1", f.Value)
	assert.False(t, f.Required)

	assert.NotNil(t, f.Choices)
	assert.NotNil(t, f.CssClasses)
	assert.NotNil(t, f.Attrs)

	f = core.NewField("field_id", "Field1", "Label1", "Value1", true)
	assert.True(t, f.Required)
}

func TestAddChoice(t *testing.T) {
	f := core.NewField("field_id", "Field1", "Label1", "Value1", false)
	f.AddChoice("", "")
	f.AddChoice("Option1", "Value1")
	f.AddChoice("Option2", "Value2")

	assert.Equal(t, 3, len(f.Choices))
	assert.False(t, f.Choices[0].Selected)
	assert.True(t, f.Choices[1].Selected)
	assert.False(t, f.Choices[2].Selected)
}

func TestNewForm(t *testing.T) {
	f := core.NewForm(constants.TypeAppSetting, constants.EmptyUUID, nil)
	assert.Equal(t, constants.TypeAppSetting, f.ObjType)
	assert.Equal(t, constants.EmptyUUID, f.ObjectID)
	assert.NotNil(t, f.Errors)
	assert.NotNil(t, f.Fields)
}

func TestAddField(t *testing.T) {
	errors := map[string]string{
		"Field1": "error one",
		"Field3": "error three",
	}
	f := core.NewForm(constants.TypeAppSetting, constants.EmptyUUID, errors)
	f1 := f.AddField("Field1", "Field One", "Value One", true)
	f2 := f.AddField("Field2", "Field Two", "Value Two", false)
	f3 := f.AddField("Field3", "Field Three", "Value Three", true)

	assert.Equal(t, f1, f.Fields["Field1"])
	assert.Equal(t, "AppSetting_Field1", f1.ID)
	assert.Equal(t, "Field One", f1.Label)
	assert.Equal(t, "Field1", f1.Name)
	assert.Equal(t, "Value One", f1.Value)
	assert.Equal(t, "Value One", f1.Values[0])
	assert.Equal(t, "error one", f1.Error)
	assert.True(t, f1.Required)

	assert.Empty(t, f2.Error)
	assert.Equal(t, "error three", f3.Error)
}

func TestAddMultiValueField(t *testing.T) {
	errors := map[string]string{
		"Field1": "error one",
	}
	values := []string{
		"value1",
		"value2",
		"value3",
	}
	f := core.NewForm(constants.TypeAppSetting, constants.EmptyUUID, errors)
	f1 := f.AddMultiValueField("Field1", "Field One", values, true)

	assert.Equal(t, f1, f.Fields["Field1"])
	assert.Equal(t, "AppSetting_Field1", f1.ID)
	assert.Equal(t, "Field One", f1.Label)
	assert.Equal(t, "Field1", f1.Name)
	assert.Empty(t, f1.Value)
	assert.Equal(t, values, f1.Values)
	assert.Equal(t, "error one", f1.Error)
	assert.True(t, f1.Required)
}

func TestFormToString(t *testing.T) {
	f := core.NewForm(constants.TypeAppSetting, constants.EmptyUUID, nil)
	assert.Equal(t, "Form: AppSetting id=00000000-0000-0000-0000-000000000000", f.String())
}
