package riidoaiserver

import (
	"errors"
	"strings"
	"testing"
)

func TestAIAgentRunReadsFailClosed(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*Server, *runReadsTestContext, *runReadsTestAuthorizer)
		path   string
		code   int
	}{
		{"task denied", func(_ *Server, _ *runReadsTestContext, a *runReadsTestAuthorizer) { a.fail = true }, "/v1/client/ai-agent/runs/run/reads", 403},
		{"context unavailable", func(_ *Server, r *runReadsTestContext, _ *runReadsTestAuthorizer) {
			r.err = errors.New("secret backend")
		}, "/v1/client/ai-agent/runs/run/reads", 503},
		{"empty fallback", func(_ *Server, r *runReadsTestContext, _ *runReadsTestAuthorizer) { r.value = AIAgentTaskContext{} }, "/v1/client/ai-agent/runs/run/reads", 403},
		{"other workspace", func(*Server, *runReadsTestContext, *runReadsTestAuthorizer) {}, "/v2/client/workspaces/other/ai-agent/runs/run/reads", 404},
		{"unknown run", func(*Server, *runReadsTestContext, *runReadsTestAuthorizer) {}, "/v1/client/ai-agent/runs/unknown/reads", 404},
		{"bad page", func(*Server, *runReadsTestContext, *runReadsTestAuthorizer) {}, "/v1/client/ai-agent/runs/run/reads?page=0", 400},
	} {
		t.Run(test.name, func(t *testing.T) {
			server, reader, auth := runReadsTestServer(t, []AssignmentContextRead{{SourceKind: "task", SourceID: "task", SourcePart: "body"}})
			test.change(&server, reader, auth)
			response := runReadsRequest(server, test.path)
			if response.Code != test.code || strings.Contains(response.Body.String(), "secret backend") {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}
