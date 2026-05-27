package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAgentCatalogRBACListShowsOwnAndOtherPublicAgents(t *testing.T) {
	store, server := newAgentCatalogHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"agent-catalog:read"},
	}})
	seedAgentCatalogHTTPRecords(t, store, agentCatalogRBACRecords())

	req := httptest.NewRequest(http.MethodGet, "/v1/agent-catalog", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("agent catalog list status=%d body=%s", resp.Code, resp.Body.String())
	}
	var out AgentCatalogListResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("agent catalog list json: %v", err)
	}
	if out.SchemaVersion != SchemaVersion {
		t.Fatalf("agent catalog schema = %q", out.SchemaVersion)
	}
	if got, want := agentCatalogIDs(out.Agents), []string{"own-private", "own-public", "other-public"}; !sameStrings(got, want) {
		t.Fatalf("visible agent catalog = %v, want %v", got, want)
	}

	privateReq := httptest.NewRequest(http.MethodGet, "/v1/agent-catalog/other-private", nil)
	privateReq.Header.Set("Authorization", "Bearer user-token")
	privateResp := httptest.NewRecorder()
	server.ServeHTTP(privateResp, privateReq)
	if privateResp.Code != http.StatusForbidden {
		t.Fatalf("other private read status=%d body=%s", privateResp.Code, privateResp.Body.String())
	}
}

func TestHTTPAgentCatalogRBACAdminSeesAndMutatesPrivateAgents(t *testing.T) {
	store, server := newAgentCatalogHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "admin-1",
		Token:       "admin-token",
		Scopes:      []string{"agent-catalog:*"},
		Roles:       []AgentCatalogRole{AgentCatalogRoleAdmin},
	}})
	seedAgentCatalogHTTPRecords(t, store, agentCatalogRBACRecords())

	listReq := httptest.NewRequest(http.MethodGet, "/v1/agent-catalog", nil)
	listReq.Header.Set("Authorization", "Bearer admin-token")
	listResp := httptest.NewRecorder()
	server.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("admin catalog list status=%d body=%s", listResp.Code, listResp.Body.String())
	}
	var listOut AgentCatalogListResponse
	if err := json.Unmarshal(listResp.Body.Bytes(), &listOut); err != nil {
		t.Fatalf("admin catalog list json: %v", err)
	}
	if got, want := agentCatalogIDs(listOut.Agents), []string{"own-private", "own-public", "other-public", "other-private"}; !sameStrings(got, want) {
		t.Fatalf("admin visible agent catalog = %v, want %v", got, want)
	}

	patchReq := httptest.NewRequest(http.MethodPatch, "/v1/agent-catalog/other-private", strings.NewReader(`{"visibility":"public"}`))
	patchReq.Header.Set("Authorization", "Bearer admin-token")
	patchResp := httptest.NewRecorder()
	server.ServeHTTP(patchResp, patchReq)
	if patchResp.Code != http.StatusOK {
		t.Fatalf("admin patch status=%d body=%s", patchResp.Code, patchResp.Body.String())
	}
	var patchOut AgentCatalogRecordResponse
	if err := json.Unmarshal(patchResp.Body.Bytes(), &patchOut); err != nil {
		t.Fatalf("admin patch json: %v", err)
	}
	if patchOut.Agent.Visibility != AgentCatalogVisibilityPublic || patchOut.Agent.OwnerPrincipalID != "user-2" {
		t.Fatalf("patched agent = %+v", patchOut.Agent)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/v1/agent-catalog/other-private", nil)
	deleteReq.Header.Set("Authorization", "Bearer admin-token")
	deleteResp := httptest.NewRecorder()
	server.ServeHTTP(deleteResp, deleteReq)
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("admin delete status=%d body=%s", deleteResp.Code, deleteResp.Body.String())
	}
	if _, ok, err := store.GetAgentCatalog(context.Background(), "other-private"); err != nil || ok {
		t.Fatalf("deleted record still visible: ok=%v err=%v", ok, err)
	}
}

func TestHTTPAgentCatalogRBACOwnerMutatesOwnedAgents(t *testing.T) {
	store, server := newAgentCatalogHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "owner-token",
		Scopes:      []string{"agent-catalog:*"},
	}})
	seedAgentCatalogHTTPRecords(t, store, agentCatalogRBACRecords())

	patchReq := httptest.NewRequest(http.MethodPatch, "/v1/agent-catalog/own-private", strings.NewReader(`{"visibility":"public"}`))
	patchReq.Header.Set("Authorization", "Bearer owner-token")
	patchResp := httptest.NewRecorder()
	server.ServeHTTP(patchResp, patchReq)
	if patchResp.Code != http.StatusOK {
		t.Fatalf("owner patch status=%d body=%s", patchResp.Code, patchResp.Body.String())
	}
	var patchOut AgentCatalogRecordResponse
	if err := json.Unmarshal(patchResp.Body.Bytes(), &patchOut); err != nil {
		t.Fatalf("owner patch json: %v", err)
	}
	if patchOut.Agent.OwnerPrincipalID != "user-1" || patchOut.Agent.Visibility != AgentCatalogVisibilityPublic {
		t.Fatalf("owner patched agent = %+v", patchOut.Agent)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/v1/agent-catalog/own-public", nil)
	deleteReq.Header.Set("Authorization", "Bearer owner-token")
	deleteResp := httptest.NewRecorder()
	server.ServeHTTP(deleteResp, deleteReq)
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("owner delete status=%d body=%s", deleteResp.Code, deleteResp.Body.String())
	}
	if _, ok, err := store.GetAgentCatalog(context.Background(), "own-public"); err != nil || ok {
		t.Fatalf("owned record still visible: ok=%v err=%v", ok, err)
	}
}

func TestHTTPAgentCatalogRBACPublicReadDoesNotGrantMutation(t *testing.T) {
	store, server := newAgentCatalogHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"agent-catalog:*"},
	}})
	seedAgentCatalogHTTPRecords(t, store, agentCatalogRBACRecords())

	getReq := httptest.NewRequest(http.MethodGet, "/v1/agent-catalog/other-public", nil)
	getReq.Header.Set("Authorization", "Bearer user-token")
	getResp := httptest.NewRecorder()
	server.ServeHTTP(getResp, getReq)
	if getResp.Code != http.StatusOK {
		t.Fatalf("public read status=%d body=%s", getResp.Code, getResp.Body.String())
	}

	patchReq := httptest.NewRequest(http.MethodPatch, "/v1/agent-catalog/other-public", strings.NewReader(`{"visibility":"private"}`))
	patchReq.Header.Set("Authorization", "Bearer user-token")
	patchResp := httptest.NewRecorder()
	server.ServeHTTP(patchResp, patchReq)
	if patchResp.Code != http.StatusForbidden {
		t.Fatalf("public patch status=%d body=%s", patchResp.Code, patchResp.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/v1/agent-catalog/other-public", nil)
	deleteReq.Header.Set("Authorization", "Bearer user-token")
	deleteResp := httptest.NewRecorder()
	server.ServeHTTP(deleteResp, deleteReq)
	if deleteResp.Code != http.StatusForbidden {
		t.Fatalf("public delete status=%d body=%s", deleteResp.Code, deleteResp.Body.String())
	}
}

func TestHTTPAgentCatalogUsesExternalAuthorizerRBAC(t *testing.T) {
	authorizerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req externalAuthorizerRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			t.Fatalf("decode external request: %v", err)
		}
		if req.SchemaVersion != ExternalAuthorizerRequestSchemaVersion || req.Request.Resource != AuthorizationResourceAgentCatalog {
			t.Fatalf("external request = %+v", req)
		}
		switch req.BearerToken {
		case "admin-token":
			_ = json.NewEncoder(w).Encode(externalAuthorizerResponse{
				SchemaVersion: ExternalAuthorizerResponseSchemaVersion,
				Allowed:       true,
				PrincipalID:   "admin-1",
				Roles:         []AgentCatalogRole{AgentCatalogRoleAdmin},
			})
		case "user-token":
			_ = json.NewEncoder(w).Encode(externalAuthorizerResponse{
				SchemaVersion: ExternalAuthorizerResponseSchemaVersion,
				Allowed:       true,
				PrincipalID:   "user-1",
			})
		default:
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer authorizerServer.Close()
	authorizer, err := NewExternalHTTPAuthorizer(ExternalHTTPAuthorizerConfig{Endpoint: authorizerServer.URL})
	if err != nil {
		t.Fatalf("NewExternalHTTPAuthorizer: %v", err)
	}
	store := newAgentCatalogHTTPStore()
	seedAgentCatalogHTTPRecords(t, store, agentCatalogRBACRecords())
	server := NewServer(ServerConfig{AgentCatalogStore: store, Authorizer: authorizer}).Handler()

	userReq := httptest.NewRequest(http.MethodGet, "/v1/agent-catalog", nil)
	userReq.Header.Set("Authorization", "Bearer user-token")
	userResp := httptest.NewRecorder()
	server.ServeHTTP(userResp, userReq)
	if userResp.Code != http.StatusOK {
		t.Fatalf("user catalog status=%d body=%s", userResp.Code, userResp.Body.String())
	}
	var userOut AgentCatalogListResponse
	if err := json.Unmarshal(userResp.Body.Bytes(), &userOut); err != nil {
		t.Fatalf("user catalog json: %v", err)
	}
	if got, want := agentCatalogIDs(userOut.Agents), []string{"own-private", "own-public", "other-public"}; !sameStrings(got, want) {
		t.Fatalf("external user visible catalog = %v, want %v", got, want)
	}

	privateReq := httptest.NewRequest(http.MethodGet, "/v1/agent-catalog/other-private", nil)
	privateReq.Header.Set("Authorization", "Bearer user-token")
	privateResp := httptest.NewRecorder()
	server.ServeHTTP(privateResp, privateReq)
	if privateResp.Code != http.StatusForbidden {
		t.Fatalf("external user private read status=%d body=%s", privateResp.Code, privateResp.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/v1/agent-catalog/other-private", nil)
	deleteReq.Header.Set("Authorization", "Bearer admin-token")
	deleteResp := httptest.NewRecorder()
	server.ServeHTTP(deleteResp, deleteReq)
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("external admin delete status=%d body=%s", deleteResp.Code, deleteResp.Body.String())
	}
}

func TestHTTPAgentCatalogCreateStampsOwnerFromAuthorization(t *testing.T) {
	_, server := newAgentCatalogHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"agent-catalog:create", "agent-catalog:read"},
	}})

	req := httptest.NewRequest(http.MethodPost, "/v1/agent-catalog", strings.NewReader(`{"agent_id":"created-private","visibility":"private"}`))
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", resp.Code, resp.Body.String())
	}
	var out AgentCatalogRecordResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("create json: %v", err)
	}
	if out.Agent.OwnerPrincipalID != "user-1" || out.Agent.Visibility != AgentCatalogVisibilityPrivate {
		t.Fatalf("created agent = %+v", out.Agent)
	}

	badReq := httptest.NewRequest(http.MethodPost, "/v1/agent-catalog", strings.NewReader(`{"agent_id":"client-owned","owner_principal_id":"user-2","visibility":"public"}`))
	badReq.Header.Set("Authorization", "Bearer user-token")
	badResp := httptest.NewRecorder()
	server.ServeHTTP(badResp, badReq)
	if badResp.Code != http.StatusBadRequest {
		t.Fatalf("client owner create status=%d body=%s", badResp.Code, badResp.Body.String())
	}
}

func newAgentCatalogHTTPTestServer(t *testing.T, credentials []StaticTokenCredential) (*agentCatalogHTTPStore, http.Handler) {
	t.Helper()
	store := newAgentCatalogHTTPStore()
	authorizer, err := NewStaticTokenAuthorizer(credentials)
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	return store, NewServer(ServerConfig{AgentCatalogStore: store, Authorizer: authorizer}).Handler()
}

func seedAgentCatalogHTTPRecords(t *testing.T, store *agentCatalogHTTPStore, records []AgentCatalogRecord) {
	t.Helper()
	for _, record := range records {
		if _, err := store.SaveAgentCatalog(context.Background(), record); err != nil {
			t.Fatalf("seed agent catalog %s: %v", record.AgentID, err)
		}
	}
}

type agentCatalogHTTPStore struct {
	records map[string]AgentCatalogRecord
	order   []string
}

func newAgentCatalogHTTPStore() *agentCatalogHTTPStore {
	return &agentCatalogHTTPStore{records: map[string]AgentCatalogRecord{}}
}

func (s *agentCatalogHTTPStore) ListAgentCatalog(context.Context) ([]AgentCatalogRecord, error) {
	records := make([]AgentCatalogRecord, 0, len(s.order))
	for _, agentID := range s.order {
		if record, ok := s.records[agentID]; ok {
			records = append(records, record)
		}
	}
	return records, nil
}

func (s *agentCatalogHTTPStore) GetAgentCatalog(_ context.Context, agentID string) (AgentCatalogRecord, bool, error) {
	record, ok := s.records[agentID]
	return record, ok, nil
}

func (s *agentCatalogHTTPStore) SaveAgentCatalog(_ context.Context, record AgentCatalogRecord) (AgentCatalogRecord, error) {
	record = normalizeAgentCatalogRecord(record)
	if err := record.Validate(); err != nil {
		return AgentCatalogRecord{}, err
	}
	if _, exists := s.records[record.AgentID]; !exists {
		s.order = append(s.order, record.AgentID)
	}
	s.records[record.AgentID] = record
	return record, nil
}

func (s *agentCatalogHTTPStore) DeleteAgentCatalog(_ context.Context, agentID string) (bool, error) {
	if _, ok := s.records[agentID]; !ok {
		return false, nil
	}
	delete(s.records, agentID)
	filtered := s.order[:0]
	for _, candidate := range s.order {
		if candidate != agentID {
			filtered = append(filtered, candidate)
		}
	}
	s.order = filtered
	return true, nil
}
