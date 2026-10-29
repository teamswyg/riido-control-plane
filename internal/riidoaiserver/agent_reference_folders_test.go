package riidoaiserver

import (
	"context"
	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"testing"
	"time"
)

func TestAgentReferenceFoldersConfiguration(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	folders := &assignmentcontract.ReferenceFolders{DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev", Paths: []string{"/자료/project one", "/자료/project two", "/자료/project one"}}
	created, err := store.CreateAIAgent(ctx, principal, CreateAgentConfigurationRequest{Name: "reference agent", Visibility: AgentVisibilityPrivate, RuntimeID: folders.RuntimeID, ReferenceFolders: folders})
	if err != nil {
		t.Fatal(err)
	}
	id := created.Agent.AgentID
	if len(created.Agent.ReferenceFolders.Paths) != 2 {
		t.Fatal("paths not deduplicated")
	}
	created.Agent.ReferenceFolders.Paths[0] = "/mutated response"
	folders.Paths[0] = "/mutated request"
	updated, err := store.UpdateAIAgentConfiguration(ctx, principal, id, UpdateAgentConfigurationRequest{Name: "renamed"})
	if err != nil || updated.Agent.ReferenceFolders.Paths[0] != "/자료/project one" {
		t.Fatalf("omission/alias: %+v %v", updated, err)
	}
	bad := updated.Agent.ReferenceFolders.Clone()
	bad.DeviceID = "another-device"
	if _, err := store.UpdateAIAgentConfiguration(ctx, principal, id, UpdateAgentConfigurationRequest{ReferenceFolders: bad}); err == nil {
		t.Fatal("accepted another device")
	}
	if _, err := store.UpdateAIAgentConfiguration(ctx, AuthorizationResult{PrincipalID: "user-2", WorkspaceID: principal.WorkspaceID}, id, UpdateAgentConfigurationRequest{ReferenceFolders: bad}); err == nil {
		t.Fatal("accepted other user")
	}
	snapshot, err := store.snapshot(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	reopened := NewDevelopmentAIAgentClientStore()
	if err := reopened.restoreSnapshot(snapshot); err != nil {
		t.Fatal(err)
	}
	restored, err := reopened.UpdateAIAgentConfiguration(ctx, principal, id, UpdateAgentConfigurationRequest{Name: "restored"})
	if err != nil || len(restored.Agent.ReferenceFolders.Paths) != 2 {
		t.Fatalf("restore: %+v %v", restored, err)
	}
	empty := restored.Agent.ReferenceFolders.Clone()
	empty.Paths = []string{}
	cleared, err := reopened.UpdateAIAgentConfiguration(ctx, principal, id, UpdateAgentConfigurationRequest{ReferenceFolders: empty})
	if err != nil || cleared.Agent.ReferenceFolders != nil {
		t.Fatalf("clear: %+v %v", cleared, err)
	}
	changed, err := store.UpdateAIAgentConfiguration(ctx, principal, id, UpdateAgentConfigurationRequest{RuntimeID: "runtime-cursor-dev"})
	if err != nil || changed.Agent.ReferenceFolders != nil {
		t.Fatalf("runtime change: %+v %v", changed, err)
	}
}
