package riidoaiserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAIAgentRunReadsFiltersCurrentViewerAndRedacts(t *testing.T) {
	reads := []AssignmentContextRead{
		{SourceKind: "task", SourceID: "task", SourcePart: "body", SourceVersion: "sha256:past"},
		{SourceKind: "project", SourceID: "private-parent", SourcePart: "body", SourceVersion: "private-hash"},
		{SourceKind: "link", SourceID: "removed-link", SourcePart: "metadata"},
		{SourceKind: "comment", SourceID: "visible-comment", SourcePart: "body"},
		{SourceKind: "related_meeting", SourceID: "visible-meeting", SourcePart: "title"},
	}
	server, reader, auth := runReadsTestServer(t, reads)
	reader.value.Hierarchy.Project = AIAgentTaskContextReference{ID: "private-parent", Title: "Known title", DocumentStatus: "unavailable"}
	reader.value.Extended = &AIAgentTaskContextExtended{Comments: []AIAgentTaskContextComment{{ID: "visible-comment"}}, RelatedMeetings: []AIAgentTaskContextRelatedMeeting{{ID: "visible-meeting"}}}
	for _, path := range []string{"/v1/client/ai-agent/runs/run/reads", "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent/runs/run/reads"} {
		response := runReadsRequest(server, path)
		if response.Code != 200 {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		var body struct {
			Reads    []map[string]any `json:"reads"`
			Page     int              `json:"page"`
			PageSize int              `json:"page_size"`
			HasMore  bool             `json:"has_more"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if len(body.Reads) != 3 || body.Page != 1 || body.PageSize != 100 || body.HasMore {
			t.Fatalf("unexpected response: %s", response.Body.String())
		}
		for _, forbidden := range []string{"private-parent", "private-hash", "removed-link", "secret-", "current body", "requester_principal_id", "request_id", "trigger_id"} {
			if strings.Contains(response.Body.String(), forbidden) {
				t.Fatalf("leaked %q: %s", forbidden, response.Body.String())
			}
		}
	}
	for _, req := range reader.requests {
		if req.BearerToken != "viewer-token" || !req.Extended || req.WorkspaceID != defaultAIAgentClientWorkspaceID {
			t.Fatalf("unsafe context request: %+v", req)
		}
	}
	for _, token := range auth.tokens {
		if token != "viewer-token" {
			t.Fatalf("wrong auth identity: %s", token)
		}
	}
}
