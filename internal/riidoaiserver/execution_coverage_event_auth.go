package riidoaiserver

import (
	"context"
	"net/http"
)

func authenticatedCoverageEventContext(w http.ResponseWriter, r *http.Request, req AgentEventRequest, auth AuthorizationResult) (context.Context, bool) {
	// All device-authenticated events retain verified provenance, even when a
	// transient provider probe makes the runtime unavailable for new work.
	return authenticatedPollContext(w, r, PollRequest{DeviceID: req.DeviceID}, auth)
}

func (s *Store) prepareCoverageEvent(ctx context.Context, agentID string, req AgentEventRequest) (context.Context, error) {
	// Terminal delivery may finish an already admitted invocation; it does not
	// authorize another Start. Ready/nonterminal delivery must retain authority.
	if isTerminal(req.State) {
		return ctx, nil
	}
	ctx, err := s.prepareExecutionAdmission(ctx, agentID, PollRequest{DeviceID: req.DeviceID, RuntimeID: req.RuntimeID, DaemonID: req.DaemonID})
	if err != nil {
		return ctx, err
	}
	admission := executionAdmissionFromContext(ctx)
	if admission == nil || admission.coverage == nil {
		return ctx, nil
	}
	reader, ok := s.operationStore.(AssignmentProjectionReader)
	if !ok {
		return ctx, errCLIInstallationUnavailable
	}
	projection, found, err := reader.LoadAssignmentProjection(ctx, req.AssignmentID)
	if err != nil || !found {
		return ctx, errCLIInstallationUnavailable
	}
	return ctx, s.authorizeCoveredAssignment(ctx, projection.Assignment)
}
