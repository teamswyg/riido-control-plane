package riidoaiserver

import "errors"

func (a *CoalescingAuthorizer) run(key string, call *coalescedAuthorizationCall, token string, req AuthorizationRequest) {
	defer func() {
		if recover() != nil {
			call.result = AuthorizationResult{}
			call.err = errors.New("shared authorization dependency failed")
		}
		a.finish(key, call)
	}()
	// The provider retains its configured upstream timeout. Each caller's
	// cancellation retires only that waiter; the last waiter cancels this hop.
	call.result, call.err = a.next.Authorize(call.ctx, token, req)
}
