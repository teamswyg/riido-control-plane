package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

func loadContract(path string) (contract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return contract{}, fmt.Errorf("read contract: %w", err)
	}
	var c contract
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&c); err != nil {
		return contract{}, fmt.Errorf("decode contract: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return contract{}, errors.New("decode contract: trailing JSON data")
	} else if !errors.Is(err, io.EOF) {
		return contract{}, fmt.Errorf("decode contract trailer: %w", err)
	}
	return c, verifyContract(c)
}
