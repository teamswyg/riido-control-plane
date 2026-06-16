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
	requests := make(chan capturedDynamoDBRequest, 2)
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
	assertDynamoDBString(t, putPayload.Item, "sk", dynamoDBAIAgentClientSnapshotSK)
	assertDynamoDBString(t, putPayload.Item, "schema_version", AIAgentClientPersistenceSchemaVersion)
	assertDynamoDBNumber(t, putPayload.Item, "next_device_credential_seq", "1")
	assertDynamoDBNumber(t, putPayload.Item, "next_daemon_command", "2")

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
	if getPayload.TableName != "riido-ai-agent-development" || getPayload.ConsistentRead {
		t.Fatalf("get payload = %+v", getPayload)
	}
	assertDynamoDBString(t, getPayload.Key, "pk", dynamoDBAIAgentClientSnapshotPK)
	assertDynamoDBString(t, getPayload.Key, "sk", dynamoDBAIAgentClientSnapshotSK)

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
