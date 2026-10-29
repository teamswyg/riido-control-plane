package riidoaiserver

const legacyContextCondition = "attribute_not_exists(pk) OR #schema = :legacy"

func applyLegacyContextPutGuard(put *dynamoDBTransactWritePutAction) {
	put.ConditionExpression = legacyContextCondition
	put.ExpressionAttributeNames = map[string]string{"#schema": "schema"}
	put.ExpressionAttributeValues = map[string]map[string]string{":legacy": {"S": "assignment_context_snapshot/v1"}}
}

func (s *DynamoDBAssignmentOperationStore) legacyContextSnapshotFence(id string) dynamoDBTransactWriteItem {
	return dynamoDBTransactWriteItem{ConditionCheck: &dynamoDBTransactWriteConditionCheckAction{
		TableName:                 s.tableName,
		Key:                       map[string]map[string]string{"pk": {"S": dynamoDBAssignmentContextPK(id)}, "sk": {"S": dynamoDBAssignmentContextSnapshotSK}},
		ConditionExpression:       legacyContextCondition,
		ExpressionAttributeNames:  map[string]string{"#schema": "schema"},
		ExpressionAttributeValues: map[string]map[string]string{":legacy": {"S": "assignment_context_snapshot/v1"}},
	}}
}

func (s *DynamoDBAssignmentOperationStore) contextPublicationFences(id string, items []map[string]map[string]string, immutable bool) []dynamoDBTransactWriteItem {
	if !immutable {
		for _, item := range items {
			if item["sk"]["S"] == dynamoDBAssignmentContextSnapshotSK {
				return nil // The snapshot Put itself carries the same schema fence.
			}
		}
		return []dynamoDBTransactWriteItem{s.legacyContextSnapshotFence(id)}
	}
	var fences []dynamoDBTransactWriteItem
	for i := range assignmentContextMaxCommentPages {
		fences = append(fences, s.contextPayloadAbsent(id, dynamoDBAssignmentContextCommentSK(i)))
		if i >= len(items)-1 {
			fences = append(fences, s.contextPayloadAbsent(id, assignmentContentSK(i)))
		}
	}
	return fences
}

func (s *DynamoDBAssignmentOperationStore) contextPayloadAbsent(id, sk string) dynamoDBTransactWriteItem {
	return dynamoDBTransactWriteItem{ConditionCheck: &dynamoDBTransactWriteConditionCheckAction{
		TableName:           s.tableName,
		Key:                 map[string]map[string]string{"pk": {"S": dynamoDBAssignmentContextPK(id)}, "sk": {"S": sk}},
		ConditionExpression: "attribute_not_exists(pk)",
	}}
}
