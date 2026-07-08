package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDynamoDBStoreSnapshotLoadRejectsInvalidResponses(t *testing.T) {
	badSchema := mustStoreSnapshotDynamoDBResponse(t, StoreSnapshot{SchemaVersion: "old"})
	for _, tc := range []struct {
		name string
		body string
		want string
	}{
		{"malformed dynamodb json", `{`, "decode DynamoDB store snapshot response"},
		{"missing snapshot json", `{"Item":{"snapshot_json":{"S":""}}}`, "snapshot_json is required"},
		{"malformed snapshot json", `{"Item":{"snapshot_json":{"S":"{"}}}`, "decode DynamoDB store snapshot json"},
		{"unsupported schema", badSchema, "unsupported store snapshot schema_version"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := newDynamoDBStoreSnapshotForBoundary(t, tc.body)
			defer store.Close()
			_, _, err := store.LoadStoreSnapshot(context.Background())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("LoadStoreSnapshot error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestDynamoDBStoreSnapshotSaveRejectsInvalidSchemaAndPutError(t *testing.T) {
	store := newDynamoDBStoreSnapshotForBoundary(t, `{}`)
	defer store.Close()
	err := store.SaveStoreSnapshot(context.Background(), StoreSnapshot{SchemaVersion: "old"})
	if err == nil || !strings.Contains(err.Error(), "unsupported store snapshot schema_version") {
		t.Fatalf("SaveStoreSnapshot invalid schema error = %v", err)
	}

	putFail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"message":"boom"}`, http.StatusInternalServerError)
	}))
	defer putFail.Close()
	store, err = NewDynamoDBStoreSnapshot(DynamoDBStoreSnapshotConfig{
		Region:              "ap-northeast-2",
		TableName:           "assignments",
		Endpoint:            putFail.URL,
		CredentialsProvider: mustStaticAWSTestProvider(t, "AKID", ""),
	})
	if err != nil {
		t.Fatalf("NewDynamoDBStoreSnapshot: %v", err)
	}
	defer store.Close()
	err = store.SaveStoreSnapshot(context.Background(), StoreSnapshot{SchemaVersion: StoreSnapshotSchemaVersion})
	if err == nil || !strings.Contains(err.Error(), "dynamodb save store snapshot") {
		t.Fatalf("SaveStoreSnapshot put error = %v", err)
	}
}

func mustStoreSnapshotDynamoDBResponse(t *testing.T, snapshot StoreSnapshot) string {
	t.Helper()
	raw, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	response, err := json.Marshal(map[string]any{"Item": map[string]any{"snapshot_json": map[string]string{"S": string(raw)}}})
	if err != nil {
		t.Fatalf("marshal DynamoDB response: %v", err)
	}
	return string(response)
}
