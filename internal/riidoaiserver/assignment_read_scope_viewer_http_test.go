package riidoaiserver

import (
	"strings"
	"testing"
)

func TestAssignmentReadScopeViewerHTTPAndStreamHideRemovedLeaf(t *testing.T) {
	reads := []AssignmentContextRead{{SourceKind: "related_backlog", SourceID: "selected", SourcePart: "body", SourceVersion: "body-v1"}, {SourceKind: "comment", SourceID: "visible-comment", SourcePart: "body", SourceVersion: "comment-v1"}, {SourceKind: "related_task", SourceID: "removed-private", SourcePart: "body", SourceVersion: "hidden-version"}}
	server, reader, _ := runReadsTestServer(t, reads)
	reader.value = scopedViewerContext()
	response := runReadsRequest(server, "/v1/client/ai-agent/runs/run/reads")
	if response.Code != 200 {
		t.Fatalf("viewer HTTP=%d", response.Code)
	}
	assertScopedViewerLeaves(t, response.Body.String())
	for _, mode := range []string{"live", "replay"} {
		t.Run(mode, func(t *testing.T) {
			stream, current, _, event := contextReadStreamServer(t)
			current.value = scopedViewerContext()
			payload := event.Payload.(AgentContextReadEvent)
			payload.Sources = nil
			for _, r := range reads {
				payload.Sources = append(payload.Sources, AgentContextReadSource{Kind: r.SourceKind, ID: r.SourceID, Part: r.SourcePart, Version: r.SourceVersion, Status: "served"})
			}
			event.Payload = payload
			response := contextReadStreamRequest(stream, mode, event)
			if response.Code != 200 || !strings.Contains(response.Body.String(), "event: agent_context_read") {
				t.Fatalf("stream missing=%d %s", response.Code, response.Body.String())
			}
			assertScopedViewerLeaves(t, response.Body.String())
			if len(current.requests) != 1 || current.requests[0].BearerToken != "viewer-token" {
				t.Fatal("scope did not recheck stream viewer")
			}
		})
	}
}

func scopedViewerContext() AIAgentTaskContext {
	return AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task"}, ReadScope: &AssignmentReadScope{RootID: "task", Resources: []AssignmentScopedResource{{Kind: AssignmentContentBacklog, ID: "selected", Title: "Visible backlog"}}}, readScopeCapture: &assignmentReadScopeCapture{Manifest: []AssignmentContentManifest{{Complete: true, Cursors: []string{"body", "comment"}}}, Pages: []AssignmentContentPage{{ResourceKind: AssignmentContentBacklog, SourceID: "selected", Cursor: "body", Sources: []AssignmentContextSource{{Kind: "related_backlog", ID: "selected", Part: "body", Version: "body-v1"}}}, {ResourceKind: AssignmentContentBacklog, SourceID: "selected", Cursor: "comment", Sources: []AssignmentContextSource{{Kind: "comment", ID: "visible-comment", Part: "body", Version: "comment-v1"}}}}}}
}

func assertScopedViewerLeaves(t *testing.T, body string) {
	t.Helper()
	for _, want := range []string{"selected", "visible-comment"} {
		if !strings.Contains(body, want) {
			t.Errorf("current leaf %s missing: %s", want, body)
		}
	}
	for _, hidden := range []string{"removed-private", "hidden-version", "omitted_sources", "secret-requester"} {
		if strings.Contains(body, hidden) {
			t.Fatalf("hidden metadata %s leaked: %s", hidden, body)
		}
	}
}
