package riidoaiserver

import (
	"sort"
	"time"
)

const agentDashboardPageSize = 10

// Filtering precedes counting and pagination: agent visibility grants no task access.
func buildAgentDashboard(agent AgentClientRecord, history AgentDashboardHistory, tasks map[string]AIAgentTaskContextComponent, page int, now time.Time) AgentDashboardResponse {
	zone := time.FixedZone("Asia/Seoul", 9*60*60)
	end := now.In(zone)
	end = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, zone)
	start := time.Date(end.Year(), end.Month()-5, 1, 0, 0, 0, 0, zone)
	out := AgentDashboardResponse{SchemaVersion: SchemaVersion, AgentID: agent.AgentID, WorkStatus: string(agent.WorkStatus), Queue: []AgentDashboardTask{}, History: []AgentDashboardTask{}, Activity: []AgentDashboardDay{}, Page: page, PageSize: agentDashboardPageSize, TimeZone: zone.String(), Scope: "accessible_tasks", HistoryDurable: history.Durable, ActivityStart: start.Format("2006-01-02"), ActivityEnd: end.Format("2006-01-02")}
	completedByDay := map[string]int{}
	for _, run := range history.Runs {
		task, allowed := tasks[run.Assignment.TaskID]
		if !allowed {
			continue
		}
		assignment := run.Assignment
		item := AgentDashboardTask{CreatedBy: assignment.CreatedBy, AgentID: assignment.AgentID, TaskID: assignment.TaskID, AssignmentID: assignment.ID, RunID: "run-" + assignment.ID, Title: task.Title, KeyNumber: task.KeyNumber, State: string(assignment.State), Message: run.Message, ConfirmationReason: string(assignment.ConfirmationReason), QueuePosition: run.QueuePosition, QueuedAt: assignment.CreatedAt, StartedAt: run.StartedAt, EndedAt: run.EndedAt}
		if !assignment.CreatedAt.IsZero() && (out.ObservedSince == nil || assignment.CreatedAt.Before(*out.ObservedSince)) {
			at := assignment.CreatedAt
			out.ObservedSince = &at
		}
		if run.StartedAt != nil && run.EndedAt != nil && !run.EndedAt.Before(*run.StartedAt) {
			seconds := int64(run.EndedAt.Sub(*run.StartedAt).Seconds())
			item.DurationSeconds = &seconds
		}
		switch {
		case isDashboardTerminal(assignment.State):
			out.History = append(out.History, item)
			if assignment.State == AssignmentCompleted {
				out.CompletedRuns++
				if run.EndedAt != nil {
					day := run.EndedAt.In(zone).Format("2006-01-02")
					completedByDay[day]++
				}
			}
		case assignment.State == AssignmentQueued:
			if run.QueuePosition > 0 {
				out.Queue = append(out.Queue, item)
			}
		default:
			if history.UsesActiveLease && history.ActiveAssignmentID != assignment.ID {
				continue
			}
			if run.Waiting {
				item.State = "waiting_for_user"
			}
			if out.Current == nil || item.QueuedAt.Before(out.Current.QueuedAt) {
				current := item
				out.Current = &current
			}
		}
	}
	if out.Current != nil {
		if agent.WorkStatus != AgentWorkStatusOffline {
			out.WorkStatus = out.Current.State
		}
		if out.Current.State == "waiting_for_user" {
			out.ConfirmationCount = 1
		}
	}
	sort.Slice(out.Queue, func(i, j int) bool { return out.Queue[i].QueuePosition < out.Queue[j].QueuePosition })
	out.WaitingCount = len(out.Queue)
	sort.Slice(out.History, func(i, j int) bool {
		left, right := out.History[i], out.History[j]
		if left.EndedAt != nil && right.EndedAt != nil && !left.EndedAt.Equal(*right.EndedAt) {
			return left.EndedAt.After(*right.EndedAt)
		}
		return left.AssignmentID > right.AssignmentID
	})
	out.Total = len(out.History)
	lastPage := (out.Total + agentDashboardPageSize - 1) / agentDashboardPageSize
	if lastPage < 1 {
		lastPage = 1
	}
	if out.Page > lastPage {
		out.Page = lastPage
	}
	if out.Page < 1 {
		out.Page = 1
	}
	from := (out.Page - 1) * agentDashboardPageSize
	to := from + agentDashboardPageSize
	if to > out.Total {
		to = out.Total
	}
	out.History = out.History[from:to]
	out.HasNextPage = to < out.Total
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		count := completedByDay[key]
		out.Activity = append(out.Activity, AgentDashboardDay{Date: key, Completed: count})
		out.ActivityCompleted += count
		if count > 0 {
			out.ActiveDays++
		}
	}
	return out
}
