package riidoaiserver

import (
	"encoding/json"
	"net/http"
)

func (s Server) handleAIAgentClientOnboardingFixtures(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client store is not configured")
		return
	}
	if r.URL.Path == "/v1/client/ai-agent/onboarding/fixtures" || r.URL.Path == "/v1/client/ai-agent/onboarding/fixtures/" {
		if r.Method != http.MethodGet {
			writeMethodNotAllowed(w)
			return
		}
		principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead})
		if !ok {
			return
		}
		response, err := s.aiAgent.ListAIAgentOnboardingFixtures(r.Context(), principal)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, response)
		return
	}
	fixtureID, suffix, ok := splitNestedResourcePath(r.URL.Path, "/v1/client/ai-agent/onboarding/fixtures/")
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch {
	case suffix == "agents" && r.Method == http.MethodPost:
		s.handleAIAgentClientCreateFromOnboardingFixture(w, r, fixtureID)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (s Server) handleAIAgentClientCreateFromOnboardingFixture(w http.ResponseWriter, r *http.Request, fixtureID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionCreate})
	if !ok {
		return
	}
	var req CreateOnboardingAgentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx := r.Context()
	if len(req.RequestID) != 0 {
		var requestID string
		if err := json.Unmarshal(req.RequestID, &requestID); err != nil || requestID == "" {
			writeError(w, http.StatusBadRequest, "request_id must be an ASCII token of 1 to 128 bytes")
			return
		}
		ctx = withOnboardingAgentCreateRequestID(ctx, requestID)
	}
	response, err := s.aiAgent.CreateAIAgentFromOnboardingFixture(ctx, principal, fixtureID, req.CreateAgentConfigurationRequest)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}
