package riidoaiserver

import "context"

func waitForCoalescedAuthorization(ctx context.Context, call *coalescedAuthorizationCall) (AuthorizationResult, error) {
	select {
	case <-ctx.Done():
		return AuthorizationResult{}, ctx.Err()
	case <-call.done:
		return cloneAuthorizationResult(call.result), call.err
	}
}

func cloneAuthorizationResult(result AuthorizationResult) AuthorizationResult {
	if len(result.Roles) == 0 {
		return result
	}
	result.Roles = append([]AgentCatalogRole(nil), result.Roles...)
	return result
}
