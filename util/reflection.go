package util

import (
	"reflect"

	"github.com/APTrust/dart-runner/constants"
)

// SetStringValue sets the value of obj.fieldName to value.
// Returns an error if the field does not exist, or is the wrong
// type, or cannot be set.
func SetStringValue(obj interface{}, fieldName, value string) error {
	field, err := getField(obj, fieldName, reflect.TypeOf(value))
	if err != nil {
		return err
	}
	field.SetString(value)
	return nil
}

// SetIntValue sets the value of obj.fieldName to value.
// Returns an error if the field does not exist, or is the wrong
// type, or cannot be set.
func SetIntValue(obj interface{}, fieldName string, value int64) error {
	field, err := getField(obj, fieldName, reflect.TypeOf(value))
	if err != nil {
		return err
	}
	field.SetInt(value)
	return nil
}

// SetFloatValue sets the value of obj.fieldName to value.
// Returns an error if the field does not exist, or is the wrong
// type, or cannot be set.
func SetFloatValue(obj interface{}, fieldName string, value float64) error {
	field, err := getField(obj, fieldName, reflect.TypeOf(value))
	if err != nil {
		return err
	}
	field.SetFloat(value)
	return nil
}

// SetBoolValue sets the value of obj.fieldName to value.
// Returns an error if the field does not exist, or is the wrong
// type, or cannot be set.
func SetBoolValue(obj interface{}, fieldName string, value bool) error {
	field, err := getField(obj, fieldName, reflect.TypeOf(value))
	if err != nil {
		return err
	}
	field.SetBool(value)
	return nil
}

func getField(obj interface{}, fieldName string, requiredType reflect.Type) (reflect.Value, error) {
	var field reflect.Value
	if reflect.ValueOf(obj).Kind() != reflect.Ptr {
		return field, constants.ErrNotAPointer
	}
	objValue := reflect.ValueOf(obj).Elem()
	field = objValue.FieldByName(fieldName)
	if !field.CanSet() || field.Type() != requiredType {
		return field, constants.ErrInvalidOperation
	}
	return field, nil
}
