package riidoaiserver

import (
	"context"
	"errors"
)

var errTaskAgentPublicationRecoveryChanged = errors.New("failed publication recovery observation changed")

func (s *DynamoDBAssignmentOperationStore) recoverFailedTaskAgentPublication(ctx context.Context, assignment Assignment, minimumEventSeq int64) (*taskAgentPublicationRecoveryResult, error) {
	credentials, err := s.credentialsProvider.Credentials(ctx)
	if err != nil {
		return nil, err
	}
	for attempt := 0; attempt < 3; attempt++ {
		result, err := s.recoverFailedTaskAgentPublicationAttempt(ctx, assignment, minimumEventSeq, credentials)
		if !errors.Is(err, errTaskAgentPublicationRecoveryChanged) {
			return result, err
		}
	}
	return nil, ErrAIAgentMembershipChanged
}
