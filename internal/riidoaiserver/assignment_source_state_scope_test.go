package riidoaiserver

import "testing"

func TestAssignmentSourceStatesNeedSameConversationAndConfig(t *testing.T) {
	server, calls := newSourceStateTestServer(t, echoSourceStates(map[string]string{"gone": "deleted", "moved": "visible", "secret": "hidden"}))
	for name, mutate := range map[string]func(*Server, *AssignmentContextSnapshot){
		"no config": func(s *Server, _ *AssignmentContextSnapshot) { s.config.AgentSourceState = AgentSourceStateConfig{} },
		"short key": func(s *Server, _ *AssignmentContextSnapshot) { s.config.AgentSourceState.Key = "short" },
		"plain remote http": func(s *Server, _ *AssignmentContextSnapshot) {
			s.config.AgentSourceState.BaseURL = "http://example.com"
		},
		"no previous":        func(_ *Server, c *AssignmentContextSnapshot) { c.PreviousAssignmentID = "" },
		"other thread":       func(_ *Server, c *AssignmentContextSnapshot) { c.PreviousThreadID = "other-thread" },
		"other conversation": func(_ *Server, c *AssignmentContextSnapshot) { c.ConversationID = "other" },
		"other requester":    func(_ *Server, c *AssignmentContextSnapshot) { c.RequesterPrincipalID = "other" },
		"other team":         func(_ *Server, c *AssignmentContextSnapshot) { c.Context.TeamID = "other" },
		"legacy capture":     func(_ *Server, c *AssignmentContextSnapshot) { c.Context.ReadScope = nil },
	} {
		s, current := sourceStateTestServer(t, server)
		mutate(&s, &current)
		if states := s.assignmentSourceStates(t.Context(), current); states != nil {
			t.Fatalf("%s: states = %+v", name, states)
		}
	}
	if calls.Load() != 0 {
		t.Fatal("lookup ran without a verified previous capture")
	}
}

func TestAssignmentSourceStatesSkipOversizedOmissions(t *testing.T) {
	server, calls := newSourceStateTestServer(t, echoSourceStates(nil))
	extra := make([]string, assignmentSourceStateLimit+1)
	for i := range extra {
		extra[i] = "gone-" + string(rune('a'+i%26)) + string(rune('a'+i/26))
	}
	s, current := sourceStateTestServer(t, server, extra...)
	if states := s.assignmentSourceStates(t.Context(), current); states != nil || calls.Load() != 0 {
		t.Fatal("oversized omission set was looked up")
	}
}

func TestContextComparisonBudgetDropsSourceStatesWithLineage(t *testing.T) {
	snapshot := contextChangesTestSnapshot("current")
	snapshot.SourceStates = []AssignmentSourceState{{ResourceKind: AssignmentContentComponent, SourceID: "gone", State: AssignmentSourceDeleted}}
	clearAssignmentContextLineage(&snapshot)
	if snapshot.SourceStates != nil || snapshot.ConversationID != "" {
		t.Fatal("states kept without the lineage that gives them meaning")
	}
}
