package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"
)

func loadFigmaProjectionManifest(t *testing.T, path string) figmaProjectionManifest {
	t.Helper()
	var manifest figmaProjectionManifest
	decodeStrictJSON(t, path, &manifest)
	return manifest
}

func loadFigmaSourceCoverageManifest(t *testing.T, path string) figmaSourceCoverageManifest {
	t.Helper()
	var manifest figmaSourceCoverageManifest
	decodeStrictJSON(t, path, &manifest)
	return manifest
}

func decodeStrictJSON(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	var trailing struct{}
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		t.Fatalf("decode %s: trailing JSON document: %v", path, err)
	}
}
