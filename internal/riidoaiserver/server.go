package riidoaiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const SchemaVersion = "riido-ai-server.v1"

type ServerConfig struct {
	Authorizer        RequestAuthorizer
	AgentCatalogStore AgentCatalogStore
}

type Server struct {
	agentCatalog AgentCatalogStore
	config       ServerConfig
}

func NewServer(config ServerConfig) Server {
	return Server{agentCatalog: config.AgentCatalogStore, config: config}
}

func (s Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agent-catalog", s.handleAgentCatalog)
	mux.HandleFunc("/v1/agent-catalog/", s.handleAgentCatalog)
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

func (s Server) authorizeAgentCatalog(w http.ResponseWriter, r *http.Request, req AuthorizationRequest) (AgentCatalogPrincipal, bool) {
	if s.config.Authorizer == nil {
		writeError(w, http.StatusServiceUnavailable, "agent catalog requires scoped request authorizer")
		return AgentCatalogPrincipal{}, false
	}
	token, ok := bearerToken(r)
	if !ok {
		writeUnauthorized(w)
		return AgentCatalogPrincipal{}, false
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
		return AgentCatalogPrincipal{}, false
	}
	principal := AgentCatalogPrincipalFromAuthorization(result)
	if err := principal.Validate(); err != nil {
		writeError(w, http.StatusForbidden, "forbidden")
		return AgentCatalogPrincipal{}, false
	}
	return principal, true
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
