package riidoaiserver

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestAssignmentSourceStatesAskOnlyAboutPreviousOmissions(t *testing.T) {
	var asked []any
	server, calls := newSourceStateTestServer(t, func(w http.ResponseWriter, body map[string]any) {
		asked = body["sources"].([]any)
		if body["workspace_id"] != "workspace" || body["principal_id"] != "owner" || len(body) != 3 {
			http.Error(w, "wrong requester", http.StatusBadRequest)
			return
		}
		echoSourceStates(map[string]string{"gone": "deleted", "moved": "visible", "secret": "hidden"})(w, body)
	})
	s, current := sourceStateTestServer(t, server)
	states := s.assignmentSourceStates(t.Context(), current)
	want := map[string]string{"gone": AssignmentSourceDeleted, "moved": AssignmentSourceUnselected, "secret": AssignmentSourceHidden}
	if len(states) != len(want) || calls.Load() != 1 || len(asked) != len(want) {
		t.Fatalf("states = %+v asked = %v", states, asked)
	}
	for _, state := range states {
		if want[state.SourceID] != state.State || state.ResourceKind != AssignmentContentComponent {
			t.Fatalf("state = %+v", state)
		}
	}
	for _, item := range asked {
		if source := item.(map[string]any); source["team_id"] != "team" || len(source) != 3 {
			t.Fatalf("asked with unexpected identity: %v", source)
		}
	}
}

func TestAssignmentSourceStatesFailClosedToUnknown(t *testing.T) {
	reorder := func(w http.ResponseWriter, _ map[string]any) {
		_ = json.NewEncoder(w).Encode(map[string]any{"states": []map[string]any{
			{"kind": "component", "id": "moved", "team_id": "team", "state": "visible"},
			{"kind": "component", "id": "gone", "team_id": "team", "state": "deleted"},
			{"kind": "component", "id": "secret", "team_id": "team", "state": "hidden"},
		}})
	}
	for name, handler := range map[string]sourceStateHandler{
		"server error":  func(w http.ResponseWriter, _ map[string]any) { http.Error(w, "down", http.StatusServiceUnavailable) },
		"denied state":  echoSourceStates(map[string]string{"gone": "denied", "moved": "visible", "secret": "hidden"}),
		"reordered":     reorder,
		"short answer":  func(w http.ResponseWriter, _ map[string]any) { _, _ = io.WriteString(w, `{"states":[]}`) },
		"extra field":   func(w http.ResponseWriter, _ map[string]any) { _, _ = io.WriteString(w, `{"states":[],"title":"x"}`) },
		"trailing data": func(w http.ResponseWriter, _ map[string]any) { _, _ = io.WriteString(w, `{"states":[]}{}`) },
	} {
		server, calls := newSourceStateTestServer(t, handler)
		s, current := sourceStateTestServer(t, server)
		if states := s.assignmentSourceStates(t.Context(), current); states != nil || calls.Load() != 1 {
			t.Fatalf("%s: states = %+v", name, states)
		}
	}
}
