package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
)

func (s *DynamoDBAssignmentOperationStore) recoverFailedTaskAgentPublicationAttempt(ctx context.Context, assignment Assignment, minimumEventSeq int64, credentials AWSCredentials) (*taskAgentPublicationRecoveryResult, error) {
	marker, _, err := s.getTaskAgentItem(ctx, taskAgentDynamoKey(assignment.TaskID, assignment.AgentID, taskAgentGenerationSK), credentials)
	if err != nil {
		return nil, err
	}
	observedGeneration := dynamoDBStringValue(marker, "generation")
	receipt, _, err := s.getTaskAgentItem(ctx, taskAgentDynamoKey(assignment.TaskID, assignment.AgentID, taskAgentExecutionReceiptSK(assignment)), credentials)
	if err != nil {
		return nil, err
	}
	if dynamoDBStringValue(receipt, "publication_committed") == "true" {
		return nil, nil
	}
	member, memberHash, err := s.readPublicationMembershipBaseline(ctx, assignment, credentials)
	if err != nil {
		return nil, err
	}
	projection, found, err := s.loadAssignmentProjection(ctx, assignment.ID, credentials)
	if err != nil || !found {
		return nil, err
	}
	if !sameTaskAgentCreation(projection.Assignment, assignment) {
		return nil, ErrAIAgentMembershipChanged
	}
	result := &taskAgentPublicationRecoveryResult{projection: AssignmentProjection(projection)}
	// A committed publication wins even if its HTTP response was lost.
	if member.DurableGeneration == taskAgentAssignmentGeneration(assignment) {
		return result, nil
	}
	operation := failedPublicationCancellation(projection, minimumEventSeq, s.now())
	actions, err := s.publicationRecoveryActions(assignment, member, memberHash, observedGeneration, projection, operation)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(struct{ TransactItems []dynamoDBTransactWriteItem }{actions})
	if err != nil {
		return nil, err
	}
	_, err = doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBTransactWriteTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
	if isDynamoDBTransactionContention(err) {
		return nil, errTaskAgentPublicationRecoveryChanged
	}
	if err != nil {
		return nil, fmt.Errorf("dynamodb recover failed task Agent publication: %w", err)
	}
	if operation != nil {
		result.operation = operation
		result.projection = AssignmentProjection{Assignment: operation.Assignment, LastEventSeq: assignmentOperationLastEventSeq(*operation)}
	}
	return result, nil
}
