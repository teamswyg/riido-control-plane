package riidoaiserver

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDynamoDBAIAgentClientSnapshotSavesAndLoads(t *testing.T) {
	fixedNow := fixedSnapshotTestNow()
	snapshot := snapshotTestRecord(fixedNow)
	fixture := newSnapshotDynamoDBFixture(t, fixedNow, nil, NewAIAgentClientPersistenceMetrics())
	defer fixture.close()

	if err := fixture.store.SaveAIAgentClientSnapshot(context.Background(), snapshot); err != nil {
		t.Fatalf("SaveAIAgentClientSnapshot: %v", err)
	}
	loaded, ok, err := fixture.store.LoadAIAgentClientSnapshot(context.Background())
	if err != nil {
		t.Fatalf("LoadAIAgentClientSnapshot: %v", err)
	}
	if !ok || loaded.WorkspaceID != "workspace-dev" || loaded.Agents[0].AgentID != "agent-a" {
		t.Fatalf("loaded snapshot ok=%v snapshot=%+v", ok, loaded)
	}
	assertSnapshotManifestAndParts(t, fixture.items, "1", "2")
	assertSnapshotFirstPutRequest(t, <-fixture.requests)
	assertSnapshotQueryRequest(t, drainSnapshotQueryRequest(t, fixture.requests))
	assertSnapshotMetrics(t, fixture.metrics)
}

func assertSnapshotFirstPutRequest(t *testing.T, request capturedDynamoDBRequest) {
	t.Helper()
	assertDynamoDBTarget(t, request, dynamoDBPutItemTarget)
	var payload struct {
		TableName string                       `json:"TableName"`
		Item      map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(request.body, &payload); err != nil {
		t.Fatalf("decode PutItem payload: %v", err)
	}
	if payload.TableName != "riido-ai-agent-development" {
		t.Fatalf("put table = %q", payload.TableName)
	}
	assertDynamoDBString(t, payload.Item, "pk", dynamoDBAIAgentClientSnapshotPK)
	assertDynamoDBString(t, payload.Item, "schema_version", AIAgentClientPersistenceSchemaVersion)
}

func assertSnapshotQueryRequest(t *testing.T, request capturedDynamoDBRequest) {
	t.Helper()
	assertDynamoDBTarget(t, request, dynamoDBQueryTarget)
	var payload snapshotQueryTestPayload
	if err := json.Unmarshal(request.body, &payload); err != nil {
		t.Fatalf("decode Query payload: %v", err)
	}
	if payload.TableName != "riido-ai-agent-development" || payload.ConsistentRead || payload.KeyConditionExpression != "pk = :pk" {
		t.Fatalf("query payload = %+v", payload)
	}
	assertDynamoDBString(t, payload.ExpressionValues, ":pk", dynamoDBAIAgentClientSnapshotPK)
}
