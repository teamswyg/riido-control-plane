package riidoaiserver

import (
	"strings"
	"testing"
	"time"
)

func TestValidateStoredToolApprovalRequiresStableIdentity(t *testing.T) {
	now := time.Date(2026, 7, 7, 10, 0, 0, 0, time.UTC)
	valid := ToolApprovalRequest{
		ApprovalID:   "approval-1",
		AssignmentID: "asn-1",
		TaskID:       "task-1",
		AgentID:      "agent-1",
		ToolID:       "tool-1",
		Status:       ApprovalPending,
		RequestedAt:  now,
		ExpiresAt:    now.Add(time.Minute),
	}
	if err := validateStoredToolApproval(valid); err != nil {
		t.Fatalf("valid approval rejected: %v", err)
	}

	cases := []struct {
		name   string
		mutate func(*ToolApprovalRequest)
		want   string
	}{
		{"approval", func(r *ToolApprovalRequest) { r.ApprovalID = " " }, "approval_id is required"},
		{"assignment", func(r *ToolApprovalRequest) { r.AssignmentID = "" }, "assignment_id is required"},
		{"task", func(r *ToolApprovalRequest) { r.TaskID = "" }, "task_id is required"},
		{"agent", func(r *ToolApprovalRequest) { r.AgentID = "" }, "agent_id is required"},
		{"tool", func(r *ToolApprovalRequest) { r.ToolID = "" }, "tool_id is required"},
		{"status", func(r *ToolApprovalRequest) { r.Status = "" }, "status is required"},
		{"requested", func(r *ToolApprovalRequest) { r.RequestedAt = time.Time{} }, "requested_at is required"},
		{"expires", func(r *ToolApprovalRequest) { r.ExpiresAt = time.Time{} }, "expires_at is required"},
		{
			"backwards expiry",
			func(r *ToolApprovalRequest) { r.ExpiresAt = now.Add(-time.Second) },
			"expires_at must be after requested_at",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := valid
			tc.mutate(&req)
			if err := validateStoredToolApproval(req); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("validateStoredToolApproval error = %v, want %q", err, tc.want)
			}
		})
	}
}
