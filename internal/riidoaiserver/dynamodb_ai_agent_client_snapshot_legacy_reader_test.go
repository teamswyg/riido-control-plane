package riidoaiserver

import (
	"io"
	"strings"
	"testing"
)

func TestLegacyDynamoDBAIAgentClientSnapshotReaderPrefersGzip(t *testing.T) {
	gzipped, err := gzipBase64([]byte(`{"schema_version":"gzip"}`))
	if err != nil {
		t.Fatalf("gzipBase64: %v", err)
	}
	item := legacySnapshotItem(gzipped, `{"schema_version":"json"}`)

	r, size, err := legacyDynamoDBAIAgentClientSnapshotReader(item)
	if err != nil {
		t.Fatalf("legacy reader: %v", err)
	}
	body, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if string(body) != `{"schema_version":"gzip"}` || size != int64(len(gzipped)) {
		t.Fatalf("body=%s size=%d, want gzip body and compressed size", body, size)
	}
}

func TestLegacyDynamoDBAIAgentClientSnapshotReaderReadsJSONFallback(t *testing.T) {
	raw := `{"schema_version":"json"}`
	r, size, err := legacyDynamoDBAIAgentClientSnapshotReader(legacySnapshotItem("", raw))
	if err != nil {
		t.Fatalf("legacy reader: %v", err)
	}
	body, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if string(body) != raw || size != int64(len(raw)) {
		t.Fatalf("body=%s size=%d, want raw JSON and size", body, size)
	}
}

func TestLegacyDynamoDBAIAgentClientSnapshotReaderRejectsInvalidGzip(t *testing.T) {
	_, _, err := legacyDynamoDBAIAgentClientSnapshotReader(legacySnapshotItem("not-base64", ""))
	if err == nil || !strings.Contains(err.Error(), "decode DynamoDB AI Agent client snapshot gzip") {
		t.Fatalf("err=%v, want gzip decode error", err)
	}
}

func TestLegacyDynamoDBAIAgentClientSnapshotReaderRequiresSnapshot(t *testing.T) {
	_, _, err := legacyDynamoDBAIAgentClientSnapshotReader(nil)
	if err == nil || !strings.Contains(err.Error(), "snapshot_gzip or snapshot_json is required") {
		t.Fatalf("err=%v, want missing snapshot error", err)
	}
}

func legacySnapshotItem(gzipValue, jsonValue string) map[string]map[string]string {
	item := map[string]map[string]string{}
	if gzipValue != "" {
		item["snapshot_gzip"] = map[string]string{"S": gzipValue}
	}
	if jsonValue != "" {
		item["snapshot_json"] = map[string]string{"S": jsonValue}
	}
	return item
}
