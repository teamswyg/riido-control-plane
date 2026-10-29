package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTaskReadScopeTypedJSONRestoresLargeCanonicalPayload(t *testing.T) {
	resource := AssignmentScopedResource{Kind: AssignmentContentComponent, ID: "work", TeamID: "team"}
	selector, _ := normalizeReadScopePartSelector(resource.Kind, "metadata", nil)
	raw, _ := json.Marshal(map[string]any{"github": []any{map[string]string{"id": "pr", "body": strings.Repeat("큰본문", 18000)}}})
	version := readScopePartVersion(resource.Kind, resource.ID, resource.TeamID, "metadata", "riido_metadata", string(raw), selector)
	calls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Variables struct {
				Offset        int
				Selector      taskReadPartSelector
				Part, Version string
			}
		}
		if json.NewDecoder(r.Body).Decode(&req) != nil || req.Variables.Part != "metadata" || req.Variables.Version != version || r.Header.Get("Authorization") != "Bearer synthetic" {
			t.Error("wrong typed continuation request")
		}
		calls++
		page := typedJSONFixture(resource.ID, "metadata", version, string(raw), req.Variables.Offset, &selector)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getComponentReadScopeBody": page}})
	}))
	defer upstream.Close()
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL})
	first := typedJSONFixture(resource.ID, "metadata", version, string(raw), 0, &selector)
	got, complete, err := client.readScopeTypedJSON(t.Context(), "synthetic", resource, first, "metadata", selector)
	if err != nil || !complete || string(got) != string(raw) || calls < 2 {
		t.Fatalf("lost typed payload: bytes=%d/%d complete=%v calls=%d err=%v", len(got), len(raw), complete, calls, err)
	}
	first.Source = "wrong"
	if _, _, err = client.readScopeTypedJSON(t.Context(), "synthetic", resource, first, "metadata", selector); err == nil {
		t.Fatal("changed version accepted")
	}
}

func typedJSONFixture(id, part, version, raw string, offset int, selector *taskReadPartSelector) *taskReadBodyPage {
	end := offset + 8190
	if end > len(raw) {
		end = len(raw)
	}
	for end < len(raw) && raw[end]&0xc0 == 0x80 {
		end--
	}
	complete := end == len(raw)
	status := "partial"
	var next *int
	if complete {
		status = "available"
	} else {
		next = &end
	}
	return &taskReadBodyPage{ResourceID: id, Part: part, Version: version, Source: "riido_" + part, Format: "json", Offset: &offset, NextOffset: next, Complete: &complete, Status: status, Content: raw[offset:end], Selector: selector}
}
