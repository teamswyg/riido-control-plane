package riidoaiserver

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"maps"
	"net/http"
	"testing"
)

func transferEvidenceBody(t *testing.T, f *cliTransferFixture) map[string]string {
	t.Helper()
	data, _ := json.Marshal(f.wire)
	var body map[string]string
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatal(err)
	}
	body["schema_version"] = cliInstallationTransferEvidenceSchema
	return body
}

func assertTransferEvidenceRead(t *testing.T, f *cliTransferFixture, body map[string]string, found bool) {
	t.Helper()
	f.dynamo.mu.Lock()
	before := maps.Clone(f.dynamo.requests)
	f.dynamo.mu.Unlock()
	w := f.post(t, installationEvidencePath, body, "service", "")
	var got cliInstallationIntentEvidence
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &got) != nil {
		t.Fatalf("transfer evidence status=%d", w.Code)
	}
	if got.SchemaVersion != cliInstallationTransferEvidenceSchema || got.RequestID != body["request_id"] || got.Found != found || (got.Intent != nil) != found {
		t.Fatal("transfer evidence lost schema, identity or presence")
	}
	if found && *got.Intent != f.intent {
		t.Fatal("transfer evidence changed the original response")
	}
	f.dynamo.mu.Lock()
	defer f.dynamo.mu.Unlock()
	for operation, count := range f.dynamo.requests {
		if operation != dynamoDBGetItemTarget && before[operation] != count {
			t.Fatalf("read invoked %s", operation)
		}
	}
}

func TestCLIInstallationTransferEvidenceReadsWithoutCreating(t *testing.T) {
	f := newCLITransferFixture(t)
	body := transferEvidenceBody(t, f)
	missing := maps.Clone(body)
	missing["request_id"] = base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{8}, 32))
	assertTransferEvidenceRead(t, f, missing, false)
	assertTransferEvidenceRead(t, f, body, true)
	f.assertOriginal(t)
}
