package riidoaiserver

import (
	"context"
	"testing"
)

func TestAIAgentDaemonRuntimeSnapshotKeepsAppVersion(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-app-version", WorkspaceID: "workspace-app-version"}
	req := DeviceRuntimeSnapshotSyncRequest{
		DaemonID:          "daemon-app-version",
		DeviceID:          "device-app-version",
		DeviceDisplayName: "App Version Mac",
		AppVersion:        " riido-daemon v0.0.39 ",
		Runtimes: []RuntimeSnapshotRecord{{
			RuntimeID: "device-app-version:codex",
			Kind:      RuntimeKindCodex,
		}},
	}
	first, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, req)
	if err != nil {
		t.Fatalf("SyncAIAgentDaemonRuntimeSnapshot first: %v", err)
	}
	if first.Daemon.AppVersion != "riido-daemon v0.0.39" {
		t.Fatalf("first app_version=%q", first.Daemon.AppVersion)
	}

	req.AppVersion = ""
	second, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, req)
	if err != nil {
		t.Fatalf("SyncAIAgentDaemonRuntimeSnapshot heartbeat: %v", err)
	}
	if second.Daemon.AppVersion != "riido-daemon v0.0.39" {
		t.Fatalf("preserved app_version=%q", second.Daemon.AppVersion)
	}

	detail, err := store.GetAIAgentDeviceDaemon(ctx, principal, req.DeviceID)
	if err != nil {
		t.Fatalf("GetAIAgentDeviceDaemon: %v", err)
	}
	if detail.Daemon.AppVersion != "riido-daemon v0.0.39" {
		t.Fatalf("detail app_version=%q", detail.Daemon.AppVersion)
	}
}
