package riidoaiserver

import (
	"log"
	"time"
)

const terminalProgressCompensationWindow = 30 * time.Minute

func (s *DevelopmentAIAgentClientStore) compensateRecentTerminalProgressLocked() {
	now := time.Now().UTC()
	published := retainedTerminalProgressIdentities(s.events)
	for _, threads := range s.taskThreads {
		for _, thread := range threads {
			identity := terminalProgressIdentityFromThread(thread)
			if !terminalThreadNeedsProgressCompensation(thread, now) {
				continue
			}
			if _, ok := published[identity]; ok {
				continue
			}
			event, ok := terminalProgressEventFromThread(thread)
			if !ok {
				continue
			}
			s.appendClientEventLocked(event.EventType, event)
			published[identity] = struct{}{}
			log.Printf(
				"riido_ai_agent_sse event=terminal_replay_compensated assignment_id=%q thread_id=%q run_id=%q state=%q",
				thread.AssignmentID, thread.ThreadID, thread.RunID, thread.AssignmentState,
			)
		}
	}
}

func terminalThreadNeedsProgressCompensation(thread AIAgentTaskThreadRecord, now time.Time) bool {
	if !agentAssignmentStateIsTerminal(thread.AssignmentState) || thread.CompletedAt.IsZero() {
		return false
	}
	age := now.Sub(thread.CompletedAt)
	return age >= 0 && age <= terminalProgressCompensationWindow
}

func retainedTerminalProgressIdentities(events []ClientStreamEvent) map[terminalProgressIdentity]struct{} {
	identities := make(map[terminalProgressIdentity]struct{})
	for _, event := range events {
		identity, ok := terminalProgressIdentityFromEvent(event)
		if ok {
			identities[identity] = struct{}{}
		}
	}
	return identities
}
