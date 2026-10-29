package riidoaiserver

import (
	"context"
	"encoding/json"
)

func (s *DynamoDBAIAgentClientSnapshot) readEnrollmentCoverage(ctx context.Context, device string, credentials AWSCredentials) (executionCoverage, error) {
	payload, err := json.Marshal(struct {
		TableName      string                       `json:"TableName"`
		ConsistentRead bool                         `json:"ConsistentRead"`
		Key            map[string]map[string]string `json:"Key"`
	}{s.tableName, true, map[string]map[string]string{"pk": {"S": "DEVICE_POLICY#" + device}, "sk": {"S": "REVISION"}}})
	if err != nil {
		return executionCoverage{}, err
	}
	body, err := doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBGetItemTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
	if err != nil {
		return executionCoverage{}, err
	}
	var result struct {
		Item map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return executionCoverage{}, err
	}
	return executionCoverageFromItem(result.Item, device)
}

func (s *DynamoDBAIAgentClientSnapshot) verifyEnrollmentCoverage(ctx context.Context, response EnrollDeviceResponse) error {
	credentials, err := s.desktopPresenceCredentials(ctx)
	if err != nil {
		return err
	}
	coverage, err := s.readEnrollmentCoverage(ctx, response.DeviceID, credentials)
	if err != nil {
		return err
	}
	if coverage.Generation != response.ExecutionCoverageGeneration || coverage.CredentialGeneration != response.CredentialGeneration || coverage.OwnerPrincipalID != response.OwnerPrincipalID || coverage.State == "revoked" {
		return errCLIInstallationUnavailable
	}
	return nil
}

func (s *DynamoDBAIAgentClientSnapshot) enrollmentCoverageAction(ctx context.Context, snapshot AIAgentClientSnapshot, credentials AWSCredentials) (*dynamoDBTransactWriteItem, error) {
	mutation, scoped := ctx.Value(deviceEnrollmentMutationContextKey{}).(deviceEnrollmentMutation)
	if !scoped {
		return nil, nil
	}
	after, ok := snapshotDeviceCredential(snapshot, mutation.deviceID)
	if !ok {
		return nil, errAIAgentSnapshotConflict
	}
	if err := validateCredentialCoverage(after); err != nil {
		return nil, err
	}
	if after.ExecutionCoverageGeneration == "" {
		if mutation.before.ExecutionCoverageGeneration != "" {
			return nil, errCLIInstallationUnavailable
		}
		return nil, nil
	}
	if !mutation.existed {
		item, err := executionCoverageItem(executionCoverage{DeviceID: after.DeviceID, Generation: after.ExecutionCoverageGeneration, CredentialGeneration: after.CredentialGeneration, OwnerPrincipalID: after.OwnerPrincipalID, State: "complete", Revision: 1})
		if err != nil {
			return nil, err
		}
		return &dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{TableName: s.tableName, Item: item, ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)"}}, nil
	}
	before, err := s.readEnrollmentCoverage(ctx, after.DeviceID, credentials)
	if err != nil {
		return nil, err
	}
	if before.Generation != mutation.before.ExecutionCoverageGeneration || before.CredentialGeneration != mutation.before.CredentialGeneration || before.OwnerPrincipalID != mutation.before.OwnerPrincipalID || before.Generation != after.ExecutionCoverageGeneration || before.OwnerPrincipalID != after.OwnerPrincipalID || before.State == "revoked" {
		return nil, errCLIInstallationUnavailable
	}
	action, err := executionCoverageChangeAction(s.tableName, before, 0, after.CredentialGeneration)
	return &action, err
}
