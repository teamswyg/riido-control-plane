package main

import "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"

// assignmentContextStoreFromEnv enables Riido CLI assignment context
// snapshots. Off (the default) leaves assignment prompts and daemon routes as
// they were. On, snapshots share the assignment DynamoDB table when one is
// configured and otherwise live in process memory for local development.
func assignmentContextStoreFromEnv(operations riidoaiserver.AssignmentOperationStore) (riidoaiserver.AssignmentContextStore, error) {
	enabled, err := envOptionalBool(envAssignmentContext)
	if err != nil || !enabled {
		return nil, err
	}
	if store, ok := operations.(riidoaiserver.AssignmentContextStore); ok {
		return store, nil
	}
	return riidoaiserver.NewMemoryAssignmentContextStore(), nil
}
