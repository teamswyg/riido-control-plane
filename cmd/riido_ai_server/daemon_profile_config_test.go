package main

import (
	"strings"
	"testing"
)

func TestConfigFromEnvRejectsUnknownDaemonProfile(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envAIAgentDaemonProfile, "qa-shadow")
	_, err := configFromEnv()
	if err == nil || !strings.Contains(err.Error(), envAIAgentDaemonProfile) {
		t.Fatalf("configFromEnv err=%v", err)
	}
}
