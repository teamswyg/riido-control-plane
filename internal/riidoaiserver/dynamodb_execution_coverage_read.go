package riidoaiserver

import (
	"context"
	"strings"
)

func (s *DynamoDBAssignmentOperationStore) loadExecutionCoverage(ctx context.Context, device string) (executionCoverage, bool, error) {
	item, err := s.executionTerminationGet(ctx, "DEVICE_POLICY#"+device, "REVISION")
	if err != nil {
		return executionCoverage{}, false, err
	}
	if dynamoDBStringValue(item, "policy_schema") == executionCoverageSchema {
		coverage, err := executionCoverageFromItem(item, device)
		return coverage, true, err
	}
	// A missing/corrupt security marker cannot turn a server-born device legacy.
	if strings.HasPrefix(device, executionCoveragePrefix) {
		return executionCoverage{}, true, errCLIInstallationUnavailable
	}
	_, err = devicePolicyRevision(item, device)
	return executionCoverage{}, false, err
}
