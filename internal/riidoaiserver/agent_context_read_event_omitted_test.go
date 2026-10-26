package riidoaiserver

import (
	"encoding/json"
	"strconv"
	"testing"
)

func TestAgentContextReadOmittedCountExcludesHiddenAgents(t *testing.T) {
	s := NewDevelopmentAIAgentClientStore()
	reads := make([]AssignmentContextRead, 0, 34)
	for i := range agentContextReadEventMaxSources {
		reads = append(reads, AssignmentContextRead{AgentID: "agent-public-openclaw", SourceKind: "agent", SourceID: "missing-" + strconv.Itoa(i), SourcePart: "profile"})
	}
	reads = append(reads,
		AssignmentContextRead{AgentID: "agent-public-openclaw", SourceKind: "agent", SourceID: "agent-private-cursor", SourcePart: "profile"},
		AssignmentContextRead{AgentID: "agent-public-openclaw", SourceKind: "agent", SourceID: "agent-public-openclaw", SourcePart: "profile"},
	)
	payload := newAgentContextReadEvent(reads)
	if payload.OmittedSources != 2 {
		t.Fatalf("server-side omitted=%d", payload.OmittedSources)
	}
	omitted := func(principal AuthorizationResult, event AgentContextReadEvent) int {
		s.mu.Lock()
		defer s.mu.Unlock()
		visible, ok := clientEventForPrincipalLocked(s, principal, ClientStreamEvent{EventType: AgentClientEventContextRead, Payload: event})
		if !ok {
			t.Fatalf("event hidden from %s", principal.PrincipalID)
		}
		return visible.Payload.(AgentContextReadEvent).OmittedSources
	}
	if got := omitted(AuthorizationResult{PrincipalID: "user-1"}, payload); got != 1 {
		t.Fatalf("non-owner omitted=%d, want only the public agent", got)
	}
	if got := omitted(AuthorizationResult{PrincipalID: "user-2"}, payload); got != 2 {
		t.Fatalf("owner omitted=%d", got)
	}
	// A replayed event restored from JSON lost the bookkeeping: report none.
	raw, _ := json.Marshal(payload)
	var restored AgentContextReadEvent
	if err := json.Unmarshal(raw, &restored); err != nil {
		t.Fatal(err)
	}
	if got := omitted(AuthorizationResult{PrincipalID: "user-1"}, restored); got != 0 {
		t.Fatalf("restored directory event omitted=%d", got)
	}
	plain := newAgentContextReadEvent([]AssignmentContextRead{{AgentID: "agent-public-openclaw", SourceKind: "task", SourceID: "t", SourcePart: "body"}})
	plain.OmittedSources = 5 // restored non-directory event keeps its count
	if got := omitted(AuthorizationResult{PrincipalID: "user-1"}, plain); got != 5 {
		t.Fatalf("non-directory omitted=%d", got)
	}
}
