package riidoaiserver

func restoreExecutionHistory(state *storeState, snapshot StoreSnapshot) error {
	for _, claim := range snapshot.ExecutionClaims {
		if err := rememberExecutionClaim(state, &claim); err != nil {
			return err
		}
	}
	for _, receipt := range snapshot.ExecutionTerminations {
		if err := rememberExecutionTermination(state, receipt); err != nil {
			return err
		}
	}
	return nil
}
