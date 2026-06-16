package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The AI Agent client snapshot is one DynamoDB item (400 KB hard limit). Progress
// lines and threads accumulate per run; uncapped they eventually make every
// client-projection write fail with a ValidationException. snapshot() must bound
// both, keeping the most recent lines/threads.
func TestSnapshotBoundsThreadLinesAndThreadCount(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()

	manyLines := make([]AgentThreadProgressLine, aiAgentClientThreadProgressLineLimit+500)
	for i := range manyLines {
		manyLines[i] = AgentThreadProgressLine{Seq: i + 1, Message: "line"}
	}
	threads := make([]AIAgentTaskThreadRecord, aiAgentClientThreadsPerTaskLimit+25)
	for i := range threads {
		threads[i] = AIAgentTaskThreadRecord{
			ThreadID: "thread-" + strconv.Itoa(i),
			TaskID:   "task-cap",
			RunID:    "run-" + strconv.Itoa(i),
			Lines:    append([]AgentThreadProgressLine(nil), manyLines...),
		}
	}
	store.mu.Lock()
	store.taskThreads["task-cap"] = threads
	store.mu.Unlock()

	snap, err := store.snapshot(time.Now().UTC())
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	got := snap.TaskThreads["task-cap"]
	if len(got) != aiAgentClientThreadsPerTaskLimit {
		t.Fatalf("threads not capped: got %d want %d", len(got), aiAgentClientThreadsPerTaskLimit)
	}
	// Most recent threads (tail) are kept.
	if want := "thread-" + strconv.Itoa(len(threads)-1); got[len(got)-1].ThreadID != want {
		t.Fatalf("did not keep most recent thread: got %q want %q", got[len(got)-1].ThreadID, want)
	}
	for _, th := range got {
		if len(th.Lines) != aiAgentClientThreadProgressLineLimit {
			t.Fatalf("lines not capped: got %d want %d", len(th.Lines), aiAgentClientThreadProgressLineLimit)
		}
		// Latest lines (tail) are kept.
		if th.Lines[len(th.Lines)-1].Seq != len(manyLines) {
			t.Fatalf("did not keep latest line: last seq %d want %d", th.Lines[len(th.Lines)-1].Seq, len(manyLines))
		}
	}
}

func TestPersistentAIAgentClientStoreRestoresDevelopmentState(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("OpenPersistentAIAgentClientStore: %v", err)
	}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-dev"}
	created, err := store.CreateAIAgent(ctx, principal, CreateAgentConfigurationRequest{
		Name:        "Persistent Agent",
		Visibility:  AgentVisibilityPrivate,
		RuntimeID:   "runtime-cursor-dev",
		Description: stringPtr("development persistence"),
		Instruction: stringPtr("persist this configuration"),
	})
	if err != nil {
		t.Fatalf("CreateAIAgent: %v", err)
	}
	enrollment, err := store.EnrollDeviceCredential(ctx, principal, "workspace-dev", EnrollDeviceRequest{DisplayName: "Development Mac"})
	if err != nil {
		t.Fatalf("EnrollDeviceCredential: %v", err)
	}
	if _, err := store.SubmitAIAgentTaskComment(ctx, principal, "task-persist", SubmitAIAgentTaskCommentRequest{
		AgentID: created.Agent.AgentID,
		Body:    "continue from persisted state",
	}); err != nil {
		t.Fatalf("SubmitAIAgentTaskComment: %v", err)
	}
	rawSnapshot, err := json.Marshal(snapshots.snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	if strings.Contains(string(rawSnapshot), enrollment.DeviceSecret) {
		t.Fatal("snapshot must not contain the raw one-time device secret")
	}

	reopened, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("reopen persistent store: %v", err)
	}
	bootstrap, err := reopened.BootstrapAIAgentClient(ctx, principal, ClientKindWeb)
	if err != nil {
		t.Fatalf("BootstrapAIAgentClient: %v", err)
	}
	if !agentListContains(bootstrap.Agents, created.Agent.AgentID) {
		t.Fatalf("reopened bootstrap missing created agent: %+v", bootstrap.Agents)
	}
	devicePrincipal, err := reopened.AuthorizeDeviceCredential(ctx, enrollment.DeviceID, enrollment.DeviceSecret, AuthorizationRequest{Resource: AuthorizationResourceAgent})
	if err != nil {
		t.Fatalf("AuthorizeDeviceCredential: %v", err)
	}
	if devicePrincipal.PrincipalID != principal.PrincipalID {
		t.Fatalf("device principal = %+v", devicePrincipal)
	}
	events, err := reopened.AIAgentClientEvents(ctx, principal)
	if err != nil {
		t.Fatalf("AIAgentClientEvents: %v", err)
	}
	var foundTypedWorkStatus bool
	for _, event := range events {
		if _, ok := event.Payload.(AgentWorkStatusChangedEvent); ok {
			foundTypedWorkStatus = true
		}
	}
	if !foundTypedWorkStatus {
		t.Fatalf("reopened events should restore typed work-status payloads: %+v", events)
	}
}

func TestPersistentAIAgentClientStoreReloadsDeviceCredentialAcrossProcesses(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-dev"}
	writer, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}
	reader, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("open stale reader: %v", err)
	}
	reader.snapshotReloadInterval = 0

	enrollment, err := writer.EnrollDeviceCredential(ctx, principal, "workspace-dev", EnrollDeviceRequest{DisplayName: "Development Mac"})
	if err != nil {
		t.Fatalf("EnrollDeviceCredential: %v", err)
	}
	startedAt := time.Now().Add(-321 * time.Second).UTC()
	if _, err := writer.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, DeviceRuntimeSnapshotSyncRequest{
		DaemonID:          "daemon-a",
		DeviceID:          enrollment.DeviceID,
		DeviceDisplayName: "Development Mac",
		Profile:           "development",
		PID:               4321,
		UptimeSeconds:     321,
		StartedAt:         startedAt,
		Runtimes: []RuntimeSnapshotRecord{{
			RuntimeID:       "daemon-a:codex",
			Kind:            RuntimeKindCodex,
			ProviderVersion: "codex-cli 0.133.0",
			Models:          []RuntimeModelRecord{{ModelID: "gpt-5.5", Label: "GPT-5.5", IsDefault: true}},
		}},
	}); err != nil {
		t.Fatalf("SyncAIAgentDaemonRuntimeSnapshot: %v", err)
	}

	devicePrincipal, err := reader.AuthorizeDeviceCredential(ctx, enrollment.DeviceID, enrollment.DeviceSecret, AuthorizationRequest{Resource: AuthorizationResourceAgent})
	if err != nil {
		t.Fatalf("stale reader should reload device credential: %v", err)
	}
	if devicePrincipal.PrincipalID != principal.PrincipalID || devicePrincipal.WorkspaceID != principal.WorkspaceID {
		t.Fatalf("device principal = %+v", devicePrincipal)
	}
	devices, err := reader.ListAIAgentDevices(ctx, principal)
	if err != nil {
		t.Fatalf("ListAIAgentDevices: %v", err)
	}
	if got := countRuntimeOccurrences(devices.Devices, "daemon-a:codex"); got != 1 {
		t.Fatalf("runtime occurrences after reload = %d, want 1; devices=%+v", got, devices.Devices)
	}
	detail, err := reader.GetAIAgentDeviceDaemon(ctx, principal, enrollment.DeviceID)
	if err != nil {
		t.Fatalf("GetAIAgentDeviceDaemon: %v", err)
	}
	if detail.Daemon.Profile != "development" || detail.Daemon.PID != 4321 || detail.Daemon.UptimeSeconds != 321 || !detail.Daemon.StartedAt.Equal(startedAt) {
		t.Fatalf("daemon detail facts after reload = %+v", detail.Daemon)
	}
	if len(detail.Runtimes) != 1 ||
		detail.Runtimes[0].RuntimeID != "daemon-a:codex" ||
		detail.Runtimes[0].ProviderVersion != "codex-cli 0.133.0" {
		t.Fatalf("daemon detail runtimes after reload = %+v", detail.Runtimes)
	}
}

func TestPersistentAIAgentClientStorePreservesRuntimeSnapshotWhenDeviceSecretRotates(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-dev"}
	writer, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}
	reader, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("open stale reader: %v", err)
	}
	reader.snapshotReloadInterval = 0

	enrollment, err := writer.EnrollDeviceCredential(ctx, principal, "workspace-dev", EnrollDeviceRequest{DisplayName: "Development Mac"})
	if err != nil {
		t.Fatalf("EnrollDeviceCredential: %v", err)
	}
	if _, err := writer.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, DeviceRuntimeSnapshotSyncRequest{
		DaemonID:          "daemon-a",
		DeviceID:          enrollment.DeviceID,
		DeviceDisplayName: "Development Mac",
		Runtimes: []RuntimeSnapshotRecord{{
			RuntimeID: "daemon-a:codex",
			Kind:      RuntimeKindCodex,
		}},
	}); err != nil {
		t.Fatalf("SyncAIAgentDaemonRuntimeSnapshot: %v", err)
	}

	rotated, err := writer.EnrollDeviceCredential(ctx, principal, "workspace-dev", EnrollDeviceRequest{
		DeviceID:    enrollment.DeviceID,
		DisplayName: "Development Mac",
	})
	if err != nil {
		t.Fatalf("EnrollDeviceCredential rotate: %v", err)
	}
	if rotated.DeviceID != enrollment.DeviceID || rotated.DeviceSecret == enrollment.DeviceSecret {
		t.Fatalf("rotated enrollment = %+v original=%+v", rotated, enrollment)
	}
	if _, err := reader.AuthorizeDeviceCredential(ctx, enrollment.DeviceID, enrollment.DeviceSecret, AuthorizationRequest{Resource: AuthorizationResourceAgent}); !errors.Is(err, ErrAuthorizationUnauthenticated) {
		t.Fatalf("old device secret must be rejected after rotation, err=%v", err)
	}
	if _, err := reader.AuthorizeDeviceCredential(ctx, rotated.DeviceID, rotated.DeviceSecret, AuthorizationRequest{Resource: AuthorizationResourceAgent}); err != nil {
		t.Fatalf("rotated device secret should authorize: %v", err)
	}

	devices, err := reader.ListAIAgentDevices(ctx, principal)
	if err != nil {
		t.Fatalf("ListAIAgentDevices: %v", err)
	}
	if got := countRuntimeOccurrences(devices.Devices, "daemon-a:codex"); got != 1 {
		t.Fatalf("runtime occurrences after credential rotation = %d, want 1; devices=%+v", got, devices.Devices)
	}
	detail, err := reader.GetAIAgentDeviceDaemon(ctx, principal, rotated.DeviceID)
	if err != nil {
		t.Fatalf("GetAIAgentDeviceDaemon: %v", err)
	}
	if detail.Daemon.DaemonID != "daemon-a" {
		t.Fatalf("daemon detail lost after credential rotation = %+v", detail.Daemon)
	}
}

func TestPersistentAIAgentClientStoreReloadsMovedRuntimeAcrossProcesses(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-dev"}
	writer, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}

	firstEnrollment, err := writer.EnrollDeviceCredential(ctx, principal, "workspace-dev", EnrollDeviceRequest{DisplayName: "Old Mac"})
	if err != nil {
		t.Fatalf("EnrollDeviceCredential first: %v", err)
	}
	if _, err := writer.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, DeviceRuntimeSnapshotSyncRequest{
		DaemonID: "daemon-local",
		DeviceID: firstEnrollment.DeviceID,
		Runtimes: []RuntimeSnapshotRecord{{
			RuntimeID: "daemon-local:codex",
			Kind:      RuntimeKindCodex,
		}},
	}); err != nil {
		t.Fatalf("SyncAIAgentDaemonRuntimeSnapshot first: %v", err)
	}

	reader, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("open stale reader: %v", err)
	}
	reader.snapshotReloadInterval = 0
	secondEnrollment, err := writer.EnrollDeviceCredential(ctx, principal, "workspace-dev", EnrollDeviceRequest{DisplayName: "Replacement Mac"})
	if err != nil {
		t.Fatalf("EnrollDeviceCredential replacement: %v", err)
	}
	if _, err := writer.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, DeviceRuntimeSnapshotSyncRequest{
		DaemonID: "daemon-local",
		DeviceID: secondEnrollment.DeviceID,
		Runtimes: []RuntimeSnapshotRecord{{
			RuntimeID: "daemon-local:codex",
			Kind:      RuntimeKindCodex,
		}},
	}); err != nil {
		t.Fatalf("SyncAIAgentDaemonRuntimeSnapshot replacement: %v", err)
	}

	devices, err := reader.ListAIAgentDevices(ctx, principal)
	if err != nil {
		t.Fatalf("ListAIAgentDevices: %v", err)
	}
	if got := countRuntimeOccurrences(devices.Devices, "daemon-local:codex"); got != 1 {
		t.Fatalf("runtime occurrences after move reload = %d, want 1; devices=%+v", got, devices.Devices)
	}
	if deviceID := deviceIDForRuntime(devices.Devices, "daemon-local:codex"); deviceID != secondEnrollment.DeviceID {
		t.Fatalf("runtime owner device = %q, want %q; devices=%+v", deviceID, secondEnrollment.DeviceID, devices.Devices)
	}
}

func TestPersistentAIAgentClientStoreCoalescesHotSnapshotReloads(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-dev"}
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("OpenPersistentAIAgentClientStore: %v", err)
	}

	loadsAfterOpen := snapshots.loads
	for range 5 {
		if _, err := store.BootstrapAIAgentClient(ctx, principal, ClientKindWeb); err != nil {
			t.Fatalf("BootstrapAIAgentClient: %v", err)
		}
		if _, err := store.ListAIAgentDevices(ctx, principal); err != nil {
			t.Fatalf("ListAIAgentDevices: %v", err)
		}
	}
	if snapshots.loads != loadsAfterOpen {
		t.Fatalf("hot read methods should reuse fresh in-process snapshot: loads=%d after open=%d", snapshots.loads, loadsAfterOpen)
	}

	store.snapshotLastReloadAt = time.Now().Add(-2 * defaultAIAgentClientSnapshotReloadInterval)
	if _, err := store.BootstrapAIAgentClient(ctx, principal, ClientKindWeb); err != nil {
		t.Fatalf("BootstrapAIAgentClient after expiry: %v", err)
	}
	if snapshots.loads != loadsAfterOpen+1 {
		t.Fatalf("expired freshness window should reload once: loads=%d after open=%d", snapshots.loads, loadsAfterOpen)
	}
}

func TestPersistentAIAgentClientStoreCoalescesRuntimeHeartbeatSnapshots(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-dev"}
	now := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("OpenPersistentAIAgentClientStore: %v", err)
	}
	store.snapshotClock = func() time.Time { return now }

	enrollment, err := store.EnrollDeviceCredential(ctx, principal, "workspace-dev", EnrollDeviceRequest{DisplayName: "Development Mac"})
	if err != nil {
		t.Fatalf("EnrollDeviceCredential: %v", err)
	}
	req := DeviceRuntimeSnapshotSyncRequest{
		DaemonID:  "daemon-hot",
		DeviceID:  enrollment.DeviceID,
		StartedAt: now.Add(-time.Minute),
		Runtimes: []RuntimeSnapshotRecord{{
			RuntimeID: "daemon-hot:codex",
			Kind:      RuntimeKindCodex,
		}},
	}
	if _, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, req); err != nil {
		t.Fatalf("SyncAIAgentDaemonRuntimeSnapshot first: %v", err)
	}
	savesAfterFirstRuntime := snapshots.saves

	now = now.Add(time.Second)
	req.UptimeSeconds = 61
	if _, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, req); err != nil {
		t.Fatalf("SyncAIAgentDaemonRuntimeSnapshot hot heartbeat: %v", err)
	}
	if snapshots.saves != savesAfterFirstRuntime {
		t.Fatalf("hot heartbeat should reuse in-memory state without snapshot save: saves=%d want %d", snapshots.saves, savesAfterFirstRuntime)
	}

	now = now.Add(defaultAIAgentClientHeartbeatSnapshotSaveInterval + time.Second)
	req.UptimeSeconds = 72
	if _, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, req); err != nil {
		t.Fatalf("SyncAIAgentDaemonRuntimeSnapshot persistence heartbeat: %v", err)
	}
	if snapshots.saves != savesAfterFirstRuntime+1 {
		t.Fatalf("heartbeat persistence window should save once: saves=%d want %d", snapshots.saves, savesAfterFirstRuntime+1)
	}
}

func TestPersistentAIAgentClientStoreSavesMeaningfulRuntimeSnapshotChangeImmediately(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-dev"}
	now := time.Date(2026, 6, 16, 13, 0, 0, 0, time.UTC)
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("OpenPersistentAIAgentClientStore: %v", err)
	}
	store.snapshotClock = func() time.Time { return now }

	enrollment, err := store.EnrollDeviceCredential(ctx, principal, "workspace-dev", EnrollDeviceRequest{DisplayName: "Development Mac"})
	if err != nil {
		t.Fatalf("EnrollDeviceCredential: %v", err)
	}
	req := DeviceRuntimeSnapshotSyncRequest{
		DaemonID:  "daemon-hot",
		DeviceID:  enrollment.DeviceID,
		StartedAt: now.Add(-time.Minute),
		Runtimes: []RuntimeSnapshotRecord{{
			RuntimeID: "daemon-hot:codex",
			Kind:      RuntimeKindCodex,
		}},
	}
	if _, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, req); err != nil {
		t.Fatalf("SyncAIAgentDaemonRuntimeSnapshot first: %v", err)
	}
	savesAfterFirstRuntime := snapshots.saves

	now = now.Add(time.Second)
	req.Runtimes[0].ProviderVersion = "codex-cli 0.133.0"
	if _, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, req); err != nil {
		t.Fatalf("SyncAIAgentDaemonRuntimeSnapshot changed runtime: %v", err)
	}
	if snapshots.saves != savesAfterFirstRuntime+1 {
		t.Fatalf("meaningful runtime change should save immediately: saves=%d want %d", snapshots.saves, savesAfterFirstRuntime+1)
	}
}

func TestPersistentAIAgentClientStoreDeduplicatesThreadProgressSeq(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	principal := AuthorizationResult{PrincipalID: "user-1"}
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("OpenPersistentAIAgentClientStore: %v", err)
	}
	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-progress-dedupe", AssignAIAgentTaskRequest{
		AgentID:      "agent-owned-codex",
		AssignmentID: "asn-progress-dedupe",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	first, err := store.RecordAIAgentThreadProgress(ctx, assigned.AgentID, AgentThreadProgressBatchRequest{
		TaskID:       assigned.TaskID,
		AssignmentID: assigned.AssignmentID,
		ThreadID:     assigned.ThreadID,
		RunID:        assigned.RunID,
		Lines:        []AgentThreadProgressLine{{Seq: 1, Message: "working"}},
	})
	if err != nil {
		t.Fatalf("RecordAIAgentThreadProgress first: %v", err)
	}
	if first.AcceptedLines != 1 {
		t.Fatalf("first accepted lines = %d, want 1", first.AcceptedLines)
	}
	savesAfterFirst := snapshots.saves
	duplicate, err := store.RecordAIAgentThreadProgress(ctx, assigned.AgentID, AgentThreadProgressBatchRequest{
		TaskID:       assigned.TaskID,
		AssignmentID: assigned.AssignmentID,
		ThreadID:     assigned.ThreadID,
		RunID:        assigned.RunID,
		Lines:        []AgentThreadProgressLine{{Seq: 1, Message: "working duplicate"}},
	})
	if err != nil {
		t.Fatalf("RecordAIAgentThreadProgress duplicate: %v", err)
	}
	if duplicate.AcceptedLines != 0 {
		t.Fatalf("duplicate accepted lines = %d, want 0", duplicate.AcceptedLines)
	}
	if snapshots.saves != savesAfterFirst {
		t.Fatalf("duplicate progress should not save snapshot: saves=%d want %d", snapshots.saves, savesAfterFirst)
	}
	threads, err := store.ListAIAgentTaskThreads(ctx, principal, assigned.TaskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if len(threads.Threads) != 1 || len(threads.Threads[0].Lines) != 1 || threads.Threads[0].Lines[0].Message != "working" {
		t.Fatalf("threads after duplicate progress = %+v", threads.Threads)
	}
}

func TestAIAgentClientSnapshotRetainsRecentReplayEvents(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	store.mu.Lock()
	store.events = nil
	for i := 1; i <= aiAgentClientReplayEventLimit+25; i++ {
		store.events = append(store.events, ClientStreamEvent{
			Seq:       int64(i),
			EventType: AgentClientEventWorkStatusChanged,
			Payload: AgentWorkStatusChangedEvent{
				EventType:       AgentClientEventWorkStatusChanged,
				SchemaVersion:   SchemaVersion,
				AgentID:         "agent-owned-codex",
				TaskID:          "task-replay",
				ThreadID:        "thread-replay",
				RunID:           "run-replay",
				WorkStatus:      AgentWorkStatusRunning,
				AssignmentState: AgentAssignmentStateRunning,
				CommentKind:     AgentTaskCommentRuntimeProgress,
			},
		})
	}
	store.mu.Unlock()

	snapshot, err := store.snapshot(time.Date(2026, 6, 3, 1, 2, 3, 0, time.UTC))
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(snapshot.Events) != aiAgentClientReplayEventLimit {
		t.Fatalf("snapshot retained %d events, want %d", len(snapshot.Events), aiAgentClientReplayEventLimit)
	}
	if snapshot.Events[0].Seq != 26 || snapshot.Events[len(snapshot.Events)-1].Seq != int64(aiAgentClientReplayEventLimit+25) {
		t.Fatalf("snapshot retained seq range %d..%d", snapshot.Events[0].Seq, snapshot.Events[len(snapshot.Events)-1].Seq)
	}

	reopened := NewDevelopmentAIAgentClientStore()
	if err := reopened.restoreSnapshot(snapshot); err != nil {
		t.Fatalf("restoreSnapshot: %v", err)
	}
	reopened.mu.Lock()
	event := reopened.appendClientEventLocked(AgentClientEventWorkStatusChanged, AgentWorkStatusChangedEvent{
		EventType:       AgentClientEventWorkStatusChanged,
		SchemaVersion:   SchemaVersion,
		AgentID:         "agent-owned-codex",
		TaskID:          "task-replay",
		ThreadID:        "thread-replay",
		RunID:           "run-replay",
		WorkStatus:      AgentWorkStatusCompleted,
		AssignmentState: AgentAssignmentStateCompleted,
		CommentKind:     AgentTaskCommentTaskCompleted,
	})
	reopened.mu.Unlock()
	if event.Seq != int64(aiAgentClientReplayEventLimit+26) {
		t.Fatalf("next replay event seq = %d, want %d", event.Seq, aiAgentClientReplayEventLimit+26)
	}
}

func countRuntimeOccurrences(devices []DeviceRecord, runtimeID string) int {
	count := 0
	for _, device := range devices {
		for _, runtime := range device.Runtimes {
			if runtime.RuntimeID == runtimeID {
				count++
			}
		}
	}
	return count
}

func deviceIDForRuntime(devices []DeviceRecord, runtimeID string) string {
	for _, device := range devices {
		for _, runtime := range device.Runtimes {
			if runtime.RuntimeID == runtimeID {
				return device.DeviceID
			}
		}
	}
	return ""
}

type memoryAIAgentClientSnapshotStore struct {
	snapshot AIAgentClientSnapshot
	ok       bool
	loads    int
	saves    int
}

func (s *memoryAIAgentClientSnapshotStore) LoadAIAgentClientSnapshot(context.Context) (AIAgentClientSnapshot, bool, error) {
	s.loads++
	return s.snapshot, s.ok, nil
}

func (s *memoryAIAgentClientSnapshotStore) SaveAIAgentClientSnapshot(_ context.Context, snapshot AIAgentClientSnapshot) error {
	s.snapshot = snapshot
	s.ok = true
	s.saves++
	return nil
}

func agentListContains(agents []AgentClientRecord, agentID string) bool {
	for _, agent := range agents {
		if agent.AgentID == agentID {
			return true
		}
	}
	return false
}
