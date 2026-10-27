package riidoaiserver

import (
	"net/http"
	"testing"
	"time"
)

func TestAgentDirectoryLastSeenPreservesOfflineHeartbeat(t *testing.T) {
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	heartbeat := time.Now().UTC().Add(-24 * time.Hour).Truncate(time.Second)
	setDirectoryHeartbeat(t, f.agents, "agent-public-openclaw", heartbeat, true)
	for _, view := range []string{AssignmentContextViewAgents, AssignmentContextViewAgent} {
		code, body := f.read(view, "agent-public-openclaw", f.leased.LeaseToken)
		if code != http.StatusOK {
			t.Fatalf("directory status=%d", code)
		}
		entry := directoryWireEntry(t, body, "agent-public-openclaw")
		runtime := entry["runtime"].(map[string]any)
		if runtime["last_seen_at"] != heartbeat.Format(time.RFC3339) || runtime["connection"] != "offline" {
			t.Fatalf("offline connection must preserve actual heartbeat: %+v", runtime)
		}
	}
}

func TestAgentDirectoryLastSeenOmitsUnobservedDaemon(t *testing.T) {
	for _, explicit := range []bool{true, false} {
		f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
		setDirectoryHeartbeat(t, f.agents, "agent-public-openclaw", time.Time{}, explicit)
		code, body := f.read(AssignmentContextViewAgent, "agent-public-openclaw", f.leased.LeaseToken)
		if code != http.StatusOK {
			t.Fatalf("directory status=%d", code)
		}
		entry := directoryWireEntry(t, body, "agent-public-openclaw")
		if runtime := entry["runtime"].(map[string]any); runtime["last_seen_at"] != nil {
			t.Fatalf("profile, detection or device update must not invent a heartbeat: %+v", runtime)
		}
	}
}
