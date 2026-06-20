package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func verifyDoc(repoRoot string, m manifest) error {
	path := filepath.Join(repoRoot, filepath.FromSlash(m.HumanDoc))
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read human doc: %w", err)
	}
	doc := string(data)
	for _, token := range []string{
		"ai-agent-snapshot-cqrs-gate.riido.json",
		"ai_agent_client_snapshot_load",
		"ai_agent_client_snapshot_save",
		"store_poll_assignment",
		"50%",
	} {
		if !strings.Contains(doc, token) {
			return fmt.Errorf("human doc missing %q", token)
		}
	}
	return nil
}
