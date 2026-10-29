package main

import (
	"fmt"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

const envExecutionCoverageEnrollment = "RIIDO_AI_SERVER_EXECUTION_COVERAGE_ENROLLMENT"

func executionCoverageEnrollmentFromEnv(snapshot riidoaiserver.AIAgentClientSnapshotStore, assignments riidoaiserver.AssignmentOperationStore) (bool, error) {
	enabled, err := envOptionalBool(envExecutionCoverageEnrollment)
	if err != nil || !enabled {
		return enabled, err
	}
	if !riidoaiserver.ExecutionCoverageAuthorityCompatible(snapshot, assignments) {
		return false, fmt.Errorf("%s requires native snapshot and assignment stores on the same Dynamo authority", envExecutionCoverageEnrollment)
	}
	return true, nil
}

func executionCoverageAssignmentStoreFromEnv(parts runtimeConfigParts, activeLease time.Duration) (riidoaiserver.AssignmentOperationStore, error) {
	enabled, err := envOptionalBool(envExecutionCoverageEnrollment)
	if err != nil {
		return nil, err
	}
	if enabled {
		return riidoaiserver.NewExecutionCoverageAssignmentStore(parts.aiAgentClientStore, activeLease)
	}
	return assignmentOperationStoreFromEnv(parts.aiAgentClientDev, activeLease)
}
