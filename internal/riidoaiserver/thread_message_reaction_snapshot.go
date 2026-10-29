package riidoaiserver

// retainedTaskThreadReactions keeps the reaction state of retained threads
// only. States are replaced, never modified, so sharing them is safe.
func retainedTaskThreadReactions(states map[string]AIAgentTaskThreadReactionState, taskThreads map[string][]AIAgentTaskThreadRecord) map[string]AIAgentTaskThreadReactionState {
	if len(states) == 0 {
		return nil
	}
	out := make(map[string]AIAgentTaskThreadReactionState, len(states))
	for _, threads := range taskThreads {
		for _, thread := range threads {
			if state, ok := states[thread.ThreadID]; ok {
				out[thread.ThreadID] = state
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
