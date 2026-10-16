package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
)

// Ordinary snapshot writers preserve the committed relation. Only a scoped
// publication/removal may apply its one task/Agent delta to the fresh map.
func (s *DynamoDBAIAgentClientSnapshot) saveMembershipOwnedSnapshot(ctx context.Context, requested AIAgentClientSnapshot, credentials AWSCredentials) (dynamoDBAIAgentClientSnapshotWriteStats, int64, error) {
	mutation, scoped := taskAgentSnapshotMutation(ctx)
	for attempt := 0; attempt < 3; attempt++ {
		committed, err := s.readMembershipCommit(ctx, nil, credentials)
		if err != nil {
			return dynamoDBAIAgentClientSnapshotWriteStats{}, 0, err
		}
		current := dynamoDBAIAgentClientSnapshotCurrentItem(committed)
		oldMember := dynamoDBAIAgentClientSnapshotItemsByPart(committed)[dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships]
		if scoped && current != nil {
			var committedHashes map[string]string
			if err := json.Unmarshal([]byte(dynamoDBStringValue(current, "part_hashes_json")), &committedHashes); err != nil {
				return dynamoDBAIAgentClientSnapshotWriteStats{}, 0, err
			}
			for part, hash := range s.partHashes {
				if committedHashes[part] != hash {
					return dynamoDBAIAgentClientSnapshotWriteStats{}, 0, errTaskAgentSnapshotChanged
				}
			}
		}
		snapshot := requested
		if current != nil {
			var relation AIAgentClientSnapshot
			if err := decodeCommittedTaskAgentMemberships(current, oldMember, &relation); err != nil {
				return dynamoDBAIAgentClientSnapshotWriteStats{}, 0, err
			}
			if relation.TaskAgentMemberships != nil {
				if scoped {
					baseline := relation.TaskAgentMemberships[mutation.scope.taskID][mutation.scope.agentID]
					if !sameTaskAgentMembershipBaseline(baseline, mutation.baseline) {
						return dynamoDBAIAgentClientSnapshotWriteStats{}, 0, ErrAIAgentMembershipChanged
					}
					delta := requested.TaskAgentMemberships[mutation.scope.taskID][mutation.scope.agentID]
					snapshot.TaskAgentMemberships = copyTaskAgentMemberships(relation.TaskAgentMemberships)
					if snapshot.TaskAgentMemberships[mutation.scope.taskID] == nil {
						snapshot.TaskAgentMemberships[mutation.scope.taskID] = map[string]AIAgentTaskMembership{}
					}
					snapshot.TaskAgentMemberships[mutation.scope.taskID][mutation.scope.agentID] = delta
				} else {
					snapshot.TaskAgentMemberships = relation.TaskAgentMemberships
				}
			}
		}
		if current != nil {
			if err := s.mergeCommittedOnboardingSnapshot(committed, &snapshot); err != nil {
				return dynamoDBAIAgentClientSnapshotWriteStats{}, 0, err
			}
		}
		var committedHashes map[string]string
		if current != nil {
			if err := json.Unmarshal([]byte(dynamoDBStringValue(current, "part_hashes_json")), &committedHashes); err != nil {
				return dynamoDBAIAgentClientSnapshotWriteStats{}, 0, err
			}
		}
		items, hashes, encoded, err := encodeSplitDynamoDBAIAgentClientSnapshot(snapshot, committedHashes)
		if err != nil {
			return dynamoDBAIAgentClientSnapshotWriteStats{}, 0, err
		}
		stats := dynamoDBAIAgentClientSnapshotWritePressure(items)
		writes := items

		actions := []dynamoDBTransactWriteItem{}
		if scoped {
			condition := "attribute_not_exists(pk) AND attribute_not_exists(sk)"
			var values map[string]map[string]string
			if mutation.scope.generation != "" {
				condition = "generation = :generation"
				values = map[string]map[string]string{":generation": {"S": mutation.scope.generation}}
			}
			actions = append(actions, dynamoDBTransactWriteItem{ConditionCheck: &dynamoDBTransactWriteConditionCheckAction{TableName: s.tableName, Key: taskAgentDynamoKey(mutation.scope.taskID, mutation.scope.agentID, taskAgentGenerationSK), ConditionExpression: condition, ExpressionAttributeValues: values}})
		}
		for _, item := range writes {
			put := &dynamoDBTransactWritePutAction{TableName: s.tableName, Item: item}
			if dynamoDBStringValue(item, "sk") == dynamoDBAIAgentClientSnapshotSK {
				if current == nil {
					put.ConditionExpression = "attribute_not_exists(pk) AND attribute_not_exists(sk)"
				} else {
					put.ConditionExpression = "part_hashes_json = :snapshot_hashes"
					put.ExpressionAttributeValues = map[string]map[string]string{":snapshot_hashes": {"S": dynamoDBStringValue(current, "part_hashes_json")}}
				}
			}
			if dynamoDBAIAgentClientSnapshotPartName(item) == dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships {
				if oldMember == nil {
					put.ConditionExpression = "attribute_not_exists(pk) AND attribute_not_exists(sk)"
				} else {
					put.ConditionExpression = "part_hash = :member_hash"
					put.ExpressionAttributeValues = map[string]map[string]string{":member_hash": {"S": dynamoDBStringValue(oldMember, "part_hash")}}
				}
			}
			actions = append(actions, dynamoDBTransactWriteItem{Put: put})
		}
		if scoped {
			receipt, err := s.publicationReceiptAction(ctx, mutation, credentials)
			if err != nil {
				return stats, encoded, err
			}
			if receipt != nil {
				actions = append(actions, *receipt)
			}
		}
		payload, err := json.Marshal(struct{ TransactItems []dynamoDBTransactWriteItem }{actions})
		if err != nil {
			return stats, encoded, err
		}
		_, err = doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBTransactWriteTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
		if err == nil {
			s.partHashes = hashes
			// The caller still holds its requested cache; retain that observation as
			// the delta baseline, not values preserved from another replica.
			s.onboardingBaseline = onboardingOwnedSnapshot(requested)
			return stats, encoded, nil
		}
		if !isDynamoDBTransactionContention(err) {
			return stats, encoded, fmt.Errorf("dynamodb commit owned task Agent membership: %w", err)
		}
		if scoped {
			marker, err := s.readTaskAgentGeneration(ctx, mutation.scope, credentials)
			if err != nil {
				return stats, encoded, err
			}
			if marker != mutation.scope.generation {
				return stats, encoded, ErrAIAgentMembershipChanged
			}
			return stats, encoded, errTaskAgentSnapshotChanged
		}
	}
	return dynamoDBAIAgentClientSnapshotWriteStats{}, 0, errTaskAgentSnapshotChanged
}
