package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestSharedAgentOnlineForOtherWorkspaceMember(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	owner := AuthorizationResult{PrincipalID: "runtime-owner", WorkspaceID: "workspace-shared"}
	member := AuthorizationResult{PrincipalID: "other-member", WorkspaceID: owner.WorkspaceID}
	enrollment := enrollTestDevice(t, ctx, store, owner, owner.WorkspaceID, "shared-machine", "Shared Mac")
	runtimeID := "shared-machine:claude"
	syncTestRuntime(t, ctx, store, owner, "shared-machine", enrollment.DeviceID, runtimeID)
	created, err := store.CreateAIAgent(ctx, owner, CreateAgentConfigurationRequest{
		Name: "Shared Claude", RuntimeID: runtimeID, Visibility: AgentVisibilityPublic,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, viewer := range []AuthorizationResult{owner, member} {
		t.Run(viewer.PrincipalID, func(t *testing.T) {
			bootstrap, err := store.BootstrapAIAgentClient(ctx, viewer, ClientKindWeb)
			if err != nil {
				t.Fatal(err)
			}
			device := requireDevice(t, bootstrap.Devices, enrollment.DeviceID)
			if len(device.Runtimes) != 1 || device.Runtimes[0].Availability != RuntimeAvailabilityOnline {
				t.Errorf("online owner's runtime is unavailable to %s: %+v", viewer.PrincipalID, device.Runtimes)
			}
			agent := agentByID(bootstrap.Agents, created.Agent.AgentID)
			if agent.AgentID == "" || agent.WorkStatus == AgentWorkStatusOffline {
				t.Errorf("online public agent is unavailable to %s: %+v", viewer.PrincipalID, agent)
			}
		})
	}
}

func TestSharedAgentOnlineAfterHeartbeatOnAnotherServer(t *testing.T) {
	for _, test := range []struct {
		name           string
		recentSnapshot bool
		cachedReader   bool
		sameServer     bool
	}{
		{name: "cold reader"},
		{name: "cached reader", cachedReader: true},
		{name: "another device saved recently", recentSnapshot: true},
		{name: "staging single server with 10 second saves", recentSnapshot: true, sameServer: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			testSharedAgentHeartbeatAcrossServers(t, test.recentSnapshot, test.cachedReader, test.sameServer)
		})
	}
}

func testSharedAgentHeartbeatAcrossServers(t *testing.T, recentSnapshot, cachedReader, sameServer bool) {
	t.Helper()
	ctx := context.Background()
	base := NewDevelopmentAIAgentClientStore()
	owner := AuthorizationResult{PrincipalID: "runtime-owner", WorkspaceID: "workspace-shared"}
	member := AuthorizationResult{PrincipalID: "other-member", WorkspaceID: owner.WorkspaceID}
	enrollment := enrollTestDevice(t, ctx, base, owner, owner.WorkspaceID, "shared-machine", "Shared Mac")
	runtimeID := "shared-machine:claude"
	syncTestRuntime(t, ctx, base, owner, "shared-machine", enrollment.DeviceID, runtimeID)
	created, err := base.CreateAIAgent(ctx, owner, CreateAgentConfigurationRequest{
		Name: "Shared Claude", RuntimeID: runtimeID, Visibility: AgentVisibilityPublic,
	})
	if err != nil {
		t.Fatal(err)
	}
	lastSeen := time.Now().UTC().Add(-time.Minute)
	snapshot, err := base.snapshot(lastSeen)
	if err != nil {
		t.Fatal(err)
	}
	for i := range snapshot.Devices {
		snapshot.Devices[i].DaemonLastSeenAt = lastSeen
	}
	for i := range snapshot.Daemons {
		snapshot.Daemons[i].LastSeenAt = lastSeen
	}
	if recentSnapshot {
		// An unrelated device or profile can save the shared snapshot without
		// refreshing this daemon's heartbeat.
		snapshot.SavedAt = time.Now().UTC()
	}
	snapshots := &memoryAIAgentClientSnapshotStore{}
	if err := snapshots.SaveAIAgentClientSnapshot(ctx, snapshot); err != nil {
		t.Fatal(err)
	}
	writer, err := OpenPersistentAIAgentClientStore(ctx, nil, snapshots)
	if err != nil {
		t.Fatal(err)
	}
	var reader *PersistentAIAgentClientStore
	if sameServer {
		if err := writer.ConfigureSnapshotCadence(0, 10*time.Second); err != nil {
			t.Fatal(err)
		}
		reader = writer
	}
	if cachedReader {
		reader, err = OpenPersistentAIAgentClientStore(ctx, nil, snapshots)
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := writer.SyncAIAgentDaemonRuntimeSnapshot(ctx, owner, DeviceRuntimeSnapshotSyncRequest{
		DaemonID: "shared-machine", DeviceID: enrollment.DeviceID,
		Runtimes: []RuntimeSnapshotRecord{{RuntimeID: runtimeID, Kind: RuntimeKindClaudeCode,
			Availability: RuntimeAvailabilityOnline, DetectionState: RuntimeDetectionStateDetected}},
	}); err != nil {
		t.Fatal(err)
	}
	if reader == nil {
		reader, err = OpenPersistentAIAgentClientStore(ctx, nil, snapshots)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		reader.snapshotLastReloadAt = time.Now().Add(-20 * time.Second)
		if sameServer {
			reader.snapshotLastReloadAt = time.Now().Add(-2 * time.Minute)
		}
	}
	bootstrap, err := reader.BootstrapAIAgentClient(ctx, member, ClientKindWeb)
	if err != nil {
		t.Fatal(err)
	}
	device := requireDevice(t, bootstrap.Devices, enrollment.DeviceID)
	if len(device.Runtimes) != 1 || device.Runtimes[0].Availability != RuntimeAvailabilityOnline {
		t.Errorf("fresh heartbeat on writer appears offline on reader: %+v", device.Runtimes)
	}
	agent := agentByID(bootstrap.Agents, created.Agent.AgentID)
	if agent.AgentID == "" || agent.WorkStatus == AgentWorkStatusOffline {
		t.Errorf("online public agent appears offline on reader: %+v", agent)
	}
	if _, _, ok := reader.LookupAgentRuntimeFact(created.Agent.AgentID); !ok {
		t.Error("online public agent cannot be assigned on reader after writer accepted heartbeat")
	}
	assertSharedAgentCallable(t, reader, member, created.Agent.AgentID)
}

func assertSharedAgentCallable(t *testing.T, reader *PersistentAIAgentClientStore, member AuthorizationResult, agentID string) {
	t.Helper()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: member.PrincipalID, Token: "member-token", Scopes: []string{"ai-agent:*", "task:shared-task:read", "task:shared-task:assign"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	assignments := NewStoreWithConfig(StoreConfig{AgentRegistry: reader})
	t.Cleanup(func() { assignments.Close() })
	server := NewServer(ServerConfig{AIAgentClient: reader, Assignment: assignments, Authorizer: authorizer}).Handler()
	path := "/v2/client/workspaces/" + member.WorkspaceID + "/ai-agent/tasks/shared-task/"
	var assignable AgentClientListResponse
	aiAgentSmokeDecode(t, aiAgentSmokeRequest(t, server, http.MethodGet, path+"assignable-agents", "member-token", "", http.StatusOK), &assignable)
	if !agentListContains(assignable.Agents, agentID) {
		t.Fatal("online public agent missing from member's HTTP assignable list")
	}
	body, err := json.Marshal(AssignAIAgentTaskRequest{AgentID: agentID, Instruction: "Check the shared task"})
	if err != nil {
		t.Fatal(err)
	}
	aiAgentSmokeRequest(t, server, http.MethodPost, path+"agent-assignments", "member-token", string(body), http.StatusAccepted)
}
