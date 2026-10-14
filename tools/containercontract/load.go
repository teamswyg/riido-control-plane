package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func loadContract(path string) (imageContract, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return imageContract{}, err
	}
	var contract imageContract
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&contract); err != nil {
		return imageContract{}, fmt.Errorf("decode container contract: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		return imageContract{}, errors.New("decode container contract: trailing data")
	}
	if err := validateContract(contract); err != nil {
		return imageContract{}, err
	}
	return contract, nil
}
