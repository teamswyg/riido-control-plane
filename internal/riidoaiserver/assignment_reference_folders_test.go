package riidoaiserver

import (
	"context"
	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"strings"
	"testing"
)

func TestReferenceFoldersAssignmentSnapshot(t *testing.T) {
	ctx := context.Background()
	store := NewStoreWithConfig(StoreConfig{AgentRegistry: mustTestAgentRegistry(t)})
	defer store.Close()
	folders := &assignmentcontract.ReferenceFolders{DeviceID: "device-1", RuntimeID: "runtime-1", Paths: []string{"/자료 with spaces"}}
	created, err := store.AssignTask(ctx, "task", AssignRequest{AgentID: "agent-a", RuntimeProvider: "claude", Prompt: "User task", AgentInstruction: "User rules", ReferenceFolders: folders})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(created.Prompt, "/자료 with spaces") || !strings.HasPrefix(created.Prompt, "User task") || created.AgentInstruction != "User rules" {
		t.Fatalf("instruction lost: %+v", created)
	}
	folders.Paths[0] = "/mutated input"
	created.ReferenceFolders.Paths[0] = "/mutated response"
	poll, err := store.PollAgent(ctx, "agent-a", PollRequest{DaemonID: "daemon-1", DeviceID: "device-1", RuntimeID: "runtime-1"})
	if err != nil {
		t.Fatal(err)
	}
	if poll.Assignment.ReferenceFolders.Paths[0] != "/자료 with spaces" {
		t.Fatal("snapshot changed through caller alias")
	}
	bad := folders.Clone()
	bad.DeviceID = "another-device"
	if _, err := store.AssignTask(ctx, "bad", AssignRequest{AgentID: "agent-a", RuntimeProvider: "claude", Prompt: "task", ReferenceFolders: bad}); err == nil {
		t.Fatal("accepted wrong binding")
	}
}
