package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
)

func writeAIAgentClientError(w http.ResponseWriter, err error) {
	var dependency assignmentDependencyError
	var onboardingDependency onboardingAgentCreateDependencyError
	var latest latestTaskDocumentError
	switch {
	case errors.Is(err, errThreadMessageExecutionCancelled):
		writeError(w, http.StatusServiceUnavailable, err.Error())
	case errors.As(err, &latest):
		writeError(w, http.StatusServiceUnavailable, latest.Error())
	case errors.As(err, &onboardingDependency):
		status := http.StatusServiceUnavailable
		if errors.Is(err, context.DeadlineExceeded) {
			status = http.StatusGatewayTimeout
		}
		writeError(w, status, "onboarding agent creation unavailable; retry the same request_id")
	case errors.Is(err, errOnboardingAgentCreateConflict):
		writeError(w, http.StatusConflict, "onboarding creation request conflicts with its existing result")
	case errors.As(err, &dependency):
		status, message := agentEventErrorResponse(err)
		writeError(w, status, message)
	case errors.Is(err, ErrAuthorizationForbidden):
		writeError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, ErrAIAgentNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, ErrAIAgentAssigned):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrAIAgentTaskThreadConflict):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrAIAgentMembershipChanged):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, errOnboardingCompletionConflict), errors.Is(err, errOnboardingCompletionStopConflict):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusBadRequest, err.Error())
	}
}
