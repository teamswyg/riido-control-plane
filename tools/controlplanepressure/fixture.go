package main

import (
	"context"
	"fmt"

	srv "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

const (
	fixtureWorkspaceID = "workspace-dev-riid"
	fixtureTaskID      = "task-pressure-local"
	fixtureAgentID     = "agent-owned-codex"
)

func pressureFixture(cfg config) (*srv.DevelopmentAIAgentClientStore, srv.AuthorizationResult, string, error) {
	ctx := context.Background()
	store := srv.NewDevelopmentAIAgentClientStore()
	principal := srv.AuthorizationResult{PrincipalID: "user-1", WorkspaceID: fixtureWorkspaceID}
	for i := range cfg.Threads {
		assigned, err := store.AssignAIAgentTask(ctx, principal, fixtureTaskID, srv.AssignAIAgentTaskRequest{
			AgentID: fixtureAgentID,
		})
		if err != nil {
			return nil, principal, "", err
		}
		if err := appendProgress(ctx, store, assigned, i, cfg.Lines); err != nil {
			return nil, principal, "", err
		}
	}
	return store, principal, fixtureTaskID, nil
}

func appendProgress(ctx context.Context, store *srv.DevelopmentAIAgentClientStore, assigned srv.AIAgentTaskActionResponse, thread, lines int) error {
	for seq := 1; seq <= lines; seq++ {
		_, err := store.RecordAIAgentThreadProgress(ctx, assigned.AgentID, srv.AgentThreadProgressBatchRequest{
			AssignmentID: assigned.AssignmentID,
			TaskID:       assigned.TaskID,
			ThreadID:     assigned.ThreadID,
			RunID:        assigned.RunID,
			Lines: []srv.AgentThreadProgressLine{{
				Seq:     seq,
				Message: fmt.Sprintf("thread=%d line=%d", thread, seq),
			}},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
