package main

import (
	"errors"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

const envAssignmentReadScope = "RIIDO_AI_SERVER_ASSIGNMENT_READ_SCOPE"

func assignmentReadScopeFromEnv(reader riidoaiserver.AIAgentTaskContextReader, store riidoaiserver.AssignmentContextStore) (bool, error) {
	enabled, err := envOptionalBool(envAssignmentReadScope)
	if err != nil || !enabled {
		return false, err
	}
	private, ok := reader.(*riidoaiserver.AIAgentPrivateTaskContextClient)
	if !ok || !private.AssignmentReadScopeEnabled() {
		return false, errors.New("assignment read scope requires a requester-private scoped reader")
	}
	if _, ok := store.(riidoaiserver.AssignmentContentStore); !ok {
		return false, errors.New("assignment read scope requires atomic assignment content storage")
	}
	return true, nil
}

func assignmentReadScopeRuntime(reader riidoaiserver.AIAgentTaskContextReader, operations riidoaiserver.AssignmentOperationStore) (riidoaiserver.AssignmentContextStore, bool, error) {
	store, err := assignmentContextStoreFromEnv(operations)
	if err != nil {
		return nil, false, err
	}
	enabled, err := assignmentReadScopeFromEnv(reader, store)
	return store, enabled, err
}
