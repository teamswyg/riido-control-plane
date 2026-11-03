package riidoaiserver

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type dashboardDetailProbe struct {
	agentDashboardProjectionReader
	fetched []string
}

func (p *dashboardDetailProbe) dashboardDetails(_ context.Context, _ string, task AgentDashboardTask, _ int64) (AgentDashboardTask, error) {
	p.fetched = append(p.fetched, task.AssignmentID)
	return task, nil
}

func TestDashboardHydratesOnlyRequestedAccessiblePage(t *testing.T) {
	probe := &dashboardDetailProbe{}
	history := AgentDashboardHistory{detailReader: probe, missing: map[string]int64{}}
	for i := range 24 {
		id := fmt.Sprintf("run-%02d", i)
		task := "visible"
		if i%3 == 0 {
			task = "denied"
		}
		ended := time.Unix(int64(i), 0)
		history.Runs = append(history.Runs, AgentDashboardRun{Assignment: Assignment{ID: id, TaskID: task, State: AssignmentCompleted}, EndedAt: &ended})
		history.missing[id] = 1
	}
	view := buildAgentDashboard(AgentClientRecord{}, history, map[string]AIAgentTaskContextComponent{"visible": {ID: "visible"}}, 2, time.Now())
	if err := history.fillVisibleDetails(context.Background(), "agent", &view); err != nil {
		t.Fatal(err)
	}
	if view.Total != 16 || len(probe.fetched) != 6 || view.CompletedRuns != 16 {
		t.Fatalf("total=%d,hydrated=%v", view.Total, probe.fetched)
	}
	for i, row := range view.History {
		if row.AssignmentID != probe.fetched[i] || row.TaskID != "visible" {
			t.Fatal("out-of-page or unauthorized detail")
		}
	}
}
