package riidoaiserver

import (
	"errors"
	"testing"
	"time"
)

func contextChangesTestSnapshot(id string) AssignmentContextSnapshot {
	return AssignmentContextSnapshot{AssignmentID: id, TaskID: "work", AgentID: "agent", WorkspaceID: "workspace", RequesterPrincipalID: "owner", CapturedAt: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), ThreadID: id + "-thread", ConversationID: "conversation", Context: AIAgentTaskContext{TeamID: "team"}}
}

func TestContextChangesBaselineRequiresServerAncestorAndSameScope(t *testing.T) {
	first, previous, current := contextChangesTestSnapshot("first"), contextChangesTestSnapshot("previous"), contextChangesTestSnapshot("current")
	previous.PreviousAssignmentID, previous.PreviousThreadID = first.AssignmentID, first.ThreadID
	current.PreviousAssignmentID, current.PreviousThreadID = previous.AssignmentID, previous.ThreadID
	for name, mutate := range map[string]func(*AssignmentContextSnapshot){
		"valid ancestor": func(*AssignmentContextSnapshot) {},
		"principal":      func(s *AssignmentContextSnapshot) { s.RequesterPrincipalID = "other" },
		"workspace":      func(s *AssignmentContextSnapshot) { s.WorkspaceID = "other" },
		"agent":          func(s *AssignmentContextSnapshot) { s.AgentID = "other" },
		"task":           func(s *AssignmentContextSnapshot) { s.TaskID = "other" },
		"team":           func(s *AssignmentContextSnapshot) { s.Context.TeamID = "other" },
		"admin":          func(s *AssignmentContextSnapshot) { s.RequesterAdmin = true },
		"thread":         func(s *AssignmentContextSnapshot) { s.ThreadID = "other" },
		"conversation":   func(s *AssignmentContextSnapshot) { s.ConversationID = "other" },
		"inheritance":    func(s *AssignmentContextSnapshot) { s.TriggerKind = AssignmentContextTriggerAgentMention },
		"newer":          func(s *AssignmentContextSnapshot) { s.CapturedAt = s.CapturedAt.Add(time.Hour) },
		"legacy time":    func(s *AssignmentContextSnapshot) { s.CapturedAt = time.Time{} },
	} {
		t.Run(name, func(t *testing.T) {
			store := NewMemoryAssignmentContextStore()
			parent := previous
			mutate(&parent)
			for _, item := range []AssignmentContextSnapshot{first, parent} {
				if err := store.SaveAssignmentContextSnapshot(t.Context(), item); err != nil {
					t.Fatal(err)
				}
			}
			s := NewServer(ServerConfig{AssignmentContext: store})
			got, err := s.assignmentContextBaseline(t.Context(), current, first.AssignmentID)
			if name == "valid ancestor" {
				if err != nil || got.AssignmentID != first.AssignmentID {
					t.Fatal("valid ancestor rejected", err)
				}
				return
			}
			if !errors.Is(err, errAssignmentContextBaseline) {
				t.Fatal("untrusted ancestor accepted", err)
			}
		})
	}
}
