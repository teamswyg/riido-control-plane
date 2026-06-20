package riidoaiserver

func recordPollAction(state *storeState, action PollAction, count bool) {
	if count {
		state.pollActionsTotal[action]++
	}
}
