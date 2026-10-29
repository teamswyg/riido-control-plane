package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTaskReadScopeTypedSlotsCaptureAndOfflineReplay(t *testing.T) {
	scope, meeting, prefixes, payloads, evidence := readScopeTypedCaptureFixture()
	calls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Query     string
			Variables struct {
				Part     string
				Offset   int
				Selector *taskReadPartSelector
			}
		}
		if json.NewDecoder(r.Body).Decode(&req) != nil {
			t.Error("invalid request")
		}
		calls++
		root := "getComponentReadScope"
		var response any = scope
		if strings.Contains(req.Query, "getMeetingReadScope(") {
			root = "getMeetingReadScope"
			response = meeting
		}
		if req.Variables.Part != "" {
			first := prefixes[req.Variables.Part]
			page := typedJSONFixture(first.ResourceID, first.Part, first.Version, payloads[first.Part], req.Variables.Offset, first.Selector)
			if readEvidenceKey(req.Variables.Selector) != readEvidenceKey(first.Selector) {
				t.Error("selector lost")
			}
			kind := AssignmentContentComponent
			root = "getComponentReadScopeBody"
			if first.ResourceID == "meeting-0" {
				kind = AssignmentContentMeeting
				root = "getMeetingReadScopeBody"
			}
			evidence.BodyPages = append(evidence.BodyPages, taskReadCapturedBody{ResourceKind: kind, Page: page})
			response = page
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{root: response}})
	}))
	defer upstream.Close()
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL})
	req := AIAgentTaskContextRequest{ComponentID: "task-private", BearerToken: "synthetic"}
	live, err := client.getTaskReadScope(t.Context(), req, "team")
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range live.readScopeCapture.Manifest {
		if !part.Complete {
			t.Fatalf("unnecessarily incomplete %s", part.Part)
		}
	}
	index, err := newTaskReadEvidenceIndex(evidence, "task-private", "team")
	if err != nil {
		t.Fatal(err)
	}
	client.readScopeEvidence = index
	before := calls
	offline, err := client.getTaskReadScope(t.Context(), req, "team")
	if err != nil || calls != before {
		t.Fatalf("offline capture failed or called server: %v", err)
	}
	if len(offline.readScopeCapture.Pages) != len(live.readScopeCapture.Pages) {
		t.Fatal("offline pages differ")
	}
	for i, p := range live.readScopeCapture.Pages {
		if string(p.Data) != string(offline.readScopeCapture.Pages[i].Data) {
			t.Fatalf("offline payload differs at %d", i)
		}
	}
}
