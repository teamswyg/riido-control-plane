package riidoaiserver

import "testing"

func TestAssignRequestWithThreadResumeSessionRequiresRuntimeMatch(t *testing.T) {
	base := AssignRequest{
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		ModelID:         "gpt-5",
	}
	thread := AIAgentTaskThreadRecord{
		AgentID:           "agent-1",
		RuntimeProvider:   "codex",
		ModelID:           "gpt-5",
		ProviderSessionID: "th-1",
	}

	resumable := assignRequestWithThreadResumeSession(base, thread)
	if resumable.ResumeSessionID != "th-1" {
		t.Fatalf("resume_session_id = %q, want %q", resumable.ResumeSessionID, "th-1")
	}

	wrongProvider := assignRequestWithThreadResumeSession(AssignRequest{
		AgentID:         "agent-1",
		RuntimeProvider: "claude",
		ModelID:         "gpt-5",
	}, thread)
	if wrongProvider.ResumeSessionID != "" {
		t.Fatalf("wrong provider reused session: %+v", wrongProvider)
	}

	wrongModel := assignRequestWithThreadResumeSession(AssignRequest{
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		ModelID:         "gpt-4.1",
	}, thread)
	if wrongModel.ResumeSessionID != "" {
		t.Fatalf("wrong model reused session: %+v", wrongModel)
	}

	missingContext := assignRequestWithThreadResumeSession(base, AIAgentTaskThreadRecord{
		AgentID:           "agent-1",
		ProviderSessionID: "th-1",
	})
	if missingContext.ResumeSessionID != "" {
		t.Fatalf("missing runtime context reused session: %+v", missingContext)
	}
}
