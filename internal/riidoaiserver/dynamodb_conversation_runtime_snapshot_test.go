package riidoaiserver

import (
	"reflect"
	"testing"
)

func TestConversationChunkedThreadsAllowDaemonRuntimeSnapshot(t *testing.T) {
	f := newSnapshotDynamoDBFixture(t, fixedSnapshotTestNow(), nil, nil)
	defer f.close()
	want := chunkedThreadSnapshotForTest(t)
	if err := f.store.SaveAIAgentClientSnapshot(t.Context(), want); err != nil {
		t.Fatal(err)
	}
	store, err := OpenPersistentAIAgentClientStore(t.Context(), nil, f.store)
	if err != nil {
		t.Fatal(err)
	}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-dev"}
	_, err = store.SyncAIAgentDaemonRuntimeSnapshot(t.Context(), principal, DeviceRuntimeSnapshotSyncRequest{
		DeviceID: "device-a", DaemonID: "daemon-chunked", Runtimes: []RuntimeSnapshotRecord{{RuntimeID: "daemon-chunked:codex", Kind: RuntimeKindCodex}},
	})
	if err != nil {
		t.Fatalf("daemon startup snapshot after conversation chunking: %v", err)
	}
	got, found, err := f.store.LoadAIAgentClientSnapshot(t.Context())
	if err != nil || !found {
		t.Fatalf("reopen after runtime snapshot: found=%t err=%v", found, err)
	}
	if !reflect.DeepEqual(got.TaskThreads, want.TaskThreads) || !reflect.DeepEqual(got.TaskThreadMessages, want.TaskThreadMessages) {
		t.Fatal("runtime snapshot changed committed conversation history")
	}
	if len(got.Daemons) != 1 || got.Daemons[0].DaemonID != "daemon-chunked" {
		t.Fatal("daemon runtime snapshot was not persisted")
	}
}
