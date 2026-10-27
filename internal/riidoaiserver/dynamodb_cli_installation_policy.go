package riidoaiserver

import "context"

func (s *DynamoDBAssignmentOperationStore) readCLIInstallationPolicy(ctx context.Context, id cliInstallationIdentity, owners cliInstallationDeviceOwnerReader) (cliInstallationPolicy, error) {
	for range 3 {
		if err := ctx.Err(); err != nil {
			return cliInstallationPolicy{}, err
		}
		link, item, err := s.loadCLIInstallationLink(ctx, id)
		if err != nil {
			return cliInstallationPolicy{}, err
		}
		set := cliInstallationReadSet{link: link, found: len(item) > 0, observation: deviceExecutionUnknown}
		actions := []dynamoDBTransactWriteItem{{ConditionCheck: s.installationItemCondition(item, cliInstallationLinkKey(id))}}
		owner := ""
		if set.found {
			key := devicePolicyKey(link.DeviceID)
			revision, err := s.executionTerminationGet(ctx, key["pk"]["S"], key["sk"]["S"])
			if err != nil {
				return cliInstallationPolicy{}, err
			}
			set.revision, err = devicePolicyRevision(revision, link.DeviceID)
			if err != nil || set.revision == 0 {
				return cliInstallationPolicy{}, errCLIInstallationUnavailable
			}
			owner, err = owners.cliInstallationDeviceOwner(ctx, link.DeviceID)
			if err != nil {
				return cliInstallationPolicy{}, err
			}
			set.observation, err = s.installationExecutionObservation(ctx, link.DeviceID)
			if err != nil {
				return cliInstallationPolicy{}, err
			}
			actions = append(actions, s.devicePolicyFence(link.DeviceID, set.revision))
		}
		_, err = s.assignmentContextRequest(ctx, dynamoDBTransactWriteTarget, struct{ TransactItems []dynamoDBTransactWriteItem }{actions})
		if err == nil {
			return installationPolicyFromRead(set, owner, s.now()), nil
		}
	}
	return cliInstallationPolicy{}, errCLIInstallationUnavailable
}

func (s *DynamoDBAssignmentOperationStore) installationExecutionObservation(ctx context.Context, device string) (deviceExecutionObservation, error) {
	claims, err := s.loadDeviceExecutionClaims(ctx, device)
	if err != nil {
		return deviceExecutionUnknown, err
	}
	for _, claim := range claims {
		projection, found, err := s.LoadAssignmentProjection(ctx, claim.AssignmentID)
		if err != nil {
			return deviceExecutionUnknown, err
		}
		if found && claimMatchesNonOwnerAssignment(claim, projection.Assignment) {
			return deviceExecutionBlocked, nil
		}
	}
	return deviceExecutionUnknown, nil
}
