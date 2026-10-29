package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestAIAgentRunReadsPaginationAfterFiltering(t *testing.T) {
	var reads []AssignmentContextRead
	for range 205 {
		reads = append(reads, AssignmentContextRead{SourceKind: "task", SourceID: "task", SourcePart: "body"}, AssignmentContextRead{SourceKind: "link", SourceID: "secret", SourcePart: "metadata"})
	}
	server, _, _ := runReadsTestServer(t, reads)
	for _, test := range []struct {
		page, count int
		more        bool
	}{{1, 100, true}, {2, 100, true}, {3, 5, false}, {4, 0, false}, {int(^uint(0) >> 1), 0, false}} {
		response := runReadsRequest(server, "/v1/client/ai-agent/runs/run/reads?page="+strconv.Itoa(test.page))
		var body AIAgentRunReadsResponse
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if response.Code != 200 || len(body.Reads) != test.count || body.HasMore != test.more {
			t.Fatalf("page=%d status=%d len=%d more=%v", test.page, response.Code, len(body.Reads), body.HasMore)
		}
		if strings.Contains(response.Body.String(), "secret") {
			t.Fatal("filtered row leaked")
		}
	}
}

func TestAIAgentRunReadsRequiresBearerAndNormalizesRoute(t *testing.T) {
	server, _, _ := runReadsTestServer(t, nil)
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		req := httptest.NewRequest(method, "/v1/client/ai-agent/runs/run/reads", nil)
		req.Header.Set(aiAgentTokenHeader, "agent-token")
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, req)
		want := 401
		if method == http.MethodPost {
			want = 405
		}
		if response.Code != want {
			t.Fatalf("status=%d", response.Code)
		}
	}
	for _, prefix := range []string{"/v1/client/ai-agent", "/v2/client/workspaces/workspace/ai-agent"} {
		route := aiAgentClientHTTPRoute(http.MethodGet, prefix+"/runs/private-id/reads")
		if !strings.HasSuffix(route, "/runs/{run_id}/reads") || strings.Contains(route, "private-id") {
			t.Fatalf("route=%s", route)
		}
	}
}
