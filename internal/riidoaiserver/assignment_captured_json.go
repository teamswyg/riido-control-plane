package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

func canonicalCapturedJSON(raw json.RawMessage) (json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	value, err := capturedJSONValue(decoder, 0)
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, errors.New("captured JSON has trailing data")
	}
	return json.Marshal(value)
}

func capturedJSONValue(decoder *json.Decoder, depth int) (any, error) {
	if depth > 128 {
		return nil, errors.New("captured JSON is too deep")
	}
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	delimiter, compound := token.(json.Delim)
	if !compound {
		return token, nil
	}
	if delimiter == '{' {
		return capturedJSONObject(decoder, depth+1)
	}
	if delimiter != '[' {
		return nil, errors.New("invalid captured JSON delimiter")
	}
	values := []any{}
	for decoder.More() {
		value, err := capturedJSONValue(decoder, depth+1)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	if end, err := decoder.Token(); err != nil || end != json.Delim(']') {
		return nil, errors.New("invalid captured JSON array")
	}
	return values, nil
}
