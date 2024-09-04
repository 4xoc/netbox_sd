// MIT License
//
// Copyright (c) 2024 WIIT AG
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
// documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the
// Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
// WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
// OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package netbox

import (
	"encoding/json"
	"errors"
)

// Possible custom field value types.
const (
	CustomFieldText   string = "text"
	CustomFieldNumber string = "integer"
	CustomFieldBool   string = "boolean"
)

// Possible errors returned when working with custom fields.
var (
	ErrCFMUnsupportedDataType = errors.New("custom field data type not supported")
	ErrCFCantConvertValue     = errors.New("custom field value cannot be converted to destination type")
)

// CustomField describes a single Netbox custom field's data. Fields are exported to allow for mocking.
type CustomField struct {
	Datatype string
	Value    interface{}
}

// CFMap implements the CustomFieldMap interface.
type CFMap struct {
	entries map[string]*CustomField
}

// UnmarshalJSON implements a custom JSON unmarshal interface for CFMap (and therefore CustomFieldMap).
func (cfm *CFMap) UnmarshalJSON(b []byte) error {
	var (
		tmp map[string]interface{} = make(map[string]interface{})
		err error
		cf  *CustomField
		key string
		val interface{}
	)

	// Convert JSON blob into some data type we can easily work with.
	if err = json.Unmarshal(b, &tmp); err != nil {
		return err
	}

	cfm.entries = make(map[string]*CustomField)

	for key, val = range tmp {

		if val == nil {
			// ignore any custom field that has no value
			continue
		}

		cf = new(CustomField)

		switch val.(type) {
		// JSON package always uses float64 for JSON numbers so we keep it that way.
		case float64:
			cf.Datatype = CustomFieldNumber
			cf.Value = val

		case string:
			cf.Datatype = CustomFieldText
			cf.Value = val

		case bool:
			cf.Datatype = CustomFieldBool
			cf.Value = val

		default:
			return ErrCFMUnsupportedDataType
		}

		// Adding entry to map.
		cfm.entries[key] = cf
	}

	return nil
}

// GetEntry implements CustomFieldMap.GetEntry.
func (cfm CFMap) GetEntry(name string) *CustomField {
	var (
		val *CustomField
		ok  bool
	)

	if val, ok = cfm.entries[name]; !ok {
		return nil
	}

	return val
}

// GetAllEntries implements CustomFieldMap.GetAllEntries.
func (cfm CFMap) GetAllEntries(callback func(string, *CustomField)) {
	var key string

	for key = range cfm.entries {
		callback(key, cfm.entries[key])
	}
}

// AsString takes a given CustomField and tries to returns it's value as string. If the underlying datatype doesn't
// support being returned as string, an error is returned.
func (cf *CustomField) AsString() (string, error) {

	if cf.Datatype != CustomFieldText {
		return "", ErrCFCantConvertValue
	}

	return cf.Value.(string), nil
}

// AsFloat takes a given CustomField and tries to returns it's value as int64. If the underlying datatype doesn't
// support being returned as float64, an error is returned.
func (cf *CustomField) AsFloat() (float64, error) {

	if cf.Datatype != CustomFieldNumber {
		return 0, ErrCFCantConvertValue
	}

	return cf.Value.(float64), nil
}

// AsBool takes a given CustomField and tries to returns it's value as bool. If the underlying datatype doesn't
// support being returned as bool, an error is returned.
func (cf *CustomField) AsBool() (bool, error) {

	if cf.Datatype != CustomFieldBool {
		return false, ErrCFCantConvertValue
	}

	return cf.Value.(bool), nil
}
