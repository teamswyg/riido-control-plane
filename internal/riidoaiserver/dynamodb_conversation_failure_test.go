package riidoaiserver

import (
	"reflect"
	"testing"
)

func TestConversationDynamoDBFailedPublicationKeepsPreviousHistory(t *testing.T) {
	for _, failure := range []string{"chunk", "commit", "stale-version"} {
		t.Run(failure, func(t *testing.T) {
			f := newSnapshotDynamoDBFixture(t, fixedSnapshotTestNow(), nil, nil)
			defer f.close()
			before := conversationSnapshotForTest(1)
			if err := f.store.SaveAIAgentClientSnapshot(t.Context(), before); err != nil {
				t.Fatal(err)
			}
			after := conversationSnapshotForTest(200)
			after.Version, after.ExpectedVersion = 2, 1
			f.mu.Lock()
			f.failConversationChunk = failure == "chunk"
			f.failMembershipCommit = failure == "commit"
			f.mu.Unlock()
			if failure == "stale-version" {
				after.ExpectedVersion = 0
			}
			if err := f.store.SaveAIAgentClientSnapshot(t.Context(), after); err == nil {
				t.Fatal("failed publication was acknowledged")
			}
			got, found, err := f.store.LoadAIAgentClientSnapshot(t.Context())
			if err != nil || !found || !reflect.DeepEqual(got.TaskThreadMessages, before.TaskThreadMessages) {
				t.Fatalf("previous history damaged: found=%t err=%v", found, err)
			}
			f.mu.Lock()
			f.failConversationChunk, f.failMembershipCommit = false, false
			f.mu.Unlock()
			after.ExpectedVersion = 1
			if err := f.store.SaveAIAgentClientSnapshot(t.Context(), after); err != nil {
				t.Fatalf("retry: %v", err)
			}
		})
	}
}
