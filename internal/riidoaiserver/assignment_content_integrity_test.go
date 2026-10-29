package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestAssignmentContentTamperedStoredPayloadFailsClosed(t *testing.T) {
	fixture := &contentBundleFixture{}
	store := newContentBundleDynamo(t, fixture)
	snapshot, pages := assignmentContentTestBundle()
	if err := store.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err != nil {
		t.Fatal(err)
	}
	fixture.mu.Lock()
	item := fixture.items[dynamoDBAssignmentContextPK("assignment")+"|"+assignmentContentSK(0)]
	decoded, err := gunzipBase64(item["payload_gzip"]["S"])
	if err != nil {
		fixture.mu.Unlock()
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(decoded, &payload); err != nil {
		fixture.mu.Unlock()
		t.Fatal(err)
	}
	if entries, ok := payload["pages"].([]any); ok {
		entries[0].(map[string]any)["data"] = map[string]string{"content": "forged"}
	} else {
		payload["data"] = map[string]string{"content": "forged"}
	}
	encoded, _ := json.Marshal(payload)
	compressed, _ := gzipBase64(encoded)
	item["payload_gzip"] = map[string]string{"S": compressed}
	fixture.mu.Unlock()
	if _, found, err := store.LoadAssignmentContentPage(t.Context(), "assignment", 0); err == nil || found {
		t.Fatalf("tampered data served: found=%v err=%v", found, err)
	}
}
