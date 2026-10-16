package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeImageBuildBindsTraceVersionToWorkflowSource(t *testing.T) {
	root := containerContractRepoRoot(t)
	for _, fixture := range []struct{ path, required string }{
		{"packaging/containers/riido_ai_server.Dockerfile", "ARG RIIDO_SOURCE_REVISION"},
		{"packaging/containers/riido_ai_server.Dockerfile", "-X main.buildRevision=${RIIDO_SOURCE_REVISION}"},
		{".github/workflows/deploy-ai-agent-testnet.yml", `--build-arg "RIIDO_SOURCE_REVISION=${GITHUB_SHA}"`},
	} {
		body, err := os.ReadFile(filepath.Join(root, fixture.path))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), fixture.required) {
			t.Fatalf("source revision build binding missing from %s", fixture.path)
		}
	}
}
