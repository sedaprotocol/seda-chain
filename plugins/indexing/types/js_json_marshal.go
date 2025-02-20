package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// MarshalJSJSON marshals a struct to a JSON object that can be used in JavaScript.
// It converts all int64 and uint64 values to strings as these exceed the safe integer
// limit for JavaScript.
func MarshalJSJSON(p interface{}) ([]byte, error) {
	fields := make(map[string]interface{})

	// Use reflection to iterate through struct fields
	v := reflect.ValueOf(p)
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Get the json tag name, or use field name if no tag
		jsonTag := field.Tag.Get("json")
		name := strings.Split(jsonTag, ",")[0]
		if name == "" {
			name = field.Name
		}

		// Convert specific types to strings
		switch value.Kind() {
		case reflect.Int64:
			fields[name] = fmt.Sprintf("%d", value.Int())
		case reflect.Uint64:
			fields[name] = fmt.Sprintf("%d", value.Uint())
		default:
			fields[name] = value.Interface()
		}
	}

	return json.Marshal(fields)
}
