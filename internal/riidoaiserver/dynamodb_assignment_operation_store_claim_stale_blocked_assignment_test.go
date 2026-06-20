package riidoaiserver

import "time"

func staleBlockedCurrentAssignment() Assignment {
	now := time.Date(2026, 6, 9, 4, 0, 0, 0, time.UTC)
	return Assignment{
		ID:                    "asn-000002",
		TaskID:                "task-a",
		ComponentID:           "component-1",
		AgentID:               "jykim-new",
		RuntimeProvider:       "codex",
		Prompt:                "new work",
		State:                 AssignmentQueued,
		ReplacesAssignmentID:  "asn-000001",
		BlockedByAssignmentID: "asn-000001",
		CreatedAt:             now.Add(-2 * time.Minute),
		UpdatedAt:             now.Add(-2 * time.Minute),
	}
}

func staleBlockedBlockerAssignment() Assignment {
	now := time.Date(2026, 6, 9, 4, 0, 0, 0, time.UTC)
	return Assignment{
		ID:              "asn-000001",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim-old",
		RuntimeProvider: "codex",
		Prompt:          "old work",
		State:           AssignmentRunning,
		LeaseToken:      "lease-old",
		CreatedAt:       now.Add(-3 * time.Minute),
		UpdatedAt:       now.Add(-3 * time.Minute),
	}
}
