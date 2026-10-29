package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func decodeCLIInstallation(w http.ResponseWriter, r *http.Request, keys string, target any) (err error) {
	defer func() {
		var large *http.MaxBytesError
		if err != nil && !errors.As(err, &large) {
			err = errCLIInstallationInvalid
		}
	}()
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	defer r.Body.Close()
	d := json.NewDecoder(r.Body)
	first, err := d.Token()
	if err != nil {
		return err
	}
	if first != json.Delim('{') {
		return errCLIInstallationInvalid
	}
	allowed := strings.Fields(keys)
	values := map[string]json.RawMessage{}
	for d.More() {
		token, err := d.Token()
		if err != nil {
			return err
		}
		key, ok := token.(string)
		if !ok || values[key] != nil || !installationKeyAllowed(allowed, key) {
			return errCLIInstallationInvalid
		}
		var value json.RawMessage
		if err = d.Decode(&value); err != nil {
			return err
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return errCLIInstallationInvalid
		}
		values[key] = value
	}
	last, err := d.Token()
	if err != nil {
		return err
	}
	if last != json.Delim('}') || len(values) != len(allowed) {
		return errCLIInstallationInvalid
	}
	if _, err = d.Token(); !errors.Is(err, io.EOF) {
		return errCLIInstallationInvalid
	}
	data, err := json.Marshal(values)
	if err != nil || json.Unmarshal(data, target) != nil {
		return errCLIInstallationInvalid
	}
	return nil
}

func installationKeyAllowed(keys []string, key string) bool {
	for _, candidate := range keys {
		if key == candidate {
			return true
		}
	}
	return false
}
