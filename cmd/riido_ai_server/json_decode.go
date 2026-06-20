package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

func strictDecodeJSON(raw string, out any) error {
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		return errors.New("decode json: trailing data")
	}
	return nil
}
