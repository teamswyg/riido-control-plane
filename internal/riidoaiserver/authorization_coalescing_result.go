package riidoaiserver

import "context"

func waitForCoalescedAuthorization(ctx context.Context, call *coalescedAuthorizationCall) (AuthorizationResult, error) {
	if err := ctx.Err(); err != nil {
		return AuthorizationResult{}, err
	}
	select {
	case <-ctx.Done():
		return AuthorizationResult{}, ctx.Err()
	case <-call.done:
		if err := ctx.Err(); err != nil {
			return AuthorizationResult{}, err
		}
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
