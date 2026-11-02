package riidoaiserver

import (
	"sort"
)

func dashboardHistoryFromSnapshot(snapshot StoreSnapshot, agentID string) AgentDashboardHistory {
	result := AgentDashboardHistory{Runs: []AgentDashboardRun{}}
	runs := make(map[string]*dashboardRunAccumulator, len(snapshot.Assignments))
	for _, a := range snapshot.Assignments {
		if a.AgentID == agentID {
			runs[a.ID] = &dashboardRunAccumulator{run: AgentDashboardRun{Assignment: a, Waiting: a.ConfirmationReason != ""}}
		}
	}
	for taskID, events := range snapshot.Events {
		for _, event := range events {
			if run := runs[event.AssignmentID]; run != nil && run.run.Assignment.TaskID == taskID {
				run.observe(event)
			}
		}
	}
	for _, run := range runs {
		result.Runs = append(result.Runs, run.run)
	}
	sort.Slice(result.Runs, func(i, j int) bool {
		return assignmentQueueSort(result.Runs[i].Assignment) < assignmentQueueSort(result.Runs[j].Assignment)
	})
	position := 0
	for index := range result.Runs {
		if result.Runs[index].Assignment.State == AssignmentQueued {
			position++
			result.Runs[index].QueuePosition = position
		}
	}
	return result
}

func isDashboardTerminal(state AssignmentState) bool {
	return state == AssignmentCompleted || state == AssignmentFailed || state == AssignmentCancelled
}
