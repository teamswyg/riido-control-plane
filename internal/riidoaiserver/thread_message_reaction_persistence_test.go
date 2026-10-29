package riidoaiserver

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestThreadMessageReactionsSurviveSnapshotAndSplitParts(t *testing.T) {
	ctx := context.Background()
	store, owner, _, taskID, threadID := seedReactionThread(t)
	user := reactionHistoryMessage(t, store, owner, taskID, threadID, AIAgentTaskThreadMessageRoleUser)
	if _, err := store.ReactToAIAgentTaskThreadMessage(ctx, owner, taskID, threadID, user.MessageID, "1f44d", true); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordCommentDeliveryEvent(ctx, owner.WorkspaceID, threadID, "event-1"); err != nil {
		t.Fatal(err)
	}
	snapshot, err := store.snapshot(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	parts, err := encodeDynamoDBAIAgentClientSnapshotParts(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var decoded AIAgentClientSnapshot
	for _, part := range parts {
		if part.name != dynamoDBAIAgentClientSnapshotPartTaskThreadReactions {
			continue
		}
		raw, err := gunzipBase64(part.gzip)
		if err != nil {
			t.Fatal(err)
		}
		if err := unmarshalDynamoDBAIAgentClientSnapshotPart(part.name, raw, &decoded); err != nil {
			t.Fatal(err)
		}
	}
	if got := decoded.TaskThreadReactions[threadID]; got.CommentEventID != "event-1" || len(got.Members) != 1 {
		t.Fatalf("split part = %+v", got)
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip AIAgentClientSnapshot
	if err := json.Unmarshal(encoded, &roundTrip); err != nil {
		t.Fatal(err)
	}
	restored := NewDevelopmentAIAgentClientStore()
	if err := restored.restoreSnapshot(roundTrip); err != nil {
		t.Fatal(err)
	}
	reloaded := reactionHistoryMessage(t, restored, owner, taskID, threadID, AIAgentTaskThreadMessageRoleUser)
	assertReactions(t, reloaded.Reactions, "1f44d:member:user-1", "1f440:agent:agent-public-openclaw")
}
