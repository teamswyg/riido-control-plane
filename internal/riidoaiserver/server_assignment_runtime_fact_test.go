package riidoaiserver

import (
	"strings"
	"testing"
)

type fakeRuntimeFactStore struct {
	AIAgentClientStore
	found bool
}

func (f fakeRuntimeFactStore) LookupAgentRuntimeFact(string) (AgentRuntimeBinding, RuntimeRecord, bool) {
	return AgentRuntimeBinding{AgentID: "agent-a", RuntimeID: "runtime-a"},
		RuntimeRecord{RuntimeID: "runtime-a", Kind: RuntimeKindCodex}, f.found
}

type fakeLegacyAgentRegistry struct {
	AIAgentClientStore
	found bool
}

func (f fakeLegacyAgentRegistry) LookupAgent(string) (AgentRuntimeBinding, bool) {
	return AgentRuntimeBinding{AgentID: "agent-a", RuntimeID: "legacy-runtime"}, f.found
}

type fakeAIAgentStoreOnly struct{ AIAgentClientStore }

func TestResolveAgentRuntimeFactUsesFactRegistry(t *testing.T) {
	binding, runtime, err := (Server{aiAgent: fakeRuntimeFactStore{found: true}}).
		resolveAgentRuntimeFact("agent-a")
	if err != nil {
		t.Fatalf("resolve fact registry: %v", err)
	}
	if binding.RuntimeID != "runtime-a" || runtime.Kind != RuntimeKindCodex {
		t.Fatalf("fact result binding=%+v runtime=%+v", binding, runtime)
	}
}

func TestResolveAgentRuntimeFactReportsMissingFact(t *testing.T) {
	_, _, err := (Server{aiAgent: fakeRuntimeFactStore{found: false}}).
		resolveAgentRuntimeFact("agent-a")
	if err == nil || !strings.Contains(err.Error(), "binding is not configured") {
		t.Fatalf("missing fact err = %v", err)
	}
}

func TestResolveAgentRuntimeFactFallsBackToLegacyRegistry(t *testing.T) {
	binding, runtime, err := (Server{aiAgent: fakeLegacyAgentRegistry{found: true}}).
		resolveAgentRuntimeFact("agent-a")
	if err != nil {
		t.Fatalf("resolve legacy registry: %v", err)
	}
	if binding.RuntimeID != "legacy-runtime" || runtime.RuntimeID != "" {
		t.Fatalf("legacy result binding=%+v runtime=%+v", binding, runtime)
	}
}

func TestResolveAgentRuntimeFactReportsMissingRegistry(t *testing.T) {
	_, _, err := (Server{aiAgent: fakeAIAgentStoreOnly{}}).resolveAgentRuntimeFact("agent-a")
	if err == nil || !strings.Contains(err.Error(), "registry is not configured") {
		t.Fatalf("missing registry err = %v", err)
	}
}
