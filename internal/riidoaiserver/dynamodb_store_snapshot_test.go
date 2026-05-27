package riidoaiserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDynamoDBStoreSnapshotSavesAndLoads(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 2, 3, 4, 0, time.UTC)
	requests := make(chan capturedDynamoDBRequest, 2)
	snapshot := StoreSnapshot{
		SchemaVersion: StoreSnapshotSchemaVersion,
		SavedAt:       fixedNow,
		Tasks: []StoreSnapshotTask{
			{ID: "task-a", ComponentID: "component-1", CurrentAssignmentID: "asn-000001"},
		},
		Assignments: []Assignment{
			{
				ID:              "asn-000001",
				TaskID:          "task-a",
				ComponentID:     "component-1",
				AgentID:         "jykim1",
				RuntimeProvider: "codex",
				Prompt:          "make hello world",
				State:           AssignmentQueued,
				CreatedAt:       fixedNow,
				UpdatedAt:       fixedNow,
			},
		},
		AgentAssignments:  map[string][]string{"jykim1": {"asn-000001"}},
		Events:            map[string][]TaskEvent{"task-a": {{Seq: 1, TaskID: "task-a", AssignmentID: "asn-000001", AgentID: "jykim1", Type: EventAssignmentQueued, State: AssignmentQueued, At: fixedNow}}},
		NextAssignmentSeq: 1,
		NextEventSeq:      1,
	}
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		requests <- capturedDynamoDBRequest{
			method: r.Method,
			header: r.Header.Clone(),
			body:   body,
		}
		if r.Header.Get("X-Amz-Target") == dynamoDBGetItemTarget {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"Item": map[string]any{
					"snapshot_json": map[string]string{"S": string(snapshotJSON)},
				},
			})
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	store, err := NewDynamoDBStoreSnapshot(DynamoDBStoreSnapshotConfig{
		Region:              "ap-northeast-2",
		TableName:           "riido-ai-server-assignments",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
		Now:                 func() time.Time { return fixedNow },
	})
	if err != nil {
		t.Fatalf("NewDynamoDBStoreSnapshot: %v", err)
	}
	defer store.Close()

	if err := store.SaveStoreSnapshot(context.Background(), snapshot); err != nil {
		t.Fatalf("SaveStoreSnapshot: %v", err)
	}
	loaded, ok, err := store.LoadStoreSnapshot(context.Background())
	if err != nil {
		t.Fatalf("LoadStoreSnapshot: %v", err)
	}
	if !ok {
		t.Fatal("expected snapshot")
	}
	if loaded.NextAssignmentSeq != 1 || len(loaded.Assignments) != 1 || loaded.Assignments[0].ID != "asn-000001" {
		t.Fatalf("loaded snapshot = %+v", loaded)
	}

	putRequest := <-requests
	assertDynamoDBTarget(t, putRequest, dynamoDBPutItemTarget)
	var putPayload struct {
		TableName string                       `json:"TableName"`
		Item      map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(putRequest.body, &putPayload); err != nil {
		t.Fatalf("decode PutItem payload: %v", err)
	}
	if putPayload.TableName != "riido-ai-server-assignments" {
		t.Fatalf("put table = %q", putPayload.TableName)
	}
	assertDynamoDBString(t, putPayload.Item, "pk", dynamoDBSnapshotPK)
	assertDynamoDBString(t, putPayload.Item, "sk", dynamoDBSnapshotSK)
	assertDynamoDBString(t, putPayload.Item, "schema_version", StoreSnapshotSchemaVersion)
	assertDynamoDBNumber(t, putPayload.Item, "next_assignment_seq", "1")
	if putPayload.Item["snapshot_json"]["S"] == "" {
		t.Fatal("snapshot_json missing")
	}

	getRequest := <-requests
	assertDynamoDBTarget(t, getRequest, dynamoDBGetItemTarget)
	var getPayload struct {
		TableName      string                       `json:"TableName"`
		ConsistentRead bool                         `json:"ConsistentRead"`
		Key            map[string]map[string]string `json:"Key"`
	}
	if err := json.Unmarshal(getRequest.body, &getPayload); err != nil {
		t.Fatalf("decode GetItem payload: %v", err)
	}
	if getPayload.TableName != "riido-ai-server-assignments" || !getPayload.ConsistentRead {
		t.Fatalf("get payload = %+v", getPayload)
	}
	assertDynamoDBString(t, getPayload.Key, "pk", dynamoDBSnapshotPK)
	assertDynamoDBString(t, getPayload.Key, "sk", dynamoDBSnapshotSK)
}

func TestDynamoDBStoreSnapshotLoadsMissingSnapshot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	store, err := NewDynamoDBStoreSnapshot(DynamoDBStoreSnapshotConfig{
		Region:              "ap-northeast-2",
		TableName:           "assignments",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
	})
	if err != nil {
		t.Fatalf("NewDynamoDBStoreSnapshot: %v", err)
	}
	defer store.Close()

	_, ok, err := store.LoadStoreSnapshot(context.Background())
	if err != nil {
		t.Fatalf("LoadStoreSnapshot: %v", err)
	}
	if ok {
		t.Fatal("expected missing snapshot")
	}
}

func TestDynamoDBStoreSnapshotRejectsInvalidConfig(t *testing.T) {
	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	if _, err := NewDynamoDBStoreSnapshot(DynamoDBStoreSnapshotConfig{Region: "ap-northeast-2", CredentialsProvider: provider}); err == nil {
		t.Fatal("expected missing table error")
	}
	if _, err := NewDynamoDBStoreSnapshot(DynamoDBStoreSnapshotConfig{Region: "ap-northeast-2", TableName: "assignments"}); err == nil {
		t.Fatal("expected missing credentials provider error")
	}
}

func assertDynamoDBTarget(t *testing.T, request capturedDynamoDBRequest, want string) {
	t.Helper()
	if request.method != http.MethodPost {
		t.Fatalf("method = %s", request.method)
	}
	if request.header.Get("X-Amz-Target") != want {
		t.Fatalf("target = %q", request.header.Get("X-Amz-Target"))
	}
}
