package riidoaiserver

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func TestConversationOwnedSaveRetainsNativeHistoryThreads(t *testing.T) {
	f := newSnapshotDynamoDBFixture(t, fixedSnapshotTestNow(), nil, nil)
	defer f.close()
	want := chunkedThreadSnapshotForTest(t)
	want.TaskThreadMessages = map[string][]AIAgentTaskThreadHistoryMessage{}
	for i := range want.TaskThreads["task"] {
		thread := &want.TaskThreads["task"][i]
		thread.ConversationID = ""
		want.TaskThreadMessages[thread.ThreadID] = []AIAgentTaskThreadHistoryMessage{{MessageID: fmt.Sprintf("native:%d", i), Body: thread.Message}}
	}
	if err := f.store.SaveAIAgentClientSnapshot(t.Context(), want); err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(t.Context(), taskAgentPreserveMembershipContextKey{}, true)
	if err := f.store.SaveAIAgentClientSnapshot(ctx, want); err != nil {
		t.Fatal(err)
	}
	got, found, err := f.store.LoadAIAgentClientSnapshot(t.Context())
	if err != nil || !found || !reflect.DeepEqual(got.TaskThreads, want.TaskThreads) || !reflect.DeepEqual(got.TaskThreadMessages, want.TaskThreadMessages) {
		t.Fatalf("owned save lost native history threads: found=%t err=%v", found, err)
	}
}
