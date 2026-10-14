package main

import (
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestLocalStateRequiresLoopbackAndAuthentication(t *testing.T) {
	t.Setenv(envLocalStateDir, t.TempDir())
	t.Setenv(envAIAgentClientTable, "")
	auth, err := riidoaiserver.NewStaticTokenAuthorizer([]riidoaiserver.StaticTokenCredential{{PrincipalID: "local", Token: "local-test-token", Scopes: []string{"ai-agent:*"}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, address := range []string{":18080", "0.0.0.0:18080", "example.com:18080"} {
		t.Setenv(envAddr, address)
		if _, err := localStateDirectory(auth); err == nil {
			t.Fatalf("accepted %s", address)
		}
	}
	t.Setenv(envAddr, "127.0.0.1:18080")
	if _, err := localStateDirectory(nil); err == nil {
		t.Fatal("accepted missing authentication")
	}
	if dir, err := localStateDirectory(auth); err != nil || dir == "" {
		t.Fatalf("valid local config: %s %v", dir, err)
	}
}
