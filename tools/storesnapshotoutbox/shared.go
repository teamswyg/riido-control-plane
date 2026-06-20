package main

import (
	"context"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func assignAndPoll(ctx context.Context, store *riidoaiserver.Store) (riidoaiserver.Assignment, error) {
	assignment, err := store.AssignTask(ctx, "task-a", riidoaiserver.AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "make hello world",
	})
	if err != nil {
		return riidoaiserver.Assignment{}, err
	}
	_, err = store.PollAgent(ctx, "agent-1", pollRequest())
	return assignment, err
}

func pollRequest() riidoaiserver.PollRequest {
	return riidoaiserver.PollRequest{DaemonID: "daemon-1", RuntimeID: "runtime-1"}
}
