package riidoaiserver

import (
	"testing"
)

func (f *contextRetentionFixture) conflicts(actions []dynamoDBTransactWriteItem) bool {
	f.fake.mu.Lock()
	defer f.fake.mu.Unlock()
	for _, action := range actions {
		var key, values map[string]map[string]string
		var names map[string]string
		condition := ""
		if action.ConditionCheck != nil {
			key, values, condition = action.ConditionCheck.Key, action.ConditionCheck.ExpressionAttributeValues, action.ConditionCheck.ConditionExpression
			names = action.ConditionCheck.ExpressionAttributeNames
		}
		if action.Put != nil {
			key, values, condition = action.Put.Item, action.Put.ExpressionAttributeValues, action.Put.ConditionExpression
			names = action.Put.ExpressionAttributeNames
		}
		stored := f.fake.items[key["pk"]["S"]+"|"+key["sk"]["S"]]
		switch condition {
		case "attribute_not_exists(pk) OR #schema = :legacy":
			if names["#schema"] != "schema" || values[":legacy"]["S"] != "assignment_context_snapshot/v1" || (len(stored) > 0 && stored["schema"]["S"] != values[":legacy"]["S"]) {
				return true
			}
		case "attribute_not_exists(pk)":
			if len(stored) > 0 {
				return true
			}
		case "revision = :revision":
			if stored["revision"]["S"] != values[":revision"]["S"] {
				return true
			}
		case "assignment_state = :state AND last_event_seq = :seq":
			if stored["assignment_state"]["S"] != values[":state"]["S"] || stored["last_event_seq"]["N"] != values[":seq"]["N"] {
				return true
			}
		}
	}
	return false
}

func (f *contextRetentionFixture) projection(t *testing.T, id string, state AssignmentState) AssignmentOperationRecord {
	t.Helper()
	a := Assignment{ID: id, TaskID: "task", AgentID: "agent", State: state, UpdatedAt: f.now, CreatedAt: f.now}
	record := AssignmentOperationRecord{AssignmentID: id, TaskID: a.TaskID, AgentID: a.AgentID, Assignment: a, RecordedAt: f.now}
	item, err := assignmentProjectionDynamoDBItem(record)
	if err != nil {
		t.Fatal(err)
	}
	f.fake.mu.Lock()
	f.fake.put(item)
	f.fake.mu.Unlock()
	return record
}
