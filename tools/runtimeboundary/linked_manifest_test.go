package main

import (
	"strings"
	"testing"
)

func TestVerifyLinkedCDManifestRejectsDecodeError(t *testing.T) {
	t.Parallel()
	repo := runtimeBoundaryTestRepo(t)
	writeFile(t, repo, "runtime-cd.json", "{")
	err := verifyLinkedCDManifest(repo, runtimeBoundaryManifest("doc.md"))
	if err == nil || !strings.Contains(err.Error(), "decode") {
		t.Fatalf("verifyLinkedCDManifest decode err = %v", err)
	}
}

func TestVerifyLinkedCDManifestRejectsSchema(t *testing.T) {
	t.Parallel()
	repo := runtimeBoundaryTestRepo(t)
	writeFile(t, repo, "runtime-cd.json", `{
  "schema_version": "wrong",
  "id": "runtime-cd-ownership",
  "current_strategy": {"workflow": ".github/workflows/deploy-ai-agent-testnet.yml"}
}`)
	err := verifyLinkedCDManifest(repo, runtimeBoundaryManifest("doc.md"))
	if err == nil || !strings.Contains(err.Error(), "unexpected schema") {
		t.Fatalf("verifyLinkedCDManifest schema err = %v", err)
	}
}

func TestVerifyLinkedCDManifestRejectsID(t *testing.T) {
	t.Parallel()
	repo := runtimeBoundaryTestRepo(t)
	writeFile(t, repo, "runtime-cd.json", `{
  "schema_version": "riido-control-plane-runtime-cd-ownership.v1",
  "id": "wrong",
  "current_strategy": {"workflow": ".github/workflows/deploy-ai-agent-testnet.yml"}
}`)
	err := verifyLinkedCDManifest(repo, runtimeBoundaryManifest("doc.md"))
	if err == nil || !strings.Contains(err.Error(), "unexpected id") {
		t.Fatalf("verifyLinkedCDManifest id err = %v", err)
	}
}

func TestVerifyLinkedCDManifestRejectsWorkflow(t *testing.T) {
	t.Parallel()
	repo := runtimeBoundaryTestRepo(t)
	writeFile(t, repo, "runtime-cd.json", `{
  "schema_version": "riido-control-plane-runtime-cd-ownership.v1",
  "id": "runtime-cd-ownership",
  "current_strategy": {"workflow": ".github/workflows/other.yml"}
}`)
	err := verifyLinkedCDManifest(repo, runtimeBoundaryManifest("doc.md"))
	if err == nil || !strings.Contains(err.Error(), "workflow changed") {
		t.Fatalf("verifyLinkedCDManifest workflow err = %v", err)
	}
}
