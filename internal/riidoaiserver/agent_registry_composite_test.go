package riidoaiserver

import "testing"

func TestCompositeAgentRegistryFallsThroughToLaterRegistry(t *testing.T) {
	first := mustRegistryForAgent(t, "agent-a", "daemon-a")
	second := mustRegistryForAgent(t, "agent-b", "daemon-b")

	registry := NewCompositeAgentRegistry(nil, first, second)
	binding, ok := registry.LookupAgent("agent-b")
	if !ok {
		t.Fatal("expected lookup to fall through to second registry")
	}
	if binding.DaemonID != "daemon-b" {
		t.Fatalf("binding daemon = %q, want daemon-b", binding.DaemonID)
	}
}

func TestCompositeAgentRegistryReturnsFirstMatch(t *testing.T) {
	first := mustRegistryForAgent(t, "agent-a", "daemon-first")
	second := mustRegistryForAgent(t, "agent-a", "daemon-second")

	registry := NewCompositeAgentRegistry(first, second)
	binding, ok := registry.LookupAgent("agent-a")
	if !ok {
		t.Fatal("expected binding lookup")
	}
	if binding.DaemonID != "daemon-first" {
		t.Fatalf("binding daemon = %q, want daemon-first", binding.DaemonID)
	}
}

func TestCompositeAgentRegistryMissesAllRegistries(t *testing.T) {
	registry := NewCompositeAgentRegistry(
		mustRegistryForAgent(t, "agent-a", "daemon-a"),
		mustRegistryForAgent(t, "agent-b", "daemon-b"),
	)
	if _, ok := registry.LookupAgent("agent-c"); ok {
		t.Fatal("unexpected lookup hit")
	}
}

func TestNewCompositeAgentRegistryFiltersNilRegistries(t *testing.T) {
	if got := NewCompositeAgentRegistry(nil, nil); got != nil {
		t.Fatalf("all nil registries returned %T, want nil", got)
	}
	registry := mustRegistryForAgent(t, "agent-a", "daemon-a")
	if got := NewCompositeAgentRegistry(nil, registry); got != registry {
		t.Fatalf("single non-nil registry returned %T, want original", got)
	}
}

func mustRegistryForAgent(t *testing.T, agentID, daemonID string) *StaticAgentRegistry {
	t.Helper()
	registry, err := NewStaticAgentRegistry([]AgentRuntimeBinding{{
		AgentID:         agentID,
		DaemonID:        daemonID,
		RuntimeID:       "runtime-" + agentID,
		RuntimeProvider: "codex",
	}})
	if err != nil {
		t.Fatalf("NewStaticAgentRegistry: %v", err)
	}
	return registry
}
