package riidoaiserver

import (
	"context"
	"time"
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
	activeOnly         bool
	detailReader       agentDashboardProjectionReader
	missing            map[string]int64
}

type AgentDashboardRun struct {
	Assignment    Assignment
	StartedAt     *time.Time
	EndedAt       *time.Time
	Message       string
	Waiting       bool
	QueuePosition int
}

type agentDashboardSnapshotCmd struct {
	agentID    string
	activeOnly bool
	reply      chan StoreSnapshot
}

type agentDashboardProjectionReader interface {
	dashboardSnapshot(context.Context, string, StoreSnapshot, bool, string) (StoreSnapshot, map[string]int64, error)
	dashboardDetails(context.Context, string, AgentDashboardTask, int64) (AgentDashboardTask, error)
}

func (h AgentDashboardHistory) fillVisibleDetails(ctx context.Context, agentID string, view *AgentDashboardResponse) error {
	if len(h.missing) == 0 {
		return nil
	}
	rows := []*AgentDashboardTask{}
	if view.Current != nil {
		rows = append(rows, view.Current)
	}
	for i := range view.Queue {
		rows = append(rows, &view.Queue[i])
	}
	for i := range view.History {
		rows = append(rows, &view.History[i])
	}
	for _, row := range rows {
		if seq, ok := h.missing[row.AssignmentID]; ok {
			hydrated, err := h.detailReader.dashboardDetails(ctx, agentID, *row, seq)
			if err != nil {
				return err
			}
			*row = hydrated
		}
	}
	return nil
}
