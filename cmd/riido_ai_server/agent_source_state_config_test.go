package main

import (
	"errors"
	"testing"
)

func TestAgentSourceStateConfigFromEnv(t *testing.T) {
	key := "source-state-env-key-0123456789abcdef"
	t.Setenv("RIIDO_AI_SERVER_AGENT_SOURCE_STATE_URL", "")
	t.Setenv("RIIDO_AI_SERVER_AGENT_SOURCE_STATE_KEY", "")
	if config, err := withAgentSourceState(runtimeConfig{Addr: ":1"}, nil); err != nil || config.Addr != ":1" || config.AgentSourceState.BaseURL != "" {
		t.Fatal("unset source state config changed startup", err)
	}
	t.Setenv("RIIDO_AI_SERVER_AGENT_SOURCE_STATE_URL", "https://monolith.example.com")
	t.Setenv("RIIDO_AI_SERVER_AGENT_SOURCE_STATE_KEY", key)
	if config, err := withAgentSourceState(runtimeConfig{}, nil); err != nil || config.AgentSourceState.Key != key {
		t.Fatal("valid source state config rejected", err)
	}
	t.Setenv("RIIDO_AI_SERVER_AGENT_SOURCE_STATE_KEY", "short")
	if _, err := withAgentSourceState(runtimeConfig{}, nil); err == nil {
		t.Fatal("short source state key accepted")
	}
	earlier := errors.New("earlier")
	if _, err := withAgentSourceState(runtimeConfig{}, earlier); !errors.Is(err, earlier) {
		t.Fatal("earlier startup error replaced")
	}
}
