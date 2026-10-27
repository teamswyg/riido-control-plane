package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func runReadsMetadataRows(t *testing.T, server Server, path string) []map[string]any {
	t.Helper()
	response := runReadsRequest(server, path)
	var body struct {
		Reads []map[string]any `json:"reads"`
	}
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body.Reads
}

func TestAIAgentRunReadsMetadataDoesNotRestoreUnavailableTitles(t *testing.T) {
	server, reader, _ := runReadsTestServer(t, []AssignmentContextRead{
		{SourceKind: "task", SourceID: "task", SourcePart: "body"},
		{SourceKind: "project", SourceID: "removed", SourcePart: "body"},
		{SourceKind: "link", SourceID: "hidden", SourcePart: "metadata"},
	})
	reader.value.Component.Title = " \t "
	reader.value.Resources = []AIAgentTaskContextResource{{ID: "hidden", Title: "Hidden resource"}}
	reader.value.ResourceStatus = "unavailable"
	rows := runReadsMetadataRows(t, server, "/v1/client/ai-agent/runs/run/reads")
	if len(rows) != 1 {
		t.Fatalf("hidden sources restored: %v", rows)
	}
	if _, ok := rows[0]["title"]; ok {
		t.Fatalf("blank current title must be omitted: %v", rows[0])
	}
}

type runReadsMetadataAgentStore struct {
	*DevelopmentAIAgentClientStore
	calls int
}

func (s *runReadsMetadataAgentStore) ListAIAgentTaskAssignableAgents(context.Context, AuthorizationResult, string) (AgentClientListResponse, error) {
	s.calls++
	return AgentClientListResponse{Agents: []AgentClientRecord{{AgentID: "visible", Name: " Current agent ", Description: "Current role"}}}, nil
}

func TestAIAgentRunReadsMetadataReusesCurrentAgentDirectory(t *testing.T) {
	source := agentDirectorySource(AgentDirectoryEntry{Name: "Current agent", Description: "Current role"})
	server, reader, _ := runReadsTestServer(t, []AssignmentContextRead{
		{SourceKind: "agent", SourceID: "visible", SourcePart: "profile", SourceVersion: source.Version},
		{SourceKind: "agent", SourceID: "removed", SourcePart: "profile"},
	})
	store := &runReadsMetadataAgentStore{DevelopmentAIAgentClientStore: server.aiAgent.(*DevelopmentAIAgentClientStore)}
	server.aiAgent = store
	rows := runReadsMetadataRows(t, server, "/v1/client/ai-agent/runs/run/reads")
	if len(rows) != 1 || rows[0]["title"] != "Current agent" || rows[0]["modified_since_read"] != false {
		t.Fatalf("reads=%v", rows)
	}
	if store.calls != 1 || len(reader.requests) != 1 {
		t.Fatalf("directory calls=%d current context calls=%d", store.calls, len(reader.requests))
	}
}
