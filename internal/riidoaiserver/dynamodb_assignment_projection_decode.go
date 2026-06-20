package riidoaiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

func assignmentProjectionFromDynamoDBItem(item map[string]map[string]string) (assignmentProjectionRecord, error) {
	if schema := dynamoDBStringValue(item, "schema_version"); schema != AssignmentProjectionSchemaVersion {
		return assignmentProjectionRecord{}, fmt.Errorf("unsupported assignment projection schema_version %q", schema)
	}
	var assignment Assignment
	if err := json.Unmarshal([]byte(dynamoDBStringValue(item, "assignment_json")), &assignment); err != nil {
		return assignmentProjectionRecord{}, fmt.Errorf("decode assignment projection assignment_json: %w", err)
	}
	if assignment.ID == "" {
		return assignmentProjectionRecord{}, errors.New("assignment projection assignment_id is required")
	}
	lastEventSeq, err := strconv.ParseInt(dynamoDBNumberValue(item, "last_event_seq"), 10, 64)
	if err != nil {
		return assignmentProjectionRecord{}, fmt.Errorf("decode assignment projection last_event_seq: %w", err)
	}
	return assignmentProjectionRecord{Assignment: assignment, LastEventSeq: lastEventSeq}, nil
}
