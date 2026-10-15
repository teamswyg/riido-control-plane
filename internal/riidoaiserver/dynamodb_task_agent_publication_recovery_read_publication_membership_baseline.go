package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
)

func (s *DynamoDBAssignmentOperationStore) readPublicationMembershipBaseline(ctx context.Context, assignment Assignment, credentials AWSCredentials) (AIAgentTaskMembership, string, error) {
	type action struct {
		Get struct {
			TableName string
			Key       map[string]map[string]string
		}
	}
	actions := make([]action, 2)
	for i, sk := range []string{dynamoDBAIAgentClientSnapshotSK, dynamoDBAIAgentClientSnapshotPartSK(dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships)} {
		actions[i].Get.TableName = s.tableName
		actions[i].Get.Key = map[string]map[string]string{"pk": {"S": dynamoDBAIAgentClientSnapshotPK}, "sk": {"S": sk}}
	}
	payload, err := json.Marshal(struct{ TransactItems []action }{actions})
	if err != nil {
		return AIAgentTaskMembership{}, "", err
	}
	body, err := doDynamoDBJSON(ctx, dynamoDBRequest{endpoint: s.endpoint, endpointHost: s.endpointHost, region: s.region, target: dynamoDBSnapshotTransactGetTarget, payload: payload, credentials: credentials, httpClient: s.httpClient, now: s.now})
	if err != nil {
		return AIAgentTaskMembership{}, "", err
	}
	var response struct {
		Responses []struct{ Item map[string]map[string]string }
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return AIAgentTaskMembership{}, "", err
	}
	if len(response.Responses) != 2 {
		return AIAgentTaskMembership{}, "", fmt.Errorf("publication membership baseline is incomplete")
	}
	var snapshot AIAgentClientSnapshot
	if err := decodeCommittedTaskAgentMemberships(response.Responses[0].Item, response.Responses[1].Item, &snapshot); err != nil {
		return AIAgentTaskMembership{}, "", err
	}
	return snapshot.TaskAgentMemberships[assignment.TaskID][assignment.AgentID], dynamoDBStringValue(response.Responses[1].Item, "part_hash"), nil
}
