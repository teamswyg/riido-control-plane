package riidoaiserver

import "context"

func (s *DynamoDBAssignmentOperationStore) cliInstallationTransferCoverage(ctx context.Context, intent cliInstallationIntent, device authenticatedClaimDevice) (executionCoverage, executionCoverage, error) {
	source, foundSource, err := s.loadExecutionCoverage(ctx, intent.Request.PreviousDeviceID)
	if err != nil || !foundSource {
		return source, executionCoverage{}, errCLIInstallationUnavailable
	}
	target, foundTarget, err := s.loadExecutionCoverage(ctx, device.deviceID)
	if err != nil || !foundTarget || s.executionCoverageAuthority == nil {
		return source, target, errCLIInstallationUnavailable
	}
	wantSourceState := "complete"
	if intent.Consumed {
		wantSourceState = "revoked"
	}
	if source.State != wantSourceState || source.Unresolved != 0 || target.State != "complete" || (!intent.Consumed && target.Unresolved != 0) {
		return source, target, errCLIInstallationUnavailable
	}
	if source.OwnerPrincipalID != device.ownerPrincipalID || target.OwnerPrincipalID != device.ownerPrincipalID || target.Generation != device.coverage.generation || target.CredentialGeneration != device.coverage.credential {
		return source, target, ErrAuthorizationForbidden
	}
	// The same native authority supplies enrollment and admission. A credential
	// rotation changes the coverage CAS, including while the snapshot is read.
	snapshot, found, err := s.executionCoverageAuthority.LoadAIAgentClientSnapshot(ctx)
	if err != nil || !found {
		return source, target, errCLIInstallationUnavailable
	}
	sourceWorkspace, targetWorkspace := "", ""
	for _, c := range snapshot.DeviceCredentials {
		if currentTransferCredential(c, source) {
			sourceWorkspace = c.WorkspaceID
		}
		if currentTransferCredential(c, target) {
			targetWorkspace = c.WorkspaceID
		}
	}
	if sourceWorkspace == "" || sourceWorkspace != targetWorkspace {
		return source, target, ErrAuthorizationForbidden
	}
	return source, target, nil
}

func currentTransferCredential(c AIAgentClientDeviceCredentialSnapshot, coverage executionCoverage) bool {
	return c.DeviceID == coverage.DeviceID && c.OwnerPrincipalID == coverage.OwnerPrincipalID && c.ExecutionCoverageGeneration == coverage.Generation && c.CredentialGeneration == coverage.CredentialGeneration
}
