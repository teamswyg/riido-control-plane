package main

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/aiagentclientapi/requirements"
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
	m.RuntimeConfigKeys = append([]string(nil), requirements.RequiredRuntimeConfigKeys...)
	assertAIClientAPIError(t, verifyStaticLists(m), "public field")
	m.PublicFields = append([]string(nil), requirements.RequiredPublicFields...)
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
