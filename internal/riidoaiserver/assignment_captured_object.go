package riidoaiserver

import (
	"encoding/json"
	"errors"
)

func capturedJSONObject(decoder *json.Decoder, depth int) (any, error) {
	values := map[string]any{}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := token.(string)
		if !ok {
			return nil, errors.New("invalid captured JSON key")
		}
		if _, exists := values[key]; exists {
			return nil, errors.New("duplicate captured JSON key")
		}
		value, err := capturedJSONValue(decoder, depth)
		if err != nil {
			return nil, err
		}
		values[key] = value
	}
	if end, err := decoder.Token(); err != nil || end != json.Delim('}') {
		return nil, errors.New("invalid captured JSON object")
	}
	return values, nil
}
