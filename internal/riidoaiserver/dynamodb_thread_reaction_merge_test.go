package riidoaiserver

import (
	"context"
	"strings"
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
	if err := mergeCommittedThreadReactions(withThreadReactionMutation(context.Background(), threadReactionMutation{threadID: "a", before: AIAgentTaskThreadReactionState{CommentEventID: "event-a"}}), items, &writer); err != nil {
		t.Fatal(err)
	}
	if writer.TaskThreadReactions["a"].CommentReactionApplied != "1f440" || len(writer.TaskThreadReactions["b"].Members) != 1 {
		t.Fatalf("reaction writer must change only its thread: %+v", writer.TaskThreadReactions)
	}
}

// Two replicas reloaded the same thread before either saved. The later save
// applies only its own change on top of the earlier commit.
func TestThreadReactionMergeKeepsConcurrentReplicaChanges(t *testing.T) {
	member := func(principal string) AIAgentTaskThreadMemberReaction {
		return AIAgentTaskThreadMemberReaction{Target: "msg", Emoji: "1f44d", PrincipalID: principal}
	}
	before := AIAgentTaskThreadReactionState{CommentEventID: "event-a", Members: []AIAgentTaskThreadMemberReaction{member("user-3"), member("user-4")}}
	committed := AIAgentTaskThreadReactionState{CommentEventID: "event-a", CommentReactionApplied: "2705", Members: []AIAgentTaskThreadMemberReaction{member("user-4"), member("user-2")}}
	after := AIAgentTaskThreadReactionState{CommentEventID: "event-a", Members: []AIAgentTaskThreadMemberReaction{member("user-3"), member("user-1")}}

	merged := mergedThreadReactionState(committed, before, after)

	got := make([]string, 0, len(merged.Members))
	for _, reaction := range merged.Members {
		got = append(got, reaction.PrincipalID)
	}
	// user-2 was added and user-3 removed by the other replica; this write
	// removed user-4 and added user-1.
	if strings.Join(got, ",") != "user-2,user-1" || merged.CommentReactionApplied != "2705" || merged.CommentEventID != "event-a" {
		t.Fatalf("merged = %+v (members %v)", merged, got)
	}
	if empty := mergedThreadReactionState(AIAgentTaskThreadReactionState{}, before, AIAgentTaskThreadReactionState{}); !empty.isEmpty() {
		t.Fatalf("removing every reaction must leave an empty state: %+v", empty)
	}
}
