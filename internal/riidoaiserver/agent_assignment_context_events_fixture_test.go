package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
)

// contextReadEvents replays the client stream as token and returns the
// agent_context_read payloads it carries.
func (f *assignmentContextFixture) contextReadEvents(token string) []AgentContextReadEvent {
	req := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-dev-riid/ai-agent/events?replay=1", nil)
	req.Header.Set(aiAgentTokenHeader, token)
	resp := httptest.NewRecorder()
	f.server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		f.t.Fatalf("events status=%d body=%s", resp.Code, resp.Body.String())
	}
	var events []AgentContextReadEvent
	for block := range strings.SplitSeq(resp.Body.String(), "\n\n") {
		if !strings.Contains(block, "event: "+AgentClientEventContextRead) {
			continue
		}
		for line := range strings.SplitSeq(block, "\n") {
			if data, ok := strings.CutPrefix(line, "data: "); ok {
				var event AgentContextReadEvent
				if err := json.Unmarshal([]byte(data), &event); err != nil {
					f.t.Fatal(err)
				}
				events = append(events, event)
			}
		}
	}
	return events
}
