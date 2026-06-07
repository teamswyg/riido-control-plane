package riidoaiserver

import (
	"testing"
)

func TestGzipBase64RoundTrip(t *testing.T) {
	original := []byte(`{"schema_version":"x","devices":[{"device_id":"dev_1"}],"events":["a","a","a","a"]}`)
	encoded, err := gzipBase64(original)
	if err != nil {
		t.Fatalf("gzipBase64: %v", err)
	}
	if encoded == "" {
		t.Fatal("gzipBase64 returned empty")
	}
	decoded, err := gunzipBase64(encoded)
	if err != nil {
		t.Fatalf("gunzipBase64: %v", err)
	}
	if string(decoded) != string(original) {
		t.Fatalf("round trip mismatch: got %q want %q", decoded, original)
	}
}
