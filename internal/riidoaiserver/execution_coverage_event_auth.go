package riidoaiserver

import (
	"context"
	"net/http"
	"strings"
)

func authenticatedCoverageEventContext(w http.ResponseWriter, r *http.Request, req AgentEventRequest, auth AuthorizationResult) (context.Context, bool) {
	if auth.executionCoverageGeneration != "" || strings.HasPrefix(req.DeviceID, executionCoveragePrefix) || r.Header.Get(executionCoverageHeader) != "" {
		return authenticatedPollContext(w, r, PollRequest{DeviceID: req.DeviceID}, auth)
	}
	return r.Context(), true
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
