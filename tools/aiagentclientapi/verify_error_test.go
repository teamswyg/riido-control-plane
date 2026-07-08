package main

import (
	"strings"
	"testing"
)

func TestVerifyHeaderRejectsIdentityAndRequiredFields(t *testing.T) {
	t.Parallel()
	assertAIClientAPIError(t, verifyHeader(manifest{}), "unexpected manifest identity")
	m := validHeaderManifest()
	m.Title = ""
	assertAIClientAPIError(t, verifyHeader(m), "title, generated_doc")
}

func TestVerifyStaticListsRejectMissingRequirements(t *testing.T) {
	t.Parallel()
	m := validHeaderManifest()
	assertAIClientAPIError(t, verifyStaticLists(m), "runtime config")
	m.RuntimeConfigKeys = []string{"RIIDO_AI_SERVER_AI_AGENT_CLIENT_DEVELOPMENT"}
	assertAIClientAPIError(t, verifyStaticLists(m), "public field")
	m.PublicFields = []string{"workspace_id", "profile_thumbnail_url", "provider_version", "assigned-agent-profiles", "agent-assignments", "thread-stream-subscription", "conversation_id", "parent_thread_id", "agent_snapshot_id", "agent_snapshots", "messages", "author_principal_id"}
	assertAIClientAPIError(t, verifyStaticLists(m), "deployment evidence")
}

func TestVerifyLoopRejectsMissingStep(t *testing.T) {
	t.Parallel()
	assertAIClientAPIError(t, verifyLoop(loop{}), "missing evidence loop step")
}

func validHeaderManifest() manifest {
	return manifest{
		SchemaVersion:    "riido-ai-agent-client-api.v1",
		ID:               "ai-agent-client-api",
		RiidoTask:        "RIID-4721",
		Title:            "AI Agent Client API",
		GeneratedDoc:     "doc.md",
		Workflow:         "workflow.yml",
		EvidenceArtifact: "evidence",
	}
}

func assertAIClientAPIError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want contains %q", err, want)
	}
}
