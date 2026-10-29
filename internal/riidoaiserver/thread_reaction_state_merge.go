package riidoaiserver

// threadReactionMutation is the change one reaction write made to one thread:
// its state right after the fresh reload and the state it wants to save.
type threadReactionMutation struct {
	threadID string
	before   AIAgentTaskThreadReactionState
}

// mergedThreadReactionState applies the change from before to after on top of
// the committed state, so a write that raced another replica keeps the other
// replica's member reactions and comment status fields.
func mergedThreadReactionState(committed, before, after AIAgentTaskThreadReactionState) AIAgentTaskThreadReactionState {
	merged := AIAgentTaskThreadReactionState{CommentEventID: committed.CommentEventID, CommentReactionApplied: committed.CommentReactionApplied}
	if after.CommentEventID != before.CommentEventID {
		merged.CommentEventID = after.CommentEventID
	}
	if after.CommentReactionApplied != before.CommentReactionApplied {
		merged.CommentReactionApplied = after.CommentReactionApplied
	}
	beforeKeys, afterKeys := memberReactionKeys(before.Members), memberReactionKeys(after.Members)
	kept := memberReactionKeys(nil)
	for _, reaction := range committed.Members {
		key := memberReactionKey(reaction)
		_, wasBefore := beforeKeys[key]
		_, isAfter := afterKeys[key]
		if wasBefore && !isAfter {
			continue
		}
		kept[key] = struct{}{}
		merged.Members = append(merged.Members, reaction)
	}
	for _, reaction := range after.Members {
		key := memberReactionKey(reaction)
		_, wasBefore := beforeKeys[key]
		_, isKept := kept[key]
		if !wasBefore && !isKept {
			merged.Members = append(merged.Members, reaction)
		}
	}
	return merged
}

func (state AIAgentTaskThreadReactionState) isEmpty() bool {
	return len(state.Members) == 0 && state.CommentEventID == "" && state.CommentReactionApplied == ""
}

func memberReactionKey(reaction AIAgentTaskThreadMemberReaction) string {
	return reaction.Target + "\x00" + reaction.Emoji + "\x00" + reaction.PrincipalID
}

func memberReactionKeys(members []AIAgentTaskThreadMemberReaction) map[string]struct{} {
	keys := make(map[string]struct{}, len(members))
	for _, reaction := range members {
		keys[memberReactionKey(reaction)] = struct{}{}
	}
	return keys
}
