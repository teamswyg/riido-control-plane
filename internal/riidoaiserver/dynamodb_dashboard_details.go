package riidoaiserver

import (
	"context"
	"encoding/json"
)

// Projection-only reads keep permission checks, counts and pagination independent
// of event payloads. Only hydrate a visible run when its local replay is behind.
func (s *DynamoDBAssignmentOperationStore) dashboardDetails(ctx context.Context, agentID string, task AgentDashboardTask, frontier int64) (AgentDashboardTask, error) {
	credentials, err := s.credentialsProvider.Credentials(ctx)
	if err != nil {
		return task, err
	}
	a := Assignment{ID: task.AssignmentID, TaskID: task.TaskID, AgentID: agentID}
	summary := dashboardRunAccumulator{run: AgentDashboardRun{Assignment: a, Waiting: task.ConfirmationReason != ""}}
	for _, kind := range dashboardOperationPrefixes {
		err = s.dashboardItems(ctx, credentials, agentID, kind+task.AssignmentID+":", "events_json", func(item map[string]map[string]string) error {
			var events []TaskEvent
			if err := json.Unmarshal([]byte(dynamoDBStringValue(item, "events_json")), &events); err != nil {
				return err
			}
			for _, e := range events {
				if e.AssignmentID == a.ID && e.TaskID == a.TaskID && e.AgentID == agentID && e.Seq <= frontier {
					summary.observe(e)
				}
			}
			return nil
		})
		if err != nil {
			return task, err
		}
	}
	run := summary.run
	if run.StartedAt != nil {
		task.StartedAt = run.StartedAt
	}
	if run.EndedAt != nil {
		task.EndedAt = run.EndedAt
	}
	if run.Message != "" {
		task.Message = run.Message
	}
	if task.StartedAt != nil && task.EndedAt != nil && !task.EndedAt.Before(*task.StartedAt) {
		seconds := int64(task.EndedAt.Sub(*task.StartedAt).Seconds())
		task.DurationSeconds = &seconds
	}
	return task, nil
}
