package riidoaiserver

// AssignmentReadScopeEnabled is used by startup validation; it does not
// activate or authorize scoped collection for an individual caller.
func (c *AIAgentPrivateTaskContextClient) AssignmentReadScopeEnabled() bool {
	return c != nil && c.assignmentReadScope
}
