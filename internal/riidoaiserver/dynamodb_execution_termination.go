package riidoaiserver

import (
	"context"
)

const executionTerminationItemSchema = "execution_termination/v1"

type executionTerminationReadSet struct {
	claim      map[string]map[string]string
	projection map[string]map[string]string
	receipt    map[string]map[string]string
	value      executionTerminationReceipt
}

func (s *DynamoDBAssignmentOperationStore) recordExecutionTermination(ctx context.Context, id executionTerminationIdentity, req executionTerminationRequest) (executionTerminationReceipt, error) {
	var lastErr error
	for range 3 {
		if err := ctx.Err(); err != nil {
			return executionTerminationReceipt{}, err
		}
		set, err := s.loadExecutionTerminationReadSet(ctx, id)
		if err != nil {
			return executionTerminationReceipt{}, err
		}
		if len(set.receipt) > 0 {
			if _, err := matchingExecutionTermination(set.value, req); err != nil {
				return executionTerminationReceipt{}, err
			}
		} else {
			set.value = newExecutionTermination(id, req, s.now())
		}
		actions, err := s.executionTerminationActions(id, set)
		if err != nil {
			return executionTerminationReceipt{}, err
		}
		_, lastErr = s.assignmentContextRequest(ctx, dynamoDBTransactWriteTarget, struct{ TransactItems []dynamoDBTransactWriteItem }{actions})
		if lastErr == nil {
			return set.value, nil
		}
		// Even a failed response may follow a committed transaction. Reload the
		// original receipt and all guards; never overwrite or infer an ACK.
	}
	return executionTerminationReceipt{}, lastErr
}
