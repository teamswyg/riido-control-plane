package riidoaiserver

import "context"

func (c *AIAgentPrivateTaskContextClient) readScopeGraphQL(ctx context.Context, bearer, query, root string, variables, out any) error {
	if c.readScopeEvidence != nil {
		return c.readScopeEvidence.read(ctx, root, variables, out)
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	attempt := int64(0)
	return retryTaskContextRead(ctx, func() (err error) {
		attempt++
		readCtx, finish := observeTaskContextRead(ctx, "read_scope_query")
		defer func() {
			finish(taskContextReadOutcome(readCtx, err), Int64TraceAttribute("riido.task_context.attempt", attempt), StringTraceAttribute("riido.task_context.query", root))
		}()
		return c.readScopeGraphQLOnce(readCtx, bearer, query, root, variables, out)
	})
}
