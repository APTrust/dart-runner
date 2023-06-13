package server

import "fmt"

// Dict returns an interface map suitable for passing into
// sub templates.
func Dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("invalid parameter length: should be an even number")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("wrong data type: key '%v' should be a string", values[i])
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}
