package riidoaiserver

import (
	"context"
	"encoding/json"
	"testing"
)

func TestAgentCatalogAPIDTOsExposeOnlyCatalogShape(t *testing.T) {
	createPayload, err := json.Marshal(CreateAgentCatalogRequest{
		AgentID:    "agent-a",
		Visibility: AgentCatalogVisibilityPublic,
	})
	if err != nil {
		t.Fatalf("marshal create request: %v", err)
	}
	if got, want := string(createPayload), `{"agent_id":"agent-a","visibility":"public"}`; got != want {
		t.Fatalf("create request JSON = %s, want %s", got, want)
	}
	updatePayload, err := json.Marshal(UpdateAgentCatalogRequest{Visibility: AgentCatalogVisibilityPrivate})
	if err != nil {
		t.Fatalf("marshal update request: %v", err)
	}
	if got, want := string(updatePayload), `{"visibility":"private"}`; got != want {
		t.Fatalf("update request JSON = %s, want %s", got, want)
	}
	if jsonObjectHasKey(t, createPayload, "owner_principal_id") ||
		jsonObjectHasKey(t, createPayload, "roles") ||
		jsonObjectHasKey(t, updatePayload, "owner_principal_id") ||
		jsonObjectHasKey(t, updatePayload, "roles") {
		t.Fatal("agent catalog request DTOs must not accept owner or role input")
	}
}

func TestAgentCatalogAPIResponsesIncludeSchemaAndRecords(t *testing.T) {
	record := AgentCatalogRecord{
		AgentID:          "agent-a",
		OwnerPrincipalID: "user-a",
		Visibility:       AgentCatalogVisibilityPublic,
	}
	listPayload, err := json.Marshal(AgentCatalogListResponse{
		SchemaVersion: "riido-ai-server.v1",
		Agents:        []AgentCatalogRecord{record},
	})
	if err != nil {
		t.Fatalf("marshal list response: %v", err)
	}
	if got, want := string(listPayload), `{"schema_version":"riido-ai-server.v1","agents":[{"agent_id":"agent-a","owner_principal_id":"user-a","visibility":"public"}]}`; got != want {
		t.Fatalf("list response JSON = %s, want %s", got, want)
	}
	recordPayload, err := json.Marshal(AgentCatalogRecordResponse{
		SchemaVersion: "riido-ai-server.v1",
		Agent:         record,
	})
	if err != nil {
		t.Fatalf("marshal record response: %v", err)
	}
	if got, want := string(recordPayload), `{"schema_version":"riido-ai-server.v1","agent":{"agent_id":"agent-a","owner_principal_id":"user-a","visibility":"public"}}`; got != want {
		t.Fatalf("record response JSON = %s, want %s", got, want)
	}
}

func TestAgentCatalogStorePortShape(t *testing.T) {
	var store AgentCatalogStore = &memoryAgentCatalogStore{}
	ctx := context.Background()
	saved, err := store.SaveAgentCatalog(ctx, AgentCatalogRecord{
		AgentID:          "agent-a",
		OwnerPrincipalID: "user-a",
		Visibility:       AgentCatalogVisibilityPublic,
	})
	if err != nil {
		t.Fatalf("SaveAgentCatalog: %v", err)
	}
	if saved.AgentID != "agent-a" {
		t.Fatalf("saved record = %+v", saved)
	}
	got, ok, err := store.GetAgentCatalog(ctx, "agent-a")
	if err != nil || !ok || got.AgentID != "agent-a" {
		t.Fatalf("GetAgentCatalog = %+v %v %v", got, ok, err)
	}
	records, err := store.ListAgentCatalog(ctx)
	if err != nil || len(records) != 1 {
		t.Fatalf("ListAgentCatalog = %+v %v", records, err)
	}
	deleted, err := store.DeleteAgentCatalog(ctx, "agent-a")
	if err != nil || !deleted {
		t.Fatalf("DeleteAgentCatalog = %v %v", deleted, err)
	}
}

func jsonObjectHasKey(t *testing.T, data []byte, key string) bool {
	t.Helper()
	var out map[string]json.RawMessage
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal JSON object: %v", err)
	}
	_, ok := out[key]
	return ok
}

type memoryAgentCatalogStore struct {
	records map[string]AgentCatalogRecord
}

func (s *memoryAgentCatalogStore) ListAgentCatalog(context.Context) ([]AgentCatalogRecord, error) {
	records := make([]AgentCatalogRecord, 0, len(s.records))
	for _, record := range s.records {
		records = append(records, record)
	}
	return records, nil
}

func (s *memoryAgentCatalogStore) GetAgentCatalog(_ context.Context, agentID string) (AgentCatalogRecord, bool, error) {
	record, ok := s.records[agentID]
	return record, ok, nil
}

func (s *memoryAgentCatalogStore) SaveAgentCatalog(_ context.Context, record AgentCatalogRecord) (AgentCatalogRecord, error) {
	if s.records == nil {
		s.records = map[string]AgentCatalogRecord{}
	}
	s.records[record.AgentID] = record
	return record, nil
}

func (s *memoryAgentCatalogStore) DeleteAgentCatalog(_ context.Context, agentID string) (bool, error) {
	if _, ok := s.records[agentID]; !ok {
		return false, nil
	}
	delete(s.records, agentID)
	return true, nil
}
