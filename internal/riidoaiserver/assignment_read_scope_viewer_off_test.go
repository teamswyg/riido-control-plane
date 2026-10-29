package riidoaiserver

import (
	"strings"
	"testing"
)

func TestAssignmentReadScopeViewerRetainedSnapshotWithCaptureOff(t *testing.T) {
	reads := []AssignmentContextRead{{SourceKind: "related_backlog", SourceID: "selected", SourcePart: "body", SourceVersion: "body-v1"}}
	s, _, _ := runReadsTestServer(t, reads)
	reader := installCapturedScopeViewer(&s)
	response := runReadsRequest(s, "/v1/client/ai-agent/runs/run/reads")
	if response.Code != 200 || !strings.Contains(response.Body.String(), "selected") || len(reader.requests) != 1 || !reader.requests[0].capturedReadScope || reader.requests[0].BearerToken != "viewer-token" {
		t.Fatal("retained scoped HTTP reads did not recheck typed current viewer")
	}
	for _, mode := range []string{"live", "replay"} {
		t.Run(mode, func(t *testing.T) {
			stream, _, _, event := contextReadStreamServer(t)
			current := installCapturedScopeViewer(&stream)
			payload := event.Payload.(AgentContextReadEvent)
			payload.Sources = []AgentContextReadSource{{Kind: "related_backlog", ID: "selected", Part: "body", Status: "served"}}
			event.Payload = payload
			out := contextReadStreamRequest(stream, mode, event)
			if out.Code != 200 || !strings.Contains(out.Body.String(), "selected") || len(current.requests) != 1 || !current.requests[0].capturedReadScope || current.requests[0].BearerToken != "viewer-token" {
				t.Fatal("retained scoped SSE did not recheck typed current viewer")
			}
		})
	}
}
