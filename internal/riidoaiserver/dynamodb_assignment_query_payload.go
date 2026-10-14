package riidoaiserver

import "encoding/json"

func (s *DynamoDBAssignmentOperationStore) queryPayload(exclusiveStartKey map[string]map[string]string) ([]byte, error) {
	payload := struct {
		TableName                 string                       `json:"TableName"`
		ConsistentRead            bool                         `json:"ConsistentRead"`
		KeyConditionExpression    string                       `json:"KeyConditionExpression"`
		ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
		ExclusiveStartKey         map[string]map[string]string `json:"ExclusiveStartKey,omitempty"`
		ScanIndexForward          bool                         `json:"ScanIndexForward"`
		Limit                     int                          `json:"Limit"`
	}{
		TableName:              s.tableName,
		ConsistentRead:         true,
		KeyConditionExpression: "pk = :pk",
		ExpressionAttributeValues: map[string]map[string]string{
			":pk": {"S": dynamoDBAssignmentOperationPK},
		},
		ExclusiveStartKey: exclusiveStartKey,
		ScanIndexForward:  true,
		Limit:             dynamoDBAssignmentQueryLimit,
	}
	return json.Marshal(payload)
}

func (s *DynamoDBAssignmentOperationStore) agentQueueQueryPayload(agentID string, exclusiveStartKey map[string]map[string]string) ([]byte, error) {
	payload := struct {
		TableName                 string                       `json:"TableName"`
		IndexName                 string                       `json:"IndexName"`
		KeyConditionExpression    string                       `json:"KeyConditionExpression"`
		ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
		ExclusiveStartKey         map[string]map[string]string `json:"ExclusiveStartKey,omitempty"`
		ScanIndexForward          bool                         `json:"ScanIndexForward"`
		Limit                     int                          `json:"Limit"`
	}{
		TableName:              s.tableName,
		IndexName:              dynamoDBAssignmentQueueIndex,
		KeyConditionExpression: "agent_id = :agent_id",
		ExpressionAttributeValues: map[string]map[string]string{
			":agent_id": {"S": agentID},
		},
		ExclusiveStartKey: exclusiveStartKey,
		ScanIndexForward:  true,
		Limit:             dynamoDBAssignmentQueryLimit,
	}
	return json.Marshal(payload)
}
