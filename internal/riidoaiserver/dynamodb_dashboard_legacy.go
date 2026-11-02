package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
)

// A legacy journal can outlive its STATE row. Look up just this run and retain
// its latest operation, so the visibility and pagination totals stay accurate.
func (s *DynamoDBAssignmentOperationStore) dashboardLegacyProjection(ctx context.Context, credentials AWSCredentials, agentID, id string) (assignmentProjectionRecord, error) {
	var latest assignmentProjectionRecord
	for _, kind := range dashboardOperationPrefixes {
		err := s.dashboardItems(ctx, credentials, agentID, kind+id+":", "assignment_json, events_json", func(item map[string]map[string]string) error {
			var a Assignment
			if err := json.Unmarshal([]byte(dynamoDBStringValue(item, "assignment_json")), &a); err != nil {
				return err
			}
			if a.ID != id {
				return nil
			}
			if a.AgentID != agentID {
				return fmt.Errorf("dashboard journal scope mismatch")
			}
			var events []TaskEvent
			if err := json.Unmarshal([]byte(dynamoDBStringValue(item, "events_json")), &events); err != nil {
				return err
			}
			for _, e := range events {
				if e.AssignmentID == id && e.Seq > latest.LastEventSeq {
					latest.Assignment = a
					latest.LastEventSeq = e.Seq
				}
			}
			return nil
		})
		if err != nil {
			return latest, err
		}
	}
	if latest.Assignment.ID == "" {
		return latest, fmt.Errorf("dashboard assignment missing from projection and journal")
	}
	return latest, nil
}

var dashboardOperationPrefixes = []string{"assign:", "poll-start:", "agent-event:", "client-stop:"}
