package riidoaiserver

import (
	"strconv"
)

func (s *DynamoDBAssignmentOperationStore) publicationRecoveryActions(assignment Assignment, member AIAgentTaskMembership, memberHash, observedGeneration string, projection assignmentProjectionRecord, operation *AssignmentOperationRecord) ([]dynamoDBTransactWriteItem, error) {
	memberCondition := &dynamoDBTransactWriteConditionCheckAction{TableName: s.tableName, Key: map[string]map[string]string{"pk": {"S": dynamoDBAIAgentClientSnapshotPK}, "sk": {"S": dynamoDBAIAgentClientSnapshotPartSK(dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships)}}, ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)"}
	if memberHash != "" {
		memberCondition.ConditionExpression = "part_hash = :member_hash"
		memberCondition.ExpressionAttributeValues = map[string]map[string]string{":member_hash": {"S": memberHash}}
	}
	actions := []dynamoDBTransactWriteItem{{ConditionCheck: memberCondition}}
	markerKey := taskAgentDynamoKey(assignment.TaskID, assignment.AgentID, taskAgentGenerationSK)
	values := map[string]map[string]string{":generation": {"S": taskAgentAssignmentGeneration(assignment)}}
	switch {
	case observedGeneration != taskAgentAssignmentGeneration(assignment):
		guard := &dynamoDBTransactWriteConditionCheckAction{TableName: s.tableName, Key: markerKey, ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)"}
		if observedGeneration != "" {
			guard.ConditionExpression = "generation = :generation"
			guard.ExpressionAttributeValues = map[string]map[string]string{":generation": {"S": observedGeneration}}
		}
		actions = append(actions, dynamoDBTransactWriteItem{ConditionCheck: guard})
	case member.DurableGeneration == "":
		actions = append(actions, dynamoDBTransactWriteItem{Delete: &dynamoDBTransactWriteDeleteAction{TableName: s.tableName, Key: markerKey, ConditionExpression: "generation = :generation", ExpressionAttributeValues: values}})
	default:
		markerKey["generation"] = map[string]string{"S": member.DurableGeneration}
		actions = append(actions, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{TableName: s.tableName, Item: markerKey, ConditionExpression: "generation = :generation", ExpressionAttributeValues: values}})
	}
	condition := "assignment_state = :observed_state AND last_event_seq = :observed_seq"
	stateValues := map[string]map[string]string{":observed_state": {"S": string(projection.Assignment.State)}, ":observed_seq": {"N": strconv.FormatInt(projection.LastEventSeq, 10)}}
	if operation == nil {
		actions = append(actions, dynamoDBTransactWriteItem{ConditionCheck: &dynamoDBTransactWriteConditionCheckAction{TableName: s.tableName, Key: map[string]map[string]string{"pk": {"S": dynamoDBAssignmentProjectionPK(assignment.ID)}, "sk": {"S": dynamoDBAssignmentProjectionSK}}, ConditionExpression: condition, ExpressionAttributeValues: stateValues}})
	} else {
		journal, err := assignmentOperationDynamoDBItem(*operation)
		if err != nil {
			return nil, err
		}
		updated, err := assignmentProjectionDynamoDBItem(*operation)
		if err != nil {
			return nil, err
		}
		actions = append(actions, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{TableName: s.tableName, Item: journal, ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)"}}, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{TableName: s.tableName, Item: updated, ConditionExpression: condition, ExpressionAttributeValues: stateValues}})
	}
	if isTerminal(projection.Assignment.State) || (operation != nil && isTerminal(operation.Assignment.State)) {
		actions = append(actions, dynamoDBTransactWriteItem{Delete: &dynamoDBTransactWriteDeleteAction{TableName: s.tableName, Key: taskAgentDynamoKey(assignment.TaskID, assignment.AgentID, taskAgentExecutionReceiptSK(assignment)), ConditionExpression: "attribute_not_exists(publication_committed)"}})
	} else {
		actions = append(actions, dynamoDBTransactWriteItem{ConditionCheck: &dynamoDBTransactWriteConditionCheckAction{TableName: s.tableName, Key: taskAgentDynamoKey(assignment.TaskID, assignment.AgentID, taskAgentExecutionReceiptSK(assignment)), ConditionExpression: "attribute_not_exists(publication_committed)"}})
	}
	return actions, nil
}
