package riidoaiserver

import (
	"errors"
)

func (s *DynamoDBAssignmentOperationStore) loop() {
	defer close(s.done)
	var cachedCredentials AWSCredentials
	for cmd := range s.commands {
		if cmd.close {
			cmd.reply <- nil
			return
		}
		credentials, err := cachedAWSCredentials(cmd.ctx, s.now, s.credentialsProvider, &cachedCredentials)
		if err != nil {
			switch {
			case cmd.load:
				cmd.loadReply <- dynamoDBAssignmentOperationLoadResult{err: err}
			case cmd.queue:
				cmd.queueReply <- dynamoDBAssignmentQueueResult{err: err}
			case cmd.claim:
				cmd.claimReply <- dynamoDBAssignmentClaimResult{err: err}
			case cmd.active:
				cmd.activeReply <- dynamoDBAssignmentActiveLeaseResult{err: err}
			case cmd.projection:
				cmd.projectionReply <- dynamoDBAssignmentProjectionResult{err: err}
			default:
				cmd.reply <- err
			}
			continue
		}
		if cmd.load {
			records, err := s.load(cmd.ctx, credentials)
			cmd.loadReply <- dynamoDBAssignmentOperationLoadResult{records: records, err: err}
			continue
		}
		if cmd.queue {
			assignments, err := s.loadAgentQueue(cmd.ctx, cmd.agentID, credentials)
			cmd.queueReply <- dynamoDBAssignmentQueueResult{assignments: assignments, err: err}
			continue
		}
		if cmd.claim {
			result, err := s.claimNext(cmd.ctx, cmd.agentID, cmd.claimAt, credentials)
			cmd.claimReply <- dynamoDBAssignmentClaimResult{result: result, err: err}
			continue
		}
		if cmd.active {
			lease, found, err := s.loadAgentActiveAssignment(cmd.ctx, cmd.agentID, credentials)
			cmd.activeReply <- dynamoDBAssignmentActiveLeaseResult{lease: lease, found: found, err: err}
			continue
		}
		if cmd.projection {
			record, found, err := s.loadAssignmentProjection(cmd.ctx, cmd.assignmentID, credentials)
			cmd.projectionReply <- dynamoDBAssignmentProjectionResult{projection: AssignmentProjection(record), found: found, err: err}
			continue
		}
		if cmd.refresh {
			if cmd.assignment == nil {
				cmd.reply <- errors.New("riidoaiserver: nil DynamoDB active assignment refresh")
				continue
			}
			cmd.reply <- s.refreshAgentActiveAssignment(cmd.ctx, *cmd.assignment, cmd.refreshAt, credentials)
			continue
		}
		if cmd.record == nil {
			cmd.reply <- errors.New("riidoaiserver: nil DynamoDB assignment operation")
			continue
		}
		cmd.reply <- s.save(cmd.ctx, *cmd.record, credentials)
	}
}
