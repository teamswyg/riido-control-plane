package riidoaiserver

import (
	"strings"
	"testing"
)

func TestAIAgentRunReadsShowOnlyAgentsTheViewerCanSee(t *testing.T) {
	reads := []AssignmentContextRead{
		{SourceKind: "agent", SourceID: "agent-public-openclaw", SourcePart: "profile"},
		{SourceKind: "agent", SourceID: "agent-private-cursor", SourcePart: "profile"},
		{SourceKind: "agent", SourceID: "agent-owned-codex", SourcePart: "profile"},
	}
	server, _, _ := runReadsTestServer(t, reads)
	response := runReadsRequest(server, "/v1/client/ai-agent/runs/run/reads")
	body := response.Body.String()
	if response.Code != 200 || !strings.Contains(body, "agent-public-openclaw") {
		t.Fatalf("status=%d body=%s", response.Code, body)
	}
	for _, hidden := range []string{"agent-private-cursor", "agent-owned-codex"} {
		if strings.Contains(body, hidden) {
			t.Fatalf("viewer saw %s: %s", hidden, body)
		}
	}
}
