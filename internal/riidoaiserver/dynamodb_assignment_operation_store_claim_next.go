package riidoaiserver

import (
	"context"
	"fmt"
	"time"
)

func (s *DynamoDBAssignmentOperationStore) claimNext(ctx context.Context, agentID string, at time.Time, credentials AWSCredentials) (AssignmentClaimResult, error) {
	queue, err := s.loadAgentQueueRecords(ctx, agentID, credentials)
	if err != nil {
		return AssignmentClaimResult{}, err
	}
	for _, candidate := range queue {
		assignment := candidate.Assignment
		if assignment.State.Code() != AssignmentStateCodeQueued {
			continue
		}
		if assignmentQueuedPastMaxAge(assignment, at) {
			repair := dynamoDBAssignmentClaimRepair{
				Operation:            dynamoDBFailStaleQueuedOperation(candidate, at),
				ExpectedState:        AssignmentQueued,
				ExpectedLastEventSeq: candidate.LastEventSeq,
			}
			payload, err := s.claimRepairOnlyTransactionPayload(repair)
			if err != nil {
				return AssignmentClaimResult{}, err
			}
			_, err = s.writeAssignmentWithContextRetention(ctx, dynamoDBRequest{
				endpoint:     s.endpoint,
				endpointHost: s.endpointHost,
				region:       s.region,
				target:       dynamoDBTransactWriteTarget,
				payload:      payload,
				credentials:  credentials,
				httpClient:   s.httpClient,
				now:          s.now,
			})
			if err == nil || isDynamoDBTransactionContention(err) {
				continue
			}
			return AssignmentClaimResult{}, fmt.Errorf("dynamodb repair stale queued assignment: %w", err)
		}
		var repairs []dynamoDBAssignmentClaimRepair
		var activeCondition *dynamoDBTransactWriteConditionCheckAction
		clearMessage := ""
		clearMetadata := map[string]string(nil)
		if assignment.BlockedByAssignmentID != "" {
			blocker, ok, err := s.loadAssignmentProjection(ctx, assignment.BlockedByAssignmentID, credentials)
			if err != nil {
				return AssignmentClaimResult{}, err
			}
			clearMetadata = map[string]string{"blocked_by_assignment_id": assignment.BlockedByAssignmentID}
			switch {
			case !ok:
				clearMessage = "missing blocker cleared before daemon lease"
			case isTerminal(blocker.Assignment.State):
				clearMessage = "terminal blocker cleared before daemon lease"
			case blocker.Assignment.State.Code() == AssignmentStateCodeQueued:
				repairs = append(repairs, dynamoDBAssignmentClaimRepair{
					Operation:            dynamoDBCancelQueuedBlockerOperation(blocker, assignment.ID, at),
					ExpectedState:        AssignmentQueued,
					ExpectedLastEventSeq: blocker.LastEventSeq,
				})
				clearMessage = "queued blocker cleared before daemon lease"
			case assignmentHoldsActiveLease(blocker.Assignment.State):
				stale, condition, err := s.staleBlockerLeaseCondition(ctx, blocker.Assignment, at, credentials)
				if err != nil {
					return AssignmentClaimResult{}, err
				}
				if !stale {
					continue
				}
				activeCondition = condition
				repairs = append(repairs, dynamoDBAssignmentClaimRepair{
					Operation:            dynamoDBFailStaleBlockerOperation(blocker, assignment.ID, at),
					ExpectedState:        blocker.Assignment.State,
					ExpectedLastEventSeq: blocker.LastEventSeq,
				})
				clearMessage = "stale blocker cleared before daemon lease"
			default:
				continue
			}
		}
		operation := dynamoDBClaimAssignmentOperation(assignment, candidate.LastEventSeq, at, clearMessage, clearMetadata)
		operation.ExecutionClaim = newExecutionClaim(ctx, operation.Assignment)
		if err := validateOperationExecutionClaim(operation); err != nil {
			return AssignmentClaimResult{}, err
		}
		payload, err := s.claimTransactionPayload(operation, candidate.LastEventSeq)
		if len(repairs) > 0 || activeCondition != nil {
			payload, err = s.claimRepairTransactionPayload(operation, candidate.LastEventSeq, repairs, activeCondition)
		}
		if err != nil {
			return AssignmentClaimResult{}, err
		}
		_, err = s.writeAssignmentWithContextRetention(ctx, dynamoDBRequest{
			endpoint:     s.endpoint,
			endpointHost: s.endpointHost,
			region:       s.region,
			target:       dynamoDBTransactWriteTarget,
			payload:      payload,
			credentials:  credentials,
			httpClient:   s.httpClient,
			now:          s.now,
		})
		if err == nil {
			operations := make([]AssignmentOperationRecord, 0, len(repairs)+1)
			for _, repair := range repairs {
				operations = append(operations, repair.Operation)
			}
			operations = append(operations, operation)
			return AssignmentClaimResult{Claimed: true, Assignment: operation.Assignment, Operation: operation, Operations: operations}, nil
		}
		if isDynamoDBTransactionContention(err) {
			continue
		}
		return AssignmentClaimResult{}, fmt.Errorf("dynamodb claim assignment: %w", err)
	}
	return AssignmentClaimResult{}, nil
}
