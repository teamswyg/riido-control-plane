package riidoaiserver

import assignmentcontract "github.com/teamswyg/riido-contracts/assignment"

type dashboardRunAccumulator struct {
	run        AgentDashboardRun
	latestSeq  int64
	waitingSeq int64
}

func (c *dashboardRunAccumulator) observe(e TaskEvent) {
	run := &c.run
	running := e.Type == EventAssignmentRunning || (e.Type == EventAssignmentStateUpdated && e.State == AssignmentRunning)
	if running && (run.StartedAt == nil || e.At.Before(*run.StartedAt)) {
		at := e.At
		run.StartedAt = &at
	}
	if isDashboardTerminal(e.State) && (run.EndedAt == nil || e.At.Before(*run.EndedAt)) {
		at := e.At
		run.EndedAt = &at
	}
	if run.Waiting && e.Type == assignmentcontract.EventAssignmentWaitingForUser && e.Seq > c.waitingSeq {
		c.waitingSeq = e.Seq
		run.Message = e.Message
	}
	if e.Seq > c.latestSeq {
		c.latestSeq = e.Seq
		if c.waitingSeq == 0 {
			run.Message = e.Message
		}
	}
}
