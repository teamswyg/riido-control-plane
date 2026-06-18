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

func TestDynamoDBAIAgentClientSnapshotSavesAndLoads(t *testing.T) {
	fixedNow := time.Date(2026, 6, 2, 1, 2, 3, 0, time.UTC)
	requests := make(chan capturedDynamoDBRequest, 32)
	items := map[string]map[string]map[string]string{}
	snapshot := AIAgentClientSnapshot{
		SchemaVersion:           AIAgentClientPersistenceSchemaVersion,
		SavedAt:                 fixedNow,
		WorkspaceID:             "workspace-dev",
		Devices:                 []DeviceRecord{{DeviceID: "device-a", OwnerPrincipalID: "user-1"}},
		Agents:                  []AgentClientRecord{{AgentID: "agent-a", OwnerPrincipalID: "user-1", WorkspaceID: "workspace-dev", Name: "Agent A", Visibility: AgentVisibilityPrivate}},
		TaskThreads:             map[string][]AIAgentTaskThreadRecord{},
		NextDeviceCredentialSeq: 1,
		NextDaemonCommand:       2,
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
		switch r.Header.Get("X-Amz-Target") {
		case dynamoDBPutItemTarget:
			var payload struct {
				Item map[string]map[string]string `json:"Item"`
			}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Errorf("decode PutItem payload: %v", err)
			}
			items[payload.Item["sk"]["S"]] = payload.Item
			_, _ = w.Write([]byte(`{}`))
		case dynamoDBQueryTarget:
			out := make([]map[string]map[string]string, 0, len(items))
			for _, item := range items {
				out = append(out, item)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"Items": out})
		default:
			t.Errorf("target = %q", r.Header.Get("X-Amz-Target"))
			_, _ = w.Write([]byte(`{}`))
			return
		}
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	metrics := NewAIAgentClientPersistenceMetrics()
	store, err := NewDynamoDBAIAgentClientSnapshot(DynamoDBAIAgentClientSnapshotConfig{
		Region:              "ap-northeast-2",
		TableName:           "riido-ai-agent-development",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
		Now:                 func() time.Time { return fixedNow },
		Metrics:             metrics,
	})
	if err != nil {
		t.Fatalf("NewDynamoDBAIAgentClientSnapshot: %v", err)
	}
	defer store.Close()

	if err := store.SaveAIAgentClientSnapshot(context.Background(), snapshot); err != nil {
		t.Fatalf("SaveAIAgentClientSnapshot: %v", err)
	}
	loaded, ok, err := store.LoadAIAgentClientSnapshot(context.Background())
	if err != nil {
		t.Fatalf("LoadAIAgentClientSnapshot: %v", err)
	}
	if !ok {
		t.Fatal("expected snapshot")
	}
	if loaded.WorkspaceID != "workspace-dev" || len(loaded.Agents) != 1 || loaded.Agents[0].AgentID != "agent-a" {
		t.Fatalf("loaded snapshot = %+v", loaded)
	}
	current := items[dynamoDBAIAgentClientSnapshotSK]
	if current == nil {
		t.Fatal("missing CURRENT manifest item")
	}
	assertDynamoDBString(t, current, "pk", dynamoDBAIAgentClientSnapshotPK)
	assertDynamoDBString(t, current, "sk", dynamoDBAIAgentClientSnapshotSK)
	assertDynamoDBString(t, current, "schema_version", AIAgentClientPersistenceSchemaVersion)
	assertDynamoDBString(t, current, "storage_version", dynamoDBAIAgentClientSnapshotSplitStorageVersion)
	assertDynamoDBNumber(t, current, "next_device_credential_seq", "1")
	assertDynamoDBNumber(t, current, "next_daemon_command", "2")
	if current["snapshot_gzip"]["S"] != "" || current["snapshot_json"]["S"] != "" {
		t.Fatalf("CURRENT item should be a small manifest, got %+v", current)
	}
	for _, partName := range dynamoDBAIAgentClientSnapshotPartNames {
		item := items[dynamoDBAIAgentClientSnapshotPartSK(partName)]
		if item == nil {
			t.Fatalf("missing split part %s", partName)
		}
		if item["part_gzip"]["S"] == "" || item["part_hash"]["S"] == "" {
			t.Fatalf("part %s missing gzip/hash: %+v", partName, item)
		}
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
	if putPayload.TableName != "riido-ai-agent-development" {
		t.Fatalf("put table = %q", putPayload.TableName)
	}
	assertDynamoDBString(t, putPayload.Item, "pk", dynamoDBAIAgentClientSnapshotPK)
	assertDynamoDBString(t, putPayload.Item, "schema_version", AIAgentClientPersistenceSchemaVersion)

	for len(requests) > 1 {
		<-requests
	}
	queryRequest := <-requests
	assertDynamoDBTarget(t, queryRequest, dynamoDBQueryTarget)
	var queryPayload struct {
		TableName              string                       `json:"TableName"`
		ConsistentRead         bool                         `json:"ConsistentRead"`
		KeyConditionExpression string                       `json:"KeyConditionExpression"`
		ExpressionValues       map[string]map[string]string `json:"ExpressionAttributeValues"`
	}
	if err := json.Unmarshal(queryRequest.body, &queryPayload); err != nil {
		t.Fatalf("decode Query payload: %v", err)
	}
	if queryPayload.TableName != "riido-ai-agent-development" || queryPayload.ConsistentRead || queryPayload.KeyConditionExpression != "pk = :pk" {
		t.Fatalf("query payload = %+v", queryPayload)
	}
	assertDynamoDBString(t, queryPayload.ExpressionValues, ":pk", dynamoDBAIAgentClientSnapshotPK)

	metricsSnapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{})
	if metricsSnapshot.AIAgentClientSnapshotSaveCallsTotal != 1 || metricsSnapshot.AIAgentClientSnapshotLoadCallsTotal != 1 {
		t.Fatalf("snapshot persistence call metrics = %+v", metricsSnapshot)
	}
	if metricsSnapshot.AIAgentClientSnapshotSaveBytesLast <= 0 || metricsSnapshot.AIAgentClientSnapshotLoadBytesLast <= 0 {
		t.Fatalf("snapshot persistence byte metrics = %+v", metricsSnapshot)
	}
	if metricsSnapshot.AIAgentClientSnapshotSaveLatencySamplesTotal != 1 || metricsSnapshot.AIAgentClientSnapshotLoadLatencySamplesTotal != 1 {
		t.Fatalf("snapshot persistence latency metrics = %+v", metricsSnapshot)
	}
}

func TestDynamoDBAIAgentClientSnapshotMigratesLegacySnapshotToSplitParts(t *testing.T) {
	fixedNow := time.Date(2026, 6, 2, 1, 2, 3, 0, time.UTC)
	legacy := AIAgentClientSnapshot{
		SchemaVersion:           AIAgentClientPersistenceSchemaVersion,
		SavedAt:                 fixedNow,
		WorkspaceID:             "workspace-dev",
		Agents:                  []AgentClientRecord{{AgentID: "agent-a", OwnerPrincipalID: "user-1", WorkspaceID: "workspace-dev", Name: "Agent A"}},
		TaskThreads:             map[string][]AIAgentTaskThreadRecord{},
		NextDeviceCredentialSeq: 7,
		NextDaemonCommand:       9,
	}
	legacyJSON, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("marshal legacy snapshot: %v", err)
	}
	items := map[string]map[string]map[string]string{
		dynamoDBAIAgentClientSnapshotSK: {
			"pk":             {"S": dynamoDBAIAgentClientSnapshotPK},
			"sk":             {"S": dynamoDBAIAgentClientSnapshotSK},
			"schema_version": {"S": AIAgentClientPersistenceSchemaVersion},
			"snapshot_json":  {"S": string(legacyJSON)},
		},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch r.Header.Get("X-Amz-Target") {
		case dynamoDBQueryTarget:
			out := make([]map[string]map[string]string, 0, len(items))
			for _, item := range items {
				out = append(out, item)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"Items": out})
		case dynamoDBPutItemTarget:
			var payload struct {
				Item map[string]map[string]string `json:"Item"`
			}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Errorf("decode PutItem payload: %v", err)
			}
			items[payload.Item["sk"]["S"]] = payload.Item
			_, _ = w.Write([]byte(`{}`))
		default:
			t.Errorf("target = %q", r.Header.Get("X-Amz-Target"))
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	store, err := NewDynamoDBAIAgentClientSnapshot(DynamoDBAIAgentClientSnapshotConfig{
		Region:              "ap-northeast-2",
		TableName:           "riido-ai-agent-development",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
		Now:                 func() time.Time { return fixedNow },
	})
	if err != nil {
		t.Fatalf("NewDynamoDBAIAgentClientSnapshot: %v", err)
	}
	defer store.Close()

	loaded, ok, err := store.LoadAIAgentClientSnapshot(context.Background())
	if err != nil {
		t.Fatalf("LoadAIAgentClientSnapshot: %v", err)
	}
	if !ok || loaded.NextDeviceCredentialSeq != 7 || loaded.NextDaemonCommand != 9 {
		t.Fatalf("loaded legacy snapshot ok=%v snapshot=%+v", ok, loaded)
	}
	current := items[dynamoDBAIAgentClientSnapshotSK]
	if current["snapshot_json"]["S"] != "" || current["snapshot_gzip"]["S"] != "" {
		t.Fatalf("legacy CURRENT item was not replaced by manifest: %+v", current)
	}
	assertDynamoDBString(t, current, "storage_version", dynamoDBAIAgentClientSnapshotSplitStorageVersion)
	for _, partName := range dynamoDBAIAgentClientSnapshotPartNames {
		if items[dynamoDBAIAgentClientSnapshotPartSK(partName)] == nil {
			t.Fatalf("migration missing part %s", partName)
		}
	}
}

func TestDynamoDBAIAgentClientSnapshotSkipsUnchangedSplitParts(t *testing.T) {
	fixedNow := time.Date(2026, 6, 2, 1, 2, 3, 0, time.UTC)
	requests := make(chan capturedDynamoDBRequest, 32)
	items := map[string]map[string]map[string]string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests <- capturedDynamoDBRequest{method: r.Method, header: r.Header.Clone(), body: body}
		if r.Header.Get("X-Amz-Target") != dynamoDBPutItemTarget {
			t.Errorf("target = %q", r.Header.Get("X-Amz-Target"))
			_, _ = w.Write([]byte(`{}`))
			return
		}
		var payload struct {
			Item map[string]map[string]string `json:"Item"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("decode PutItem payload: %v", err)
		}
		items[payload.Item["sk"]["S"]] = payload.Item
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()
	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	store, err := NewDynamoDBAIAgentClientSnapshot(DynamoDBAIAgentClientSnapshotConfig{
		Region:              "ap-northeast-2",
		TableName:           "riido-ai-agent-development",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
		Now:                 func() time.Time { return fixedNow },
	})
	if err != nil {
		t.Fatalf("NewDynamoDBAIAgentClientSnapshot: %v", err)
	}
	defer store.Close()

	snapshot := AIAgentClientSnapshot{
		SchemaVersion:           AIAgentClientPersistenceSchemaVersion,
		SavedAt:                 fixedNow,
		WorkspaceID:             "workspace-dev",
		Agents:                  []AgentClientRecord{{AgentID: "agent-a", OwnerPrincipalID: "user-1", WorkspaceID: "workspace-dev"}},
		TaskThreads:             map[string][]AIAgentTaskThreadRecord{},
		NextDeviceCredentialSeq: 1,
		NextDaemonCommand:       2,
	}
	if err := store.SaveAIAgentClientSnapshot(context.Background(), snapshot); err != nil {
		t.Fatalf("initial SaveAIAgentClientSnapshot: %v", err)
	}
	for len(requests) > 0 {
		<-requests
	}

	snapshot.NextDaemonCommand = 3
	snapshot.SavedAt = fixedNow.Add(time.Minute)
	if err := store.SaveAIAgentClientSnapshot(context.Background(), snapshot); err != nil {
		t.Fatalf("manifest-only SaveAIAgentClientSnapshot: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("manifest-only save wrote %d items, want 1", len(requests))
	}
	var payload struct {
		Item map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal((<-requests).body, &payload); err != nil {
		t.Fatalf("decode manifest-only PutItem payload: %v", err)
	}
	assertDynamoDBString(t, payload.Item, "sk", dynamoDBAIAgentClientSnapshotSK)
	assertDynamoDBNumber(t, payload.Item, "next_daemon_command", "3")
	if len(items) != len(dynamoDBAIAgentClientSnapshotPartNames)+1 {
		t.Fatalf("unexpected stored item count = %d", len(items))
	}
}

func TestDynamoDBAIAgentClientSnapshotLoadsMissingSnapshot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	store, err := NewDynamoDBAIAgentClientSnapshot(DynamoDBAIAgentClientSnapshotConfig{
		Region:              "ap-northeast-2",
		TableName:           "riido-ai-agent-development",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
	})
	if err != nil {
		t.Fatalf("NewDynamoDBAIAgentClientSnapshot: %v", err)
	}
	defer store.Close()

	_, ok, err := store.LoadAIAgentClientSnapshot(context.Background())
	if err != nil {
		t.Fatalf("LoadAIAgentClientSnapshot: %v", err)
	}
	if ok {
		t.Fatal("expected missing snapshot")
	}
}

func TestDynamoDBAIAgentClientSnapshotRejectsInvalidConfig(t *testing.T) {
	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	if _, err := NewDynamoDBAIAgentClientSnapshot(DynamoDBAIAgentClientSnapshotConfig{Region: "ap-northeast-2", CredentialsProvider: provider}); err == nil {
		t.Fatal("expected missing table error")
	}
	if _, err := NewDynamoDBAIAgentClientSnapshot(DynamoDBAIAgentClientSnapshotConfig{Region: "ap-northeast-2", TableName: "ai-agent"}); err == nil {
		t.Fatal("expected missing credentials provider error")
	}
}
