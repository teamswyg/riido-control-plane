package riidoaiserver

import "testing"

func TestClientVisibleTaskThreadTextLocalizesSystemMessages(t *testing.T) {
	t.Parallel()
	recoveryBlocked := "recovery requires provider session id; refusing fresh start"
	cases := map[string]string{
		"agent work was stopped by user request":               clientMessageTaskStopped,
		"agent work was stopped by task participant removal":   clientMessageTaskStopped,
		"agent work was stopped by agent delete":               clientMessageAgentDeleted,
		"context canceled":                                     clientMessageTaskStopped,
		"supervisor: stopped":                                  clientMessageTaskStopped,
		"agent is busy; task assignment was queued":            clientMessageAgentBusyQueued,
		"agent is busy; task comment was queued":               clientMessageAgentBusyQueued,
		"agent is busy; task thread message was queued":        clientMessageAgentBusyQueued,
		"agent progress updated":                               clientMessageTaskRunning,
		"agent work failed":                                    clientMessageTaskFailed,
		"agent assignment timed out after runtime went silent": clientMessageTaskTimeout,
		recoveryBlocked:                                        clientMessageRecoveryBlocked,
	}
	for input, want := range cases {
		if got := clientVisibleTaskThreadText(input); got != want {
			t.Fatalf("clientVisibleTaskThreadText(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestClientVisibleTaskThreadTextStripsPartialRiidoLogFragments(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"생각 중...\n<ri":                                   "생각 중...",
		`<riido_log>{"code":1001,"args"`:                 "",
		`prefix <riido_log>{"code":1001,"args":{}}<end>`: "prefix",
		`prefix <riido_`:                                 "prefix",
	}
	for input, want := range cases {
		if got := clientVisibleTaskThreadText(input); got != want {
			t.Fatalf("clientVisibleTaskThreadText(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestClientVisibleTaskThreadMessageUsesCommentKindFallback(t *testing.T) {
	t.Parallel()
	cases := map[AgentTaskCommentKind]string{
		AgentTaskCommentQueuedByBusyAgent:     clientMessageAgentBusyQueued,
		AgentTaskCommentStoppedByAgentDeleted: clientMessageAgentDeleted,
		AgentTaskCommentStoppedByUserRequest:  clientMessageTaskStopped,
		AgentTaskCommentTaskFailed:            clientMessageTaskFailed,
	}
	for kind, want := range cases {
		got := clientVisibleTaskThreadMessage(AIAgentTaskThreadRecord{CommentKind: kind})
		if got != want {
			t.Fatalf("clientVisibleTaskThreadMessage(%q) = %q, want %q", kind, got, want)
		}
	}
}
