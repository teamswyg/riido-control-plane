package riidoaiserver

import (
	"strings"
	"testing"
)

func TestStaticAgentRegistryNormalizesAndLooksUpBindings(t *testing.T) {
	registry, err := NewStaticAgentRegistry([]AgentRuntimeBinding{{
		AgentID:         " agent-a ",
		DaemonID:        " daemon-1 ",
		DeviceID:        " device-1 ",
		RuntimeID:       " runtime-1 ",
		RuntimeProvider: " claude ",
	}})
	if err != nil {
		t.Fatalf("NewStaticAgentRegistry: %v", err)
	}
	binding, ok := registry.LookupAgent(" agent-a ")
	if !ok {
		t.Fatal("expected binding lookup")
	}
	if binding.AgentID != "agent-a" ||
		binding.DaemonID != "daemon-1" ||
		binding.DeviceID != "device-1" ||
		binding.RuntimeID != "runtime-1" ||
		binding.RuntimeProvider != "claude" {
		t.Fatalf("binding was not normalized: %+v", binding)
	}
	if _, ok := (*StaticAgentRegistry)(nil).LookupAgent("agent-a"); ok {
		t.Fatal("nil registry lookup should be false")
	}
}

func TestStaticAgentRegistryRejectsInvalidBindings(t *testing.T) {
	tests := []struct {
		name    string
		binding AgentRuntimeBinding
		want    string
	}{
		{name: "agent", binding: AgentRuntimeBinding{DaemonID: "daemon-1", RuntimeID: "runtime-1", RuntimeProvider: "claude"}, want: "agent_id is required"},
		{name: "daemon", binding: AgentRuntimeBinding{AgentID: "agent-a", RuntimeID: "runtime-1", RuntimeProvider: "claude"}, want: "daemon_id is required"},
		{name: "runtime", binding: AgentRuntimeBinding{AgentID: "agent-a", DaemonID: "daemon-1", RuntimeProvider: "claude"}, want: "runtime_id is required"},
		{name: "provider", binding: AgentRuntimeBinding{AgentID: "agent-a", DaemonID: "daemon-1", RuntimeID: "runtime-1"}, want: "runtime_provider is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewStaticAgentRegistry([]AgentRuntimeBinding{tt.binding})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("NewStaticAgentRegistry err=%v, want containing %q", err, tt.want)
			}
		})
	}
	_, err := NewStaticAgentRegistry([]AgentRuntimeBinding{
		{AgentID: "agent-a", DaemonID: "daemon-1", RuntimeID: "runtime-1", RuntimeProvider: "claude"},
		{AgentID: " agent-a ", DaemonID: "daemon-2", RuntimeID: "runtime-2", RuntimeProvider: "codex"},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate agent_id") {
		t.Fatalf("duplicate binding err=%v", err)
	}
}

func TestValidateAssignmentBinding(t *testing.T) {
	registry := mustTestAgentRegistry(t)
	if err := validateAssignmentBinding(nil, "agent-a", "claude"); err != nil {
		t.Fatalf("nil registry assignment binding: %v", err)
	}
	if err := validateAssignmentBinding(registry, "agent-a", "claude"); err != nil {
		t.Fatalf("valid assignment binding: %v", err)
	}
	for _, tt := range []struct {
		name            string
		agentID         string
		runtimeProvider string
		want            string
	}{
		{name: "unknown agent", agentID: "agent-missing", runtimeProvider: "claude", want: "not registered"},
		{name: "provider mismatch", agentID: "agent-a", runtimeProvider: "codex", want: "runtime_provider claude"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAssignmentBinding(registry, tt.agentID, tt.runtimeProvider)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("validateAssignmentBinding err=%v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestValidateDaemonBinding(t *testing.T) {
	registry := mustTestAgentRegistry(t)
	if err := validateDaemonBinding(nil, "agent-a", PollRequest{}); err != nil {
		t.Fatalf("nil registry daemon binding: %v", err)
	}
	valid := PollRequest{DaemonID: " daemon-1 ", DeviceID: " device-1 ", RuntimeID: " runtime-1 "}
	if err := validateDaemonBinding(registry, "agent-a", valid); err != nil {
		t.Fatalf("valid daemon binding: %v", err)
	}
	optionalDeviceRegistry, err := NewStaticAgentRegistry([]AgentRuntimeBinding{{
		AgentID:         "agent-b",
		DaemonID:        "daemon-2",
		RuntimeID:       "runtime-2",
		RuntimeProvider: "codex",
	}})
	if err != nil {
		t.Fatalf("NewStaticAgentRegistry optional device: %v", err)
	}
	if err := validateDaemonBinding(optionalDeviceRegistry, "agent-b", PollRequest{DaemonID: "daemon-2", DeviceID: "any-device", RuntimeID: "runtime-2"}); err != nil {
		t.Fatalf("empty binding device should not constrain request device: %v", err)
	}
	for _, tt := range []struct {
		name    string
		req     PollRequest
		agentID string
		want    string
	}{
		{name: "unknown agent", agentID: "agent-missing", req: valid, want: "not registered"},
		{name: "missing daemon", agentID: "agent-a", req: PollRequest{RuntimeID: "runtime-1"}, want: "daemon_id is required"},
		{name: "missing runtime", agentID: "agent-a", req: PollRequest{DaemonID: "daemon-1"}, want: "runtime_id is required"},
		{name: "daemon mismatch", agentID: "agent-a", req: PollRequest{DaemonID: "daemon-2", DeviceID: "device-1", RuntimeID: "runtime-1"}, want: "daemon_id daemon-1"},
		{name: "device mismatch", agentID: "agent-a", req: PollRequest{DaemonID: "daemon-1", DeviceID: "device-2", RuntimeID: "runtime-1"}, want: "device_id device-1"},
		{name: "runtime mismatch", agentID: "agent-a", req: PollRequest{DaemonID: "daemon-1", DeviceID: "device-1", RuntimeID: "runtime-2"}, want: "runtime_id runtime-1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDaemonBinding(registry, tt.agentID, tt.req)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("validateDaemonBinding err=%v, want containing %q", err, tt.want)
			}
		})
	}
}

func mustTestAgentRegistry(t *testing.T) *StaticAgentRegistry {
	t.Helper()
	registry, err := NewStaticAgentRegistry([]AgentRuntimeBinding{{
		AgentID:         "agent-a",
		DaemonID:        "daemon-1",
		DeviceID:        "device-1",
		RuntimeID:       "runtime-1",
		RuntimeProvider: "claude",
	}})
	if err != nil {
		t.Fatalf("NewStaticAgentRegistry: %v", err)
	}
	return registry
}
