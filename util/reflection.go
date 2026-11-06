package util

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/APTrust/dart-runner/constants"
)

// SetStringValue sets the value of obj.fieldName to value.
// Returns an error if the field does not exist, or is the wrong
// type, or cannot be set.
//
// fieldName may be the Go field name ("UserID"), the field's JSON
// tag name ("userId"), or a case-insensitive match of the Go name.
// This allows us to handle settings exported from DART 2, which
// uses JavaScript-style (JSON) field names.
func SetStringValue(obj interface{}, fieldName, value string) error {
	field, err := getField(obj, fieldName, reflect.String)
	if err != nil {
		return err
	}
	field.SetString(value)
	return nil
}

// SetIntValue sets the value of obj.fieldName to value.
// Returns an error if the field does not exist, or is the wrong
// type, or cannot be set. See SetStringValue for acceptable
// forms of fieldName.
func SetIntValue(obj interface{}, fieldName string, value int64) error {
	field, err := getField(obj, fieldName, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64)
	if err != nil {
		return err
	}
	if field.OverflowInt(value) {
		return fmt.Errorf("%w: value %d overflows field %s", constants.ErrInvalidOperation, value, fieldName)
	}
	field.SetInt(value)
	return nil
}

// SetFloatValue sets the value of obj.fieldName to value.
// Returns an error if the field does not exist, or is the wrong
// type, or cannot be set. See SetStringValue for acceptable
// forms of fieldName.
func SetFloatValue(obj interface{}, fieldName string, value float64) error {
	field, err := getField(obj, fieldName, reflect.Float32, reflect.Float64)
	if err != nil {
		return err
	}
	field.SetFloat(value)
	return nil
}

// SetBoolValue sets the value of obj.fieldName to value.
// Returns an error if the field does not exist, or is the wrong
// type, or cannot be set. See SetStringValue for acceptable
// forms of fieldName.
func SetBoolValue(obj interface{}, fieldName string, value bool) error {
	field, err := getField(obj, fieldName, reflect.Bool)
	if err != nil {
		return err
	}
	field.SetBool(value)
	return nil
}

// SetValueFromString converts value to the type of obj.fieldName
// and then sets the field. It supports string, bool, int and float
// fields. Bool values may be any value accepted by StringToBool
// (true/false, yes/no, y/n, 1/0). See SetStringValue for acceptable
// forms of fieldName.
//
// This is useful when the value comes from a user response or
// a form, where everything arrives as a string.
func SetValueFromString(obj interface{}, fieldName, value string) error {
	field, err := getField(obj, fieldName)
	if err != nil {
		return err
	}
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		boolValue, err := StringToBool(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("field %s requires a true/false value: %w", fieldName, err)
		}
		field.SetBool(boolValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil {
			return fmt.Errorf("field %s requires a whole number: %w", fieldName, err)
		}
		if field.OverflowInt(intValue) {
			return fmt.Errorf("%w: value %d overflows field %s", constants.ErrInvalidOperation, intValue, fieldName)
		}
		field.SetInt(intValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return fmt.Errorf("field %s requires a number: %w", fieldName, err)
		}
		field.SetFloat(floatValue)
	default:
		return fmt.Errorf("%w: cannot set field %s of type %s from a string", constants.ErrInvalidOperation, fieldName, field.Type())
	}
	return nil
}

// getField returns the settable field fieldName of the struct that
// obj points to. If allowedKinds is not empty, the field's kind must
// be one of them.
func getField(obj interface{}, fieldName string, allowedKinds ...reflect.Kind) (reflect.Value, error) {
	var field reflect.Value
	ptr := reflect.ValueOf(obj)
	if ptr.Kind() != reflect.Ptr || ptr.IsNil() {
		return field, constants.ErrNotAPointer
	}
	objValue := ptr.Elem()
	if objValue.Kind() != reflect.Struct {
		return field, fmt.Errorf("%w: %s is not a struct", constants.ErrInvalidOperation, objValue.Type())
	}
	goName := resolveFieldName(objValue.Type(), fieldName)
	if goName == "" {
		return field, fmt.Errorf("%w: %s has no field %q", constants.ErrNoSuchField, objValue.Type(), fieldName)
	}
	field = objValue.FieldByName(goName)
	if !field.CanSet() {
		return field, fmt.Errorf("%w: field %s.%s cannot be set", constants.ErrInvalidOperation, objValue.Type(), goName)
	}
	if len(allowedKinds) > 0 {
		kindOk := false
		for _, kind := range allowedKinds {
			if field.Kind() == kind {
				kindOk = true
				break
			}
		}
		if !kindOk {
			return field, fmt.Errorf("%w: field %s.%s is type %s", constants.ErrInvalidOperation, objValue.Type(), goName, field.Type())
		}
	}
	return field, nil
}

// resolveFieldName returns the Go name of the field in structType
// that matches name. It checks, in order, for an exact match on the
// Go field name, an exact match on the JSON tag name, and a
// case-insensitive match on either. Returns an empty string if
// there's no match.
func resolveFieldName(structType reflect.Type, name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if f, ok := structType.FieldByName(name); ok && f.IsExported() {
		return f.Name
	}
	caseInsensitiveMatch := ""
	for i := 0; i < structType.NumField(); i++ {
		f := structType.Field(i)
		if !f.IsExported() {
			continue
		}
		jsonName := strings.Split(f.Tag.Get("json"), ",")[0]
		if jsonName == name && jsonName != "-" {
			return f.Name
		}
		if caseInsensitiveMatch == "" && (strings.EqualFold(f.Name, name) || (jsonName != "" && jsonName != "-" && strings.EqualFold(jsonName, name))) {
			caseInsensitiveMatch = f.Name
		}
	}
	return caseInsensitiveMatch
}
