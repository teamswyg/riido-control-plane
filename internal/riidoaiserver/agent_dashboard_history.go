package riidoaiserver

import (
	"context"
	"sort"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

// Dashboard history belongs to the assignment journal, never the bounded client replay.
type AgentDashboardHistoryReader interface {
	LoadAgentDashboardHistory(context.Context, string) (AgentDashboardHistory, error)
}

type AgentDashboardHistory struct {
	ActiveAssignmentID string
	UsesActiveLease    bool
	Runs               []AgentDashboardRun
	Durable            bool
}

type AgentDashboardRun struct {
	Assignment    Assignment
	StartedAt     *time.Time
	EndedAt       *time.Time
	Message       string
	Waiting       bool
	QueuePosition int
}

type agentDashboardSnapshotCmd struct{ reply chan StoreSnapshot }

func (s *Store) LoadAgentDashboardHistory(ctx context.Context, agentID string) (AgentDashboardHistory, error) {
	var snapshot StoreSnapshot
	if loader, ok := s.operationStore.(AssignmentOperationLoader); ok {
		operations, err := loader.LoadAssignmentOperations(ctx)
		if err != nil {
			return AgentDashboardHistory{}, err
		}
		state, err := stateFromAssignmentOperations(operations)
		if err != nil {
			return AgentDashboardHistory{}, err
		}
		snapshot = snapshotFromState(&state, s.now())
	} else {
		reply := make(chan StoreSnapshot, 1)
		if err := s.send(ctx, agentDashboardSnapshotCmd{reply: reply}); err != nil {
			return AgentDashboardHistory{}, err
		}
		select {
		case snapshot = <-reply:
		case <-ctx.Done():
			return AgentDashboardHistory{}, ctx.Err()
		}
	}
	result := dashboardHistoryFromSnapshot(snapshot, agentID)
	result.Durable = s.operationStore != nil || s.snapshotStore != nil
	// The queue reader owns execution order, including cross-instance assignments.
	if reader, ok := s.operationStore.(AssignmentQueueReader); ok {
		queue, err := reader.LoadAgentQueueAssignments(ctx, agentID)
		if err != nil {
			return AgentDashboardHistory{}, err
		}
		positions := make(map[string]int, len(queue))
		for index, assignment := range queue {
			positions[assignment.ID] = index + 1
		}
		for index := range result.Runs {
			if result.Runs[index].Assignment.State == AssignmentQueued {
				result.Runs[index].QueuePosition = positions[result.Runs[index].Assignment.ID]
			}
		}
	}
	if reader, ok := s.operationStore.(AssignmentActiveLeaseStore); ok {
		lease, found, err := reader.LoadAgentActiveAssignment(ctx, agentID)
		if err != nil {
			return AgentDashboardHistory{}, err
		}
		result.UsesActiveLease = true
		if found {
			result.ActiveAssignmentID = lease.ActiveAssignmentID
		}
	}
	return result, nil
}

func dashboardHistoryFromSnapshot(snapshot StoreSnapshot, agentID string) AgentDashboardHistory {
	result := AgentDashboardHistory{Runs: []AgentDashboardRun{}}
	for _, assignment := range snapshot.Assignments {
		if assignment.AgentID != agentID {
			continue
		}
		run := AgentDashboardRun{Assignment: assignment, Waiting: assignment.ConfirmationReason != ""}
		var latest int64
		for _, event := range snapshot.Events[assignment.TaskID] {
			if event.AssignmentID != assignment.ID {
				continue
			}
			if (event.Type == EventAssignmentRunning || (event.Type == EventAssignmentStateUpdated && event.State == AssignmentRunning)) && (run.StartedAt == nil || event.At.Before(*run.StartedAt)) {
				at := event.At
				run.StartedAt = &at
			}
			if isDashboardTerminal(event.State) && (run.EndedAt == nil || event.At.Before(*run.EndedAt)) {
				at := event.At
				run.EndedAt = &at
			}
			if event.Seq > latest {
				latest = event.Seq
				run.Message = event.Message
			}
		}
		if run.Waiting {
			var waitingSeq int64
			for _, event := range snapshot.Events[assignment.TaskID] {
				if event.AssignmentID == assignment.ID && event.Type == assignmentcontract.EventAssignmentWaitingForUser && event.Seq > waitingSeq {
					waitingSeq = event.Seq
					run.Message = event.Message
				}
			}
		}

		result.Runs = append(result.Runs, run)
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
