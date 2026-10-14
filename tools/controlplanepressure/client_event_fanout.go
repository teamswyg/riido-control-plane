package main

import (
	"context"
	"sync/atomic"

	srv "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func buildClientSubscriberFanout(cfg config) (pressureOperation, error) {
	ctx := context.Background()
	store := srv.NewDevelopmentAIAgentClientStore()
	principal := srv.AuthorizationResult{PrincipalID: "user-1", WorkspaceID: fixtureWorkspaceID}
	for range cfg.Threads {
		if _, _, _, err := store.SubscribeAIAgentClientEvents(ctx, principal); err != nil {
			return pressureOperation{}, err
		}
	}
	assigned, err := store.AssignAIAgentTask(ctx, principal, fixtureTaskID, srv.AssignAIAgentTaskRequest{
		AgentID: fixtureAgentID,
	})
	if err != nil {
		return pressureOperation{}, err
	}
	var seq atomic.Int64
	return newPressureOperation(func() error {
		next := int(seq.Add(1))
		_, err := store.RecordAIAgentThreadProgress(ctx, assigned.AgentID, srv.AgentThreadProgressBatchRequest{
			AssignmentID: assigned.AssignmentID,
			TaskID:       assigned.TaskID,
			ThreadID:     assigned.ThreadID,
			RunID:        assigned.RunID,
			Lines: []srv.AgentThreadProgressLine{{
				Seq:     next,
				Message: progressFragment(next),
			}},
		})
		return err
	}), nil
}
