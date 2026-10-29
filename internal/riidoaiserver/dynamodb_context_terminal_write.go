package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
)

// All assignment projection writers use this persistence boundary, including
// legacy writes, queue repairs and failed publication recovery. Only terminal
// projections add context work; the original projection conditions stay intact.
func (s *DynamoDBAssignmentOperationStore) writeAssignmentWithContextRetention(ctx context.Context, request dynamoDBRequest) ([]byte, error) {
	if !s.enableAssignmentContextRetention {
		return doDynamoDBJSON(ctx, request)
	}
	actions, terminal, err := contextTerminalProjections(request)
	if err != nil {
		return nil, err
	}
	if len(terminal) == 0 {
		return doDynamoDBJSON(ctx, request)
	}
	if len(terminal) > 1 {
		return nil, errors.New("terminal retention action budget exceeded")
	}
	if err := validateExecutionClaimTerminalBudget(actions); err != nil {
		return nil, err
	}
	var lastErr error
	for range 3 {
		writes := append([]json.RawMessage(nil), actions...)
		for _, assignment := range terminal {
			extra, err := s.terminalContextActions(ctx, assignment)
			if err != nil {
				return nil, err
			}
			for _, action := range extra {
				encoded, err := json.Marshal(action)
				if err != nil {
					return nil, err
				}
				writes = append(writes, encoded)
			}
		}
		if len(writes) > 100 {
			return nil, errors.New("context retention transaction exceeds item limit")
		}
		request.target = dynamoDBTransactWriteTarget
		request.payload, err = json.Marshal(struct{ TransactItems []json.RawMessage }{writes})
		if err != nil {
			return nil, err
		}
		body, err := doDynamoDBJSON(ctx, request)
		if err == nil || !isDynamoDBTransactionContention(err) {
			return body, err
		}
		lastErr = err
	}
	return nil, lastErr
}
