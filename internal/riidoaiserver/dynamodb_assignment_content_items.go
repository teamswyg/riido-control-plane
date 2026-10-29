package riidoaiserver

import (
	"encoding/json"
	"fmt"
)

const assignmentContentPrefix = "CONTENT#"

func assignmentContentSK(index int) string {
	return fmt.Sprintf("%s%04d", assignmentContentPrefix, index)
}

func assignmentContentItems(snapshot AssignmentContextSnapshot, pages []AssignmentContentPage) ([]map[string]map[string]string, error) {
	blocks, refs, err := packAssignmentContentPages(pages)
	if err != nil {
		return nil, err
	}
	snapshot.ContentBlocks = refs
	item, err := assignmentContentItem(snapshot.AssignmentID, dynamoDBAssignmentContextSnapshotSK, "assignment_context_snapshot/v2", snapshot)
	if err != nil {
		return nil, err
	}
	item["context_task"], item["context_agent"] = map[string]string{"S": snapshot.TaskID}, map[string]string{"S": snapshot.AgentID}
	item["context_stored"] = map[string]string{"S": snapshot.CapturedAt.Format(dynamoDBAssignmentContextTimeFormat)}
	items := []map[string]map[string]string{item}
	for i, block := range blocks {
		item, err = assignmentContentItem(snapshot.AssignmentID, assignmentContentSK(i), "assignment_context_content_block/v1", block)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, validateContextPayloadBudget(items)
}

func assignmentContentItem(id, sk, schema string, value any) (map[string]map[string]string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	compressed, err := gzipBase64(encoded)
	if err != nil {
		return nil, err
	}
	return map[string]map[string]string{"pk": {"S": dynamoDBAssignmentContextPK(id)}, "sk": {"S": sk}, "schema": {"S": schema}, "payload_gzip": {"S": compressed}}, nil
}

func (s *DynamoDBAssignmentOperationStore) contextPayloadPuts(items []map[string]map[string]string, immutable bool) []dynamoDBTransactWriteItem {
	actions := make([]dynamoDBTransactWriteItem, 0, len(items))
	for _, item := range items {
		put := &dynamoDBTransactWritePutAction{TableName: s.tableName, Item: item}
		if immutable {
			put.ConditionExpression = "attribute_not_exists(pk)"
		} else if item["sk"]["S"] == dynamoDBAssignmentContextSnapshotSK {
			applyLegacyContextPutGuard(put)
		}
		actions = append(actions, dynamoDBTransactWriteItem{Put: put})
	}
	return actions
}
