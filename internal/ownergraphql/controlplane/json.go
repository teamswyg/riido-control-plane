package controlplane

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

func rejectControlPlaneOwnerAmbiguousJSON(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := scanControlPlaneOwnerJSON(decoder, 0); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return errors.New("trailing JSON data")
	}
	return nil
}

func scanControlPlaneOwnerJSON(decoder *json.Decoder, depth int) error {
	if depth > 16 {
		return errors.New("JSON nesting is too deep")
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			keyToken, keyErr := decoder.Token()
			key, keyOK := keyToken.(string)
			if keyErr != nil || !keyOK || seen[key] {
				return errors.New("duplicate or invalid JSON key")
			}
			seen[key] = true
			if err := scanControlPlaneOwnerJSON(decoder, depth+1); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := scanControlPlaneOwnerJSON(decoder, depth+1); err != nil {
				return err
			}
		}
	default:
		return errors.New("invalid JSON delimiter")
	}
	_, err = decoder.Token()
	return err
}
