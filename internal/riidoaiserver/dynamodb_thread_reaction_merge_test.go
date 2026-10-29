package riidoaiserver

import (
	"context"
	"testing"
)

func TestDynamoDBSaveKeepsCommittedThreadReactions(t *testing.T) {
	threads := map[string][]AIAgentTaskThreadRecord{"task": {{ThreadID: "a", TaskID: "task"}, {ThreadID: "b", TaskID: "task"}}}
	committed := AIAgentClientSnapshot{TaskThreads: threads, TaskThreadReactions: map[string]AIAgentTaskThreadReactionState{
		"a": {CommentEventID: "event-a"},
		"b": {Members: []AIAgentTaskThreadMemberReaction{{Target: "msg", Emoji: "1f44d", PrincipalID: "user-1"}}},
	}}
	items, _, _, err := encodeSplitDynamoDBAIAgentClientSnapshot(committed, nil)
	if err != nil {
		t.Fatal(err)
	}
	stale := AIAgentClientSnapshot{TaskThreads: threads}
	if err := mergeCommittedThreadReactions(context.Background(), items, &stale); err != nil {
		t.Fatal(err)
	}
	if stale.TaskThreadReactions["a"].CommentEventID != "event-a" || len(stale.TaskThreadReactions["b"].Members) != 1 {
		t.Fatalf("stale writer erased committed reactions: %+v", stale.TaskThreadReactions)
	}
	writer := AIAgentClientSnapshot{TaskThreads: threads, TaskThreadReactions: map[string]AIAgentTaskThreadReactionState{
		"a": {CommentEventID: "event-a", CommentReactionApplied: "1f440"},
	}}
	if err := mergeCommittedThreadReactions(withThreadReactionMutation(context.Background(), "a"), items, &writer); err != nil {
		t.Fatal(err)
	}
	if writer.TaskThreadReactions["a"].CommentReactionApplied != "1f440" || len(writer.TaskThreadReactions["b"].Members) != 1 {
		t.Fatalf("reaction writer must change only its thread: %+v", writer.TaskThreadReactions)
	}
}
