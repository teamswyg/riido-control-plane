package riidoaiserver

import (
	"encoding/json"
	"strings"
	"testing"
)

// A legacy thread snapshot without profile_color gets a color on backfill. Its
// content-derived ID must follow the content, and a second backfill must keep it.
func TestAIAgentProfileColorBackfillRecomputesSnapshotID(t *testing.T) {
	agents := map[string]AgentClientRecord{
		"agent-legacy": {AgentID: "agent-legacy", ProfileColor: AgentProfileColorGreen},
	}
	legacySnapshot := &AIAgentTaskThreadAgentSnapshot{AgentID: "agent-legacy", TmpColor: "#94A3B8"}
	legacyID := taskThreadAgentSnapshotID(legacySnapshot)
	threads := map[string][]AIAgentTaskThreadRecord{
		"task-1": {{AgentID: "agent-legacy", AgentSnapshot: legacySnapshot, AgentSnapshotID: legacyID}},
	}

	backfillTaskSnapshotProfileColors(threads, nil, agents)
	first := threads["task-1"][0]
	if first.AgentSnapshot.ProfileColor != AgentProfileColorGreen {
		t.Fatalf("backfilled color = %q, want green from the agent record", first.AgentSnapshot.ProfileColor)
	}
	if first.AgentSnapshotID == legacyID {
		t.Fatalf("snapshot id stayed %q after its content changed", legacyID)
	}
	if want := taskThreadAgentSnapshotID(first.AgentSnapshot); first.AgentSnapshotID != want {
		t.Fatalf("snapshot id = %q, want the content hash %q", first.AgentSnapshotID, want)
	}

	backfillTaskSnapshotProfileColors(threads, nil, agents)
	second := threads["task-1"][0]
	if second.AgentSnapshotID != first.AgentSnapshotID {
		t.Fatalf("second backfill changed the snapshot id: %q -> %q", first.AgentSnapshotID, second.AgentSnapshotID)
	}
	if second.AgentSnapshot.ProfileColor != first.AgentSnapshot.ProfileColor {
		t.Fatalf("second backfill changed the color: %q -> %q", first.AgentSnapshot.ProfileColor, second.AgentSnapshot.ProfileColor)
	}
}

func TestAIAgentProfileColorBackfillKeepsIDsOfSnapshotsThatAlreadyHaveColor(t *testing.T) {
	colored := &AIAgentTaskThreadAgentSnapshot{AgentID: "agent-colored", ProfileColor: AgentProfileColorPurple}
	id := taskThreadAgentSnapshotID(colored)
	threads := map[string][]AIAgentTaskThreadRecord{
		"task-1": {{AgentID: "agent-colored", AgentSnapshot: colored, AgentSnapshotID: id}},
	}
	backfillTaskSnapshotProfileColors(threads, nil, nil)
	if got := threads["task-1"][0].AgentSnapshotID; got != id {
		t.Fatalf("snapshot with a color changed its id: %q -> %q", id, got)
	}
}

func TestAssignedAgentProfileAlwaysSerializesProfileColor(t *testing.T) {
	body, err := json.Marshal(AssignedAgentProfile{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"profile_color":`) {
		t.Fatalf("assigned profile must always carry profile_color, got %s", body)
	}
}
