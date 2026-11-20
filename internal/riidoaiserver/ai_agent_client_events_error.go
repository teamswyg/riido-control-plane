package riidoaiserver

// Event snapshot refresh failures are storage failures, not invalid requests.
// Preserve the cause for internal checks without exposing storage details.
type aiAgentClientEventStoreError struct{ cause error }

func (e aiAgentClientEventStoreError) Error() string { return "ai agent client events unavailable" }
func (e aiAgentClientEventStoreError) Unwrap() error { return e.cause }
