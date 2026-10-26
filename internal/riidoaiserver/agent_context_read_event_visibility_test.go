package riidoaiserver

import "testing"

func TestAgentContextReadEventHidesAgentsTheSubscriberCannotSee(t *testing.T) {
	s := NewDevelopmentAIAgentClientStore()
	event := ClientStreamEvent{EventType: AgentClientEventContextRead, Payload: AgentContextReadEvent{
		EventType: AgentClientEventContextRead, AgentID: "agent-public-openclaw", Via: AssignmentContextViaCLI,
		Sources: []AgentContextReadSource{
			{Kind: "task", ID: "task-private", Part: "body"},
			{Kind: "agent", ID: "agent-public-openclaw", Part: "profile"},
			{Kind: "agent", ID: "agent-private-cursor", Part: "profile", Version: "sha256:private"},
			{Kind: "agent", ID: "agent-deleted", Part: "profile"},
		},
	}}
	ids := func(principal AuthorizationResult) []string {
		s.mu.Lock()
		defer s.mu.Unlock()
		visible, ok := clientEventForPrincipalLocked(s, principal, event)
		if !ok {
			t.Fatalf("public run event hidden from %s", principal.PrincipalID)
		}
		var out []string
		for _, source := range visible.Payload.(AgentContextReadEvent).Sources {
			out = append(out, source.ID)
		}
		return out
	}
	if got := ids(AuthorizationResult{PrincipalID: "user-1"}); len(got) != 2 || got[0] != "task-private" || got[1] != "agent-public-openclaw" {
		t.Fatalf("non-owner saw %v", got)
	}
	if got := ids(AuthorizationResult{PrincipalID: "user-2"}); len(got) != 3 || got[2] != "agent-private-cursor" {
		t.Fatalf("owner should see its private agent: %v", got)
	}
	if got := ids(AuthorizationResult{PrincipalID: "admin", Roles: []AgentCatalogRole{AgentCatalogRoleAdmin}}); len(got) != 3 {
		t.Fatalf("admin view: %v", got)
	}
	if original := event.Payload.(AgentContextReadEvent).Sources; len(original) != 4 {
		t.Fatal("filtering must not mutate the shared event")
	}
}
