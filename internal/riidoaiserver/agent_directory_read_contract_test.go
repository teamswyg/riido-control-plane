package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"os"
	"slices"
	"testing"
	"time"
)

func TestAgentDirectoryReadsMatchHTTPAndSSEOpenAPI(t *testing.T) {
	source := agentDirectorySource(AgentDirectoryEntry{AgentID: "agent-public-openclaw", Name: "Public agent"})
	reads := assignmentContextReadRecords(AssignmentContextSnapshot{}, []AssignmentContextSource{source}, AssignmentContextViaCLI, "served", "request", time.Now())
	server, _, _ := runReadsTestServer(t, reads)
	response := runReadsRequest(server, "/v1/client/ai-agent/runs/run/reads")
	if response.Code != http.StatusOK {
		t.Fatalf("run reads status=%d", response.Code)
	}
	var body AIAgentRunReadsResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Reads) != 1 {
		t.Fatalf("run reads count=%d", len(body.Reads))
	}
	event := newAgentContextReadEvent(reads)
	if len(event.Sources) != 1 {
		t.Fatalf("SSE sources count=%d", len(event.Sources))
	}
	for name, actual := range map[string]AgentContextReadSource{
		"AIAgentRunRead": body.Reads[0].AgentContextReadSource, "AgentContextReadSource": event.Sources[0],
	} {
		t.Run(name, func(t *testing.T) {
			if actual.Kind != "agent" || actual.Part != "profile" {
				t.Fatalf("agent directory source=%+v", actual)
			}
			assertAgentReadSourceOpenAPI(t, name, actual)
		})
	}
}

func assertAgentReadSourceOpenAPI(t *testing.T, name string, source AgentContextReadSource) {
	t.Helper()
	data, err := os.ReadFile("../../contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Components struct {
			Schemas map[string]struct {
				Properties map[string]struct{ Enum []string }
			}
		}
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	for field, value := range map[string]string{"kind": source.Kind, "part": source.Part} {
		if !slices.Contains(spec.Components.Schemas[name].Properties[field].Enum, value) {
			t.Errorf("%s.%s rejects emitted value %q", name, field, value)
		}
	}
}
