package deploypolicy

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"
)

func mustRead(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(body)
}

func decodeStrictJSONDocument(t *testing.T, name, body string, dst any) {
	t.Helper()

	decoder := json.NewDecoder(bytes.NewReader([]byte(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}

	var trailing struct{}
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		t.Fatalf("%s must contain exactly one JSON document", name)
	}
}

func loadRuntimeCDOwnership(t *testing.T) runtimeCDOwnership {
	t.Helper()
	body := mustRead(t, "../../docs/30-architecture/runtime-cd-ownership.riido.json")
	var parsed runtimeCDOwnership
	decodeStrictJSONDocument(t, "runtime CD ownership manifest", body, &parsed)
	return parsed
}
