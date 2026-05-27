package riidoaiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ServerConfig struct {
	Authorizer        RequestAuthorizer
	AgentCatalogStore AgentCatalogStore
	ProviderStatus    ProviderStatusStore
	ProviderRead      ProviderStatusReader
}

type Server struct {
	agentCatalog AgentCatalogStore
	provider     ProviderStatusStore
	providerRead ProviderStatusReader
	config       ServerConfig
}

func NewServer(config ServerConfig) Server {
	providerRead := config.ProviderRead
	if providerRead == nil {
		if reader, ok := config.ProviderStatus.(ProviderStatusReader); ok {
			providerRead = reader
		}
	}
	return Server{
		agentCatalog: config.AgentCatalogStore,
		provider:     config.ProviderStatus,
		providerRead: providerRead,
		config:       config,
	}
}

func (s Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agent-catalog", s.handleAgentCatalog)
	mux.HandleFunc("/v1/agent-catalog/", s.handleAgentCatalog)
	mux.HandleFunc("/v1/agents/", s.handleAgents)
	return mux
}

func (s Server) handleAgentCatalog(w http.ResponseWriter, r *http.Request) {
	if s.agentCatalog == nil {
		writeError(w, http.StatusServiceUnavailable, "agent catalog store is not configured")
		return
	}
	agentID, hasAgentID, ok := splitOptionalResourcePath(r.URL.Path, "/v1/agent-catalog")
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch {
	case !hasAgentID && r.Method == http.MethodGet:
		s.handleAgentCatalogList(w, r)
	case !hasAgentID && r.Method == http.MethodPost:
		s.handleAgentCatalogCreate(w, r)
	case hasAgentID && r.Method == http.MethodGet:
		s.handleAgentCatalogGet(w, r, agentID)
	case hasAgentID && r.Method == http.MethodPatch:
		s.handleAgentCatalogUpdate(w, r, agentID)
	case hasAgentID && r.Method == http.MethodDelete:
		s.handleAgentCatalogDelete(w, r, agentID)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s Server) handleAgentCatalogList(w http.ResponseWriter, r *http.Request) {
	principal, ok := s.authorizeAgentCatalog(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgentCatalog, Action: AuthorizationActionRead})
	if !ok {
		return
	}
	records, err := s.agentCatalog.ListAgentCatalog(r.Context())
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, AgentCatalogListResponse{
		SchemaVersion: SchemaVersion,
		Agents:        VisibleAgentCatalogRecords(principal, records),
	})
}

func (s Server) handleAgentCatalogCreate(w http.ResponseWriter, r *http.Request) {
	principal, ok := s.authorizeAgentCatalog(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgentCatalog, Action: AuthorizationActionCreate})
	if !ok {
		return
	}
	var req CreateAgentCatalogRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	record := AgentCatalogRecord{
		AgentID:          req.AgentID,
		OwnerPrincipalID: principal.PrincipalID,
		Visibility:       req.Visibility,
	}
	saved, err := s.agentCatalog.SaveAgentCatalog(r.Context(), record)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, AgentCatalogRecordResponse{SchemaVersion: SchemaVersion, Agent: saved})
}

func (s Server) handleAgentCatalogGet(w http.ResponseWriter, r *http.Request, agentID string) {
	principal, ok := s.authorizeAgentCatalog(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgentCatalog, Action: AuthorizationActionRead, AgentID: agentID})
	if !ok {
		return
	}
	record, found, err := s.agentCatalog.GetAgentCatalog(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	decision := EvaluateAgentCatalogAccess(principal, record, AgentCatalogActionRead)
	if !decision.Allowed {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	writeJSON(w, http.StatusOK, AgentCatalogRecordResponse{SchemaVersion: SchemaVersion, Agent: record})
}

func (s Server) handleAgentCatalogUpdate(w http.ResponseWriter, r *http.Request, agentID string) {
	principal, ok := s.authorizeAgentCatalog(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgentCatalog, Action: AuthorizationActionUpdate, AgentID: agentID})
	if !ok {
		return
	}
	record, found, err := s.agentCatalog.GetAgentCatalog(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	decision := EvaluateAgentCatalogAccess(principal, record, AgentCatalogActionUpdate)
	if !decision.Allowed {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var req UpdateAgentCatalogRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	record.Visibility = req.Visibility
	saved, err := s.agentCatalog.SaveAgentCatalog(r.Context(), record)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, AgentCatalogRecordResponse{SchemaVersion: SchemaVersion, Agent: saved})
}

func (s Server) handleAgentCatalogDelete(w http.ResponseWriter, r *http.Request, agentID string) {
	principal, ok := s.authorizeAgentCatalog(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgentCatalog, Action: AuthorizationActionDelete, AgentID: agentID})
	if !ok {
		return
	}
	record, found, err := s.agentCatalog.GetAgentCatalog(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	decision := EvaluateAgentCatalogAccess(principal, record, AgentCatalogActionDelete)
	if !decision.Allowed {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	deleted, err := s.agentCatalog.DeleteAgentCatalog(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, struct {
		SchemaVersion string `json:"schema_version"`
		Deleted       bool   `json:"deleted"`
	}{SchemaVersion: SchemaVersion, Deleted: deleted})
}

func (s Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	agentID, suffix, ok := splitResourcePath(r.URL.Path, "/v1/agents/")
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch {
	case suffix == "provider-status" && r.Method == http.MethodPost:
		s.handleProviderStatusSync(w, r, agentID)
	case suffix == "provider-status" && r.Method == http.MethodGet:
		s.handleProviderStatusRead(w, r, agentID)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (s Server) handleProviderStatusSync(w http.ResponseWriter, r *http.Request, agentID string) {
	if s.provider == nil {
		writeError(w, http.StatusServiceUnavailable, "provider status store is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionProviderStatusWrite, AgentID: agentID}); !ok {
		return
	}
	var req ProviderStatusSyncRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.provider.SyncProviderStatus(r.Context(), agentID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleProviderStatusRead(w http.ResponseWriter, r *http.Request, agentID string) {
	if s.providerRead == nil {
		writeError(w, http.StatusServiceUnavailable, "provider status reader is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionProviderStatusRead, AgentID: agentID}); !ok {
		return
	}
	response, found, err := s.providerRead.GetProviderStatus(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) authorizeAgentCatalog(w http.ResponseWriter, r *http.Request, req AuthorizationRequest) (AgentCatalogPrincipal, bool) {
	result, ok := s.authorizeRequest(w, r, req)
	if !ok {
		return AgentCatalogPrincipal{}, false
	}
	principal := AgentCatalogPrincipalFromAuthorization(result)
	if err := principal.Validate(); err != nil {
		writeError(w, http.StatusForbidden, "forbidden")
		return AgentCatalogPrincipal{}, false
	}
	return principal, true
}

func (s Server) authorizeRequest(w http.ResponseWriter, r *http.Request, req AuthorizationRequest) (AuthorizationResult, bool) {
	if s.config.Authorizer == nil {
		writeError(w, http.StatusServiceUnavailable, "scoped request authorizer is not configured")
		return AuthorizationResult{}, false
	}
	token, ok := bearerToken(r)
	if !ok {
		writeUnauthorized(w)
		return AuthorizationResult{}, false
	}
	result, err := s.config.Authorizer.Authorize(r.Context(), token, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrAuthorizationForbidden):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, ErrAuthorizationUnauthenticated):
			writeUnauthorized(w)
		default:
			writeError(w, http.StatusServiceUnavailable, err.Error())
		}
		return AuthorizationResult{}, false
	}
	return result, true
}

func bearerToken(r *http.Request) (string, bool) {
	got := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(got, "Bearer ") {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(got, "Bearer "))
	if token == "" {
		return "", false
	}
	return token, true
}

func splitOptionalResourcePath(path, prefix string) (string, bool, bool) {
	if path == prefix || path == prefix+"/" {
		return "", false, true
	}
	withSlash := prefix + "/"
	if !strings.HasPrefix(path, withSlash) {
		return "", false, false
	}
	rest := strings.Trim(strings.TrimPrefix(path, withSlash), "/")
	if rest == "" || strings.Contains(rest, "/") {
		return "", false, false
	}
	return rest, true, true
}

func splitResourcePath(path, prefix string) (string, string, bool) {
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}
	rest := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	parts := strings.Split(rest, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func decodeJSON(r *http.Request, out any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		return errors.New("decode json: trailing data")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, struct {
		SchemaVersion string `json:"schema_version"`
		Error         string `json:"error"`
	}{SchemaVersion: SchemaVersion, Error: message})
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="riido_ai_server"`)
	writeError(w, http.StatusUnauthorized, "unauthorized")
}
