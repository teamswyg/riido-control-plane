package riidoaiserver

import "testing"

func TestEventAgentIDRecognizesClientAgentEvents(t *testing.T) {
	cases := []struct {
		name    string
		payload any
		want    string
		ok      bool
	}{
		{"editability", AgentEditabilityChangedEvent{AgentID: "agent-a"}, "agent-a", true},
		{"work_status", AgentWorkStatusChangedEvent{AgentID: "agent-b"}, "agent-b", true},
		{"progress", AgentThreadProgressEvent{AgentID: "agent-c"}, "agent-c", true},
		{"unknown", DeviceRuntimeSnapshotEvent{}, "", false},
	}
	for _, tt := range cases {
		got, ok := eventAgentID(tt.payload)
		if got != tt.want || ok != tt.ok {
			t.Fatalf("%s: got (%q,%v), want (%q,%v)", tt.name, got, ok, tt.want, tt.ok)
		}
	}
}
