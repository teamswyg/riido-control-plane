package authpep

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

func decodeExactJWTJSON(data []byte, target any) error {
	if err := rejectDuplicateTopLevelJWTKeys(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("JWT JSON contains trailing data")
	}
	return nil
}

func rejectDuplicateTopLevelJWTKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		return errors.New("JWT JSON must be an object")
	}
	seen := map[string]struct{}{}
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return errors.New("JWT JSON key is invalid")
		}
		if _, duplicated := seen[key]; duplicated {
			return errors.New("JWT JSON key is duplicated")
		}
		seen[key] = struct{}{}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return errors.New("JWT JSON value is invalid")
		}
	}
	closing, err := decoder.Token()
	if err != nil || closing != json.Delim('}') {
		return errors.New("JWT JSON object is invalid")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("JWT JSON contains trailing data")
	}
	return nil
}
