package riidoaiserver

import (
	"testing"
	"time"
)

func TestToolApprovalMatchesThreadRequiresPendingFreshSameAssignmentAgent(t *testing.T) {
	now := time.Date(2026, 7, 8, 3, 0, 0, 0, time.UTC)
	thread := AIAgentTaskThreadRecord{AssignmentID: " asn-1 ", AgentID: " agent-1 "}
	base := ToolApprovalRequest{
		AssignmentID: "asn-1",
		AgentID:      "agent-1",
		Status:       ApprovalPending,
		ExpiresAt:    now.Add(time.Minute),
	}

	tests := []struct {
		name     string
		approval ToolApprovalRequest
		want     bool
	}{
		{"pending same identity", base, true},
		{"no expiry same identity", func() ToolApprovalRequest {
			next := base
			next.ExpiresAt = time.Time{}
			return next
		}(), true},
		{"approved terminal status", func() ToolApprovalRequest {
			next := base
			next.Status = ApprovalApproved
			return next
		}(), false},
		{"expired at now", func() ToolApprovalRequest {
			next := base
			next.ExpiresAt = now
			return next
		}(), false},
		{"different assignment", func() ToolApprovalRequest {
			next := base
			next.AssignmentID = "asn-2"
			return next
		}(), false},
		{"different agent", func() ToolApprovalRequest {
			next := base
			next.AgentID = "agent-2"
			return next
		}(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toolApprovalMatchesThread(tt.approval, thread, now); got != tt.want {
				t.Fatalf("toolApprovalMatchesThread() = %v, want %v", got, tt.want)
			}
		})
	}
}
