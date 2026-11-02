package riidoaiserver

import "context"

func (s *Store) loadDashboardQueue(ctx context.Context, agentID string, result *AgentDashboardHistory) error {
	// The queue reader owns execution order, including cross-instance assignments.
	if reader, ok := s.operationStore.(AssignmentQueueReader); ok {
		queue, err := reader.LoadAgentQueueAssignments(ctx, agentID)
		if err != nil {
			return err
		}
		positions := make(map[string]int, len(queue))
		for index, assignment := range queue {
			positions[assignment.ID] = index + 1
		}
		known := make(map[string]bool, len(result.Runs))
		for _, run := range result.Runs {
			known[run.Assignment.ID] = true
		}
		for _, a := range queue {
			if a.AgentID == agentID && a.State == AssignmentQueued && !known[a.ID] {
				result.Runs = append(result.Runs, AgentDashboardRun{Assignment: a})
			}
		}
		for index := range result.Runs {
			if result.Runs[index].Assignment.State == AssignmentQueued {
				result.Runs[index].QueuePosition = positions[result.Runs[index].Assignment.ID]
			}
		}
	}

	return nil
}
