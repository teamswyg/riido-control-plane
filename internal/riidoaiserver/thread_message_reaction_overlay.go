package riidoaiserver

// withTaskThreadReactions returns a copy of messages with their reactions. The
// input may be a cached slice, so it is never modified.
func withTaskThreadReactions(
	thread AIAgentTaskThreadRecord,
	messages []AIAgentTaskThreadHistoryMessage,
	state AIAgentTaskThreadReactionState,
) []AIAgentTaskThreadHistoryMessage {
	statusEmoji := agentStatusReactionEmoji(thread.AssignmentState)
	statusTarget := ""
	// A run started by a monolith comment keeps its status reaction on that
	// comment, never on a follow-up thread message.
	if statusEmoji != "" && state.CommentEventID == "" {
		statusTarget = taskThreadStatusReactionTarget(thread, messages)
	}
	if len(state.Members) == 0 && statusTarget == "" {
		return messages
	}
	out := make([]AIAgentTaskThreadHistoryMessage, len(messages))
	copy(out, messages)
	for i := range out {
		target, ok := taskThreadReactionTarget(out[i])
		if !ok {
			continue
		}
		agentEmoji := ""
		if target == statusTarget {
			agentEmoji = statusEmoji
		}
		out[i].Reactions = taskThreadTargetReactions(state.Members, target, thread.AgentID, agentEmoji)
	}
	return out
}

// taskThreadTargetReactions groups the reactions of one target by emoji in
// first-reaction order. The agent status reaction joins its emoji group.
func taskThreadTargetReactions(
	members []AIAgentTaskThreadMemberReaction,
	target, agentID, agentEmoji string,
) []AIAgentMessageReaction {
	var out []AIAgentMessageReaction
	add := func(emoji string, actor AIAgentMessageReactionActor) {
		for i := range out {
			if out[i].Emoji == emoji {
				out[i].Actors = append(out[i].Actors, actor)
				out[i].Count++
				return
			}
		}
		out = append(out, AIAgentMessageReaction{Emoji: emoji, Count: 1, Actors: []AIAgentMessageReactionActor{actor}})
	}
	for _, reaction := range members {
		if reaction.Target == target {
			add(reaction.Emoji, AIAgentMessageReactionActor{Kind: AIAgentMessageReactionActorMember, ID: reaction.PrincipalID})
		}
	}
	if agentEmoji != "" && agentID != "" {
		add(agentEmoji, AIAgentMessageReactionActor{Kind: AIAgentMessageReactionActorAgent, ID: agentID})
	}
	return out
}
