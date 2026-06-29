package riidoaiserver

import (
	"net/http"
)

func (s Server) handleAIAgentClientCreate(w http.ResponseWriter, r *http.Request) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionCreate})
	if !ok {
		return
	}
	var req CreateAgentConfigurationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.aiAgent.CreateAIAgent(r.Context(), principal, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (s Server) handleAIAgentClientEditability(w http.ResponseWriter, r *http.Request, agentID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, AgentID: agentID})
	if !ok {
		return
	}
	response, err := s.aiAgent.GetAIAgentEditability(r.Context(), principal, agentID)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientUpdate(w http.ResponseWriter, r *http.Request, agentID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionUpdate, AgentID: agentID})
	if !ok {
		return
	}
	var req UpdateAgentConfigurationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.aiAgent.UpdateAIAgentConfiguration(r.Context(), principal, agentID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientDelete(w http.ResponseWriter, r *http.Request, agentID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionDelete, AgentID: agentID})
	if !ok {
		return
	}
	response, err := s.aiAgent.DeleteAIAgent(r.Context(), principal, agentID)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}
