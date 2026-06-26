package main

import (
	"context"
	"time"

	srv "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func buildToolApprovalWaiters(config) (pressureOperation, error) {
	ctx := context.Background()
	store := srv.NewStore()
	assigned, err := store.AssignTask(ctx, "task-approval-pressure", srv.AssignRequest{
		ComponentID: "component-approval-pressure",
		AgentID:     "agent-approval-pressure", RuntimeProvider: "codex",
		Prompt: "approval pressure",
	})
	if err != nil {
		store.Close()
		return pressureOperation{}, err
	}
	approval, err := store.CreateToolApproval(ctx, assigned.AgentID, toolApprovalPressureRequest(assigned))
	if err != nil {
		store.Close()
		return pressureOperation{}, err
	}
	return pressureOperation{
		run: func() error {
			_, _, err := store.WaitForToolApproval(ctx, assigned.AgentID, assigned.ID, approval.ApprovalID, 2*time.Millisecond, time.Millisecond)
			return err
		},
		cleanup: func() {
			store.Close()
		},
	}, nil
}

func toolApprovalPressureRequest(assignment srv.Assignment) srv.ToolApprovalRequest {
	return srv.ToolApprovalRequest{
		ApprovalID: "approval-pressure", AssignmentID: assignment.ID,
		TaskID: assignment.TaskID, AgentID: assignment.AgentID,
		ToolID: "tool-pressure", ToolKind: "patch_apply", ToolName: "apply_patch",
		Reason: "pressure harness pending approval",
		Status: srv.ApprovalPending,
	}
}
