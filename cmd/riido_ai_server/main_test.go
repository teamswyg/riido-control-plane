package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestConfigFromEnvParsesAddressShutdownAndStaticAuthorizer(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envAddr, ":9090")
	t.Setenv(envShutdownTimeoutSeconds, "7")
	t.Setenv(envAuthzTokensJSON, `[{
		"principal_id":"daemon:agent-a",
		"token":"static-token",
		"scopes":["agent:agent-a:poll"]
	}]`)

	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	if config.Addr != ":9090" || config.ShutdownTimeout != 7*time.Second {
		t.Fatalf("config = %+v", config)
	}
	if _, err := config.Authorizer.Authorize(context.Background(), "static-token", riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceAgent,
		Action:   riidoaiserver.AuthorizationActionPoll,
		AgentID:  "agent-a",
	}); err != nil {
		t.Fatalf("static authorize: %v", err)
	}
}

func TestConfigFromEnvDefaultsToPublicHealthOnlyRuntime(t *testing.T) {
	clearRiidoAIServerEnv(t)

	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	if config.Addr != ":8080" || config.ShutdownTimeout != 10*time.Second {
		t.Fatalf("defaults = %+v", config)
	}
	if config.MetricsLogInterval != 0 {
		t.Fatalf("metrics interval should default disabled: %s", config.MetricsLogInterval)
	}
	if config.Authorizer != nil {
		t.Fatalf("optional config should be nil: %+v", config)
	}
}

func TestConfigFromEnvParsesMetricsLogInterval(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envMetricsLogInterval, "15")

	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	if config.MetricsLogInterval != 15*time.Second {
		t.Fatalf("metrics interval = %s", config.MetricsLogInterval)
	}
}

func TestConfigFromEnvParsesPprofAddr(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envPprofAddr, "127.0.0.1:6060")

	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	if config.PprofAddr != "127.0.0.1:6060" {
		t.Fatalf("pprof addr = %q", config.PprofAddr)
	}
}

func TestConfigFromEnvParsesTracing(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envTracingEnabled, "true")
	t.Setenv(envTracingSampleRatio, "0.05")
	t.Setenv(envTracingOTLPEndpoint, "http://127.0.0.1:4318")
	t.Setenv(envTracingServiceName, "riido-ai-server-development")

	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	if !config.Tracing.Enabled || config.Tracing.SampleRatio != 0.05 || config.Tracing.OTLPEndpoint != "http://127.0.0.1:4318" || config.Tracing.ServiceName != "riido-ai-server-development" {
		t.Fatalf("tracing config = %+v", config.Tracing)
	}
}

func TestTracingConfigRejectsInvalidSampleRatio(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envTracingEnabled, "true")
	t.Setenv(envTracingSampleRatio, "2")
	if _, err := configFromEnv(); err == nil || !strings.Contains(err.Error(), envTracingSampleRatio) {
		t.Fatalf("expected tracing sample ratio error, got %v", err)
	}
}

func TestPprofHandlerServesIndex(t *testing.T) {
	server := newPprofServer("127.0.0.1:0")
	if server == nil {
		t.Fatal("pprof server should be configured")
	}
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	resp := httptest.NewRecorder()
	server.Handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("pprof status=%d body=%s", resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), "profile") {
		t.Fatalf("pprof body = %s", resp.Body.String())
	}
	if newPprofServer("") != nil {
		t.Fatal("empty pprof addr should disable pprof server")
	}
}

func TestConfigFromEnvParsesWebAllowedOrigins(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envWebAllowedOrigins, " https://app.riido.io, http://localhost:5173/ , https://app.riido.io ")

	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	want := []string{"https://app.riido.io", "http://localhost:5173"}
	if !reflect.DeepEqual(config.WebAllowedOrigins, want) {
		t.Fatalf("web origins = %v, want %v", config.WebAllowedOrigins, want)
	}
}

func TestConfigFromEnvParsesAIAgentClientDevelopmentStore(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envAIAgentClientDev, "true")
	t.Setenv(envAIAgentClientTable, "riido-ai-agent-development")
	t.Setenv(envAWSRegion, "ap-northeast-2")
	t.Setenv(envAWSContainerCredentialsFullURI, "http://169.254.170.2/credentials")
	t.Setenv(envAssignmentActiveLease, "300")

	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	defer closeRuntimeConfig(config)
	if !config.AIAgentClientDev {
		t.Fatal("AI Agent client development flag should be enabled")
	}
	if config.AssignmentActiveLease != 5*time.Minute {
		t.Fatalf("assignment active lease = %s, want 5m", config.AssignmentActiveLease)
	}
	if config.AIAgentClientStore == nil {
		t.Fatal("AI Agent client snapshot store should be configured")
	}
	if config.AIAgentClientMetrics == nil {
		t.Fatal("AI Agent client persistence metrics should be configured")
	}
	if config.AssignmentOperationStore == nil {
		t.Fatal("assignment operation store should be configured with the development DynamoDB table")
	}
	if _, ok := config.AssignmentOperationStore.(riidoaiserver.AssignmentOperationLoader); !ok {
		t.Fatalf("assignment operation store should load operation journal, got %T", config.AssignmentOperationStore)
	}
	if _, ok := config.AssignmentOperationStore.(riidoaiserver.AssignmentClaimer); !ok {
		t.Fatalf("assignment operation store should claim queued assignments, got %T", config.AssignmentOperationStore)
	}
	if _, ok := config.AssignmentOperationStore.(riidoaiserver.AssignmentActiveLeaseStore); !ok {
		t.Fatalf("assignment operation store should persist active leases, got %T", config.AssignmentOperationStore)
	}
}

func TestConfigFromEnvParsesAgentProfileThumbnailUpload(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envAWSRegion, "ap-northeast-2")
	t.Setenv(envAWSContainerCredentialsFullURI, "http://169.254.170.2/credentials")
	t.Setenv(envAgentProfileThumbnailBucket, "profile-upload-test")
	t.Setenv(envAgentProfileThumbnailPrefix, "thumbnail/ai/profile/")
	t.Setenv(envAgentProfileThumbnailCDNBase, "https://cdn.example.test/")
	t.Setenv(envAgentProfileThumbnailMaxBytes, "1048576")
	t.Setenv(envAgentProfileThumbnailExpires, "60")

	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	if config.AIAgentProfileThumbnails == nil {
		t.Fatal("profile thumbnail upload service should be configured")
	}
}

func TestConfigFromEnvParsesTaskContextReader(t *testing.T) {
	clearRiidoAIServerEnv(t)
	var gotPath string
	var gotAPIKey string
	taskContextServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAPIKey = r.Header.Get(riidoaiserver.AIAgentTaskContextHeaderWorkspaceAPIKey)
		_ = json.NewEncoder(w).Encode(riidoaiserver.AIAgentTaskContext{
			Component: riidoaiserver.AIAgentTaskContextComponent{
				ID:         "component-a",
				Title:      "Task context from existing API server",
				BranchName: "RIID-4800-server-task-context-http-client-assignment-prompt-wiring",
			},
			Document: riidoaiserver.AIAgentTaskContextDocument{
				Content:       "Existing API server document markdown.",
				ContentFormat: "markdown",
			},
			Hierarchy:    riidoaiserver.AIAgentTaskContextHierarchy{},
			Repositories: []riidoaiserver.AIAgentTaskContextRepository{},
		})
	}))
	defer taskContextServer.Close()

	t.Setenv(envTaskContextBaseURL, taskContextServer.URL)
	t.Setenv(envTaskContextWorkspaceID, "workspace-a")
	t.Setenv(envTaskContextTeamID, "RIID")
	t.Setenv(envTaskContextAPIKey, "workspace-key")
	t.Setenv(envTaskContextTimeout, "1")

	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	if config.TaskContextReader == nil {
		t.Fatal("task context reader missing")
	}
	contextSnapshot, err := config.TaskContextReader.GetAIAgentTaskContext(context.Background(), "component-a")
	if err != nil {
		t.Fatalf("GetAIAgentTaskContext: %v", err)
	}
	if gotPath != "/workspaces/workspace-a/open-api/v1/teams/RIID/components/component-a/ai-agent-context" {
		t.Fatalf("task context path = %q", gotPath)
	}
	if gotAPIKey != "workspace-key" {
		t.Fatalf("task context api key = %q", gotAPIKey)
	}
	if contextSnapshot.Component.BranchName != "RIID-4800-server-task-context-http-client-assignment-prompt-wiring" {
		t.Fatalf("task context snapshot = %+v", contextSnapshot)
	}
}

func TestConfigFromEnvParsesPrivateTaskContextReader(t *testing.T) {
	clearRiidoAIServerEnv(t)
	var gotPaths []string
	var gotAuthorization []string
	taskContextServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.String())
		gotAuthorization = append(gotAuthorization, r.Header.Get("Authorization"))
		switch r.URL.Path {
		case "/public/components/component-a/workspace":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":            "component-a",
				"componentType": "task",
				"team": map[string]any{
					"id": "team-a",
					"workspace": map[string]any{
						"id": "workspace-a",
					},
				},
			})
		case "/teams/team-a/components/component-a":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":            "component-a",
				"componentType": "task",
				"title":         "Private task context from existing API server",
				"keyNumber":     "RIID-4873",
				"document": map[string]any{
					"id":               "document-a",
					"tiptapDocumentId": "doc-a",
					"HTMLContent":      "<p>Existing API server private document.</p>",
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer taskContextServer.Close()

	t.Setenv(envTaskContextBaseURL, taskContextServer.URL)
	t.Setenv(envTaskContextTimeout, "1")

	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	if config.TaskContextReader == nil {
		t.Fatal("task context reader missing")
	}
	requestReader, ok := config.TaskContextReader.(riidoaiserver.AIAgentTaskContextRequestReader)
	if !ok {
		t.Fatalf("task context reader should support request-scoped JWT")
	}
	contextSnapshot, err := requestReader.GetAIAgentTaskContextForRequest(context.Background(), riidoaiserver.AIAgentTaskContextRequest{
		ComponentID: "component-a",
		WorkspaceID: "workspace-a",
		BearerToken: "user-jwt",
	})
	if err != nil {
		t.Fatalf("GetAIAgentTaskContextForRequest: %v", err)
	}
	if !reflect.DeepEqual(gotPaths, []string{"/public/components/component-a/workspace", "/teams/team-a/components/component-a?getDocument=true"}) {
		t.Fatalf("task context paths = %v", gotPaths)
	}
	for _, got := range gotAuthorization {
		if got != "Bearer user-jwt" {
			t.Fatalf("authorization = %q", got)
		}
	}
	if contextSnapshot.Component.Title != "Private task context from existing API server" ||
		contextSnapshot.Document.Content != "<p>Existing API server private document.</p>" {
		t.Fatalf("task context snapshot = %+v", contextSnapshot)
	}
}

func TestConfigFromEnvRejectsPartialTaskContextConfig(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envTaskContextBaseURL, "https://api.riido.io")
	t.Setenv(envTaskContextWorkspaceID, "workspace-a")

	if _, err := configFromEnv(); err == nil || !strings.Contains(err.Error(), "OpenAPI task context") {
		t.Fatalf("configFromEnv err=%v", err)
	}
}

func TestConfigFromEnvRejectsAIAgentClientDevelopmentWithoutDynamoDBTable(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envAIAgentClientDev, "true")
	t.Setenv(envAWSRegion, "ap-northeast-2")
	t.Setenv(envAWSContainerCredentialsFullURI, "http://169.254.170.2/credentials")

	if _, err := configFromEnv(); err == nil || !strings.Contains(err.Error(), envAIAgentClientTable) {
		t.Fatalf("configFromEnv err=%v", err)
	}
}

func TestConfigFromEnvRejectsAIAgentClientDevelopmentWithoutCredentialEndpoint(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envAIAgentClientDev, "true")
	t.Setenv(envAIAgentClientTable, "riido-ai-agent-development")
	t.Setenv(envAWSRegion, "ap-northeast-2")

	if _, err := configFromEnv(); err == nil || !strings.Contains(err.Error(), envAWSContainerCredentialsFullURI) {
		t.Fatalf("configFromEnv err=%v", err)
	}
}

func TestParseWebAllowedOriginsRejectsInvalidOrigins(t *testing.T) {
	for _, value := range []string{
		"*",
		"ftp://app.riido.io",
		"https://app.riido.io/path",
		"https://app.riido.io?debug=true",
		"https://user@app.riido.io",
	} {
		t.Run(value, func(t *testing.T) {
			if _, err := parseWebAllowedOrigins(value); err == nil || !strings.Contains(err.Error(), envWebAllowedOrigins) {
				t.Fatalf("parseWebAllowedOrigins err=%v", err)
			}
		})
	}
}

func TestConfigFromEnvIncludesReviewAccountProvisioning(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envReviewAccountTokenHash, testTokenSHA256("review-token"))

	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	if config.ReviewProvision == nil {
		t.Fatal("review provisioning missing")
	}
	if config.ReviewProvision.Credential.Token != "" || config.ReviewProvision.Credential.TokenSHA256 == "" {
		t.Fatalf("review credential should use token hash only: %+v", config.ReviewProvision.Credential)
	}
	if _, err := config.Authorizer.Authorize(context.Background(), "review-token", riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceAgentCatalog,
		Action:   riidoaiserver.AuthorizationActionRead,
	}); err != nil {
		t.Fatalf("review token should read catalog: %v", err)
	}
}

func TestParseAuthzTokensJSONRejectsUnknownField(t *testing.T) {
	if _, err := parseAuthzTokensJSON(`[{"principal_id":"user-a","token":"static-token","scopes":["riido:*"],"extra":true}]`); err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestAuthorizerFromEnvFallsBackFromStaticToExternalOnlyWhenUnauthenticated(t *testing.T) {
	clearRiidoAIServerEnv(t)
	var externalCalls atomic.Int32
	authorizerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		externalCalls.Add(1)
		if got := r.Header.Get(riidoaiserver.ExternalAuthorizerAPIKeyHeader); got != "internal-key" {
			t.Fatalf("external authorizer api key header = %q", got)
		}
		var req struct {
			SchemaVersion string `json:"schema_version"`
			BearerToken   string `json:"bearer_token"`
			Request       struct {
				Resource string `json:"resource"`
				Action   string `json:"action"`
			} `json:"request"`
		}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.SchemaVersion != riidoaiserver.ExternalAuthorizerRequestSchemaVersion || req.BearerToken != "external-token" || req.Request.Resource != "metrics" || req.Request.Action != "read" {
			t.Fatalf("external request = %+v", req)
		}
		_ = json.NewEncoder(w).Encode(struct {
			SchemaVersion string `json:"schema_version"`
			Allowed       bool   `json:"allowed"`
			PrincipalID   string `json:"principal_id"`
		}{
			SchemaVersion: riidoaiserver.ExternalAuthorizerResponseSchemaVersion,
			Allowed:       true,
			PrincipalID:   "external-user",
		})
	}))
	defer authorizerServer.Close()

	t.Setenv(envAuthzTokensJSON, `[{"principal_id":"static-user","token":"static-token","scopes":["metrics:read"]}]`)
	t.Setenv(envExternalAuthzURL, authorizerServer.URL)
	t.Setenv(envExternalAuthzAPIKey, "internal-key")
	t.Setenv(envExternalAuthzTimeout, "1")
	authorizer, err := authorizerFromEnv()
	if err != nil {
		t.Fatalf("authorizerFromEnv: %v", err)
	}

	if _, err := authorizer.Authorize(context.Background(), "static-token", riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceMetrics,
		Action:   riidoaiserver.AuthorizationActionRead,
	}); err != nil {
		t.Fatalf("static authorize: %v", err)
	}
	if got := externalCalls.Load(); got != 0 {
		t.Fatalf("external calls after static token = %d", got)
	}
	if _, err := authorizer.Authorize(context.Background(), "external-token", riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceMetrics,
		Action:   riidoaiserver.AuthorizationActionRead,
	}); err != nil {
		t.Fatalf("external authorize: %v", err)
	}
	if got := externalCalls.Load(); got != 1 {
		t.Fatalf("external calls after fallback = %d", got)
	}
	if _, err := authorizer.Authorize(context.Background(), "static-token", riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceAgent,
		Action:   riidoaiserver.AuthorizationActionPoll,
		AgentID:  "agent-a",
	}); err == nil {
		t.Fatal("expected forbidden static scope to stop fallback")
	}
	if got := externalCalls.Load(); got != 1 {
		t.Fatalf("external should not run after forbidden static scope, calls=%d", got)
	}
}

func TestAuthorizerFromEnvRejectsExternalAPIKeyWithoutEndpoint(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envExternalAuthzAPIKey, "internal-key")
	if _, err := authorizerFromEnv(); err == nil {
		t.Fatal("expected external api key without endpoint to fail")
	}
}

func TestAuthorizerFromEnvIncludesReviewAccountCredential(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envReviewAccountTokenHash, testTokenSHA256("review-token"))

	authorizer, err := authorizerFromEnv()
	if err != nil {
		t.Fatalf("authorizerFromEnv: %v", err)
	}
	if authorizer == nil {
		t.Fatal("authorizer missing")
	}
	if _, err := authorizer.Authorize(context.Background(), "review-token", riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceAgentCatalog,
		Action:   riidoaiserver.AuthorizationActionRead,
	}); err != nil {
		t.Fatalf("review token should read catalog: %v", err)
	}
	if _, err := authorizer.Authorize(context.Background(), "review-token", riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceAgent,
		Action:   riidoaiserver.AuthorizationActionProviderStatusRead,
		AgentID:  "store-review-agent",
	}); err != nil {
		t.Fatalf("review token should read synthetic provider status: %v", err)
	}
	if _, err := authorizer.Authorize(context.Background(), "review-token", riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceAgent,
		Action:   riidoaiserver.AuthorizationActionPoll,
		AgentID:  "store-review-agent",
	}); err == nil {
		t.Fatal("review token must not poll as daemon agent")
	}
}

func TestEnvDurationSecondsRejectsNonPositiveValues(t *testing.T) {
	for _, value := range []string{"0", "-1", "nope"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv(envShutdownTimeoutSeconds, value)
			if _, err := envDurationSeconds(envShutdownTimeoutSeconds, time.Second); err == nil || !strings.Contains(err.Error(), envShutdownTimeoutSeconds) {
				t.Fatalf("envDurationSeconds err=%v", err)
			}
		})
	}
}

func TestEnvOptionalDurationSecondsRejectsNonPositiveValues(t *testing.T) {
	for _, value := range []string{"0", "-1", "nope"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv(envMetricsLogInterval, value)
			if _, err := envOptionalDurationSeconds(envMetricsLogInterval); err == nil || !strings.Contains(err.Error(), envMetricsLogInterval) {
				t.Fatalf("envOptionalDurationSeconds err=%v", err)
			}
		})
	}
}

func TestStartMetricsPublisherWritesCloudWatchEMF(t *testing.T) {
	store := riidoaiserver.NewStoreWithClock(func() time.Time { return time.Unix(2000, 0).UTC() })
	defer store.Close()
	if _, err := store.AssignTask(context.Background(), "task-a", riidoaiserver.AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-a",
		RuntimeProvider: "codex",
		Prompt:          "hello",
	}); err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	writer := metricsCaptureWriter{ch: make(chan string, 1)}
	cancel, errCh := startMetricsPublisher(store, time.Hour, writer)
	defer stopMetricsPublisher(cancel, errCh)

	select {
	case body := <-writer.ch:
		if !strings.Contains(body, "\"_aws\"") || !strings.Contains(body, "\"assignments_total\":1") {
			t.Fatalf("metrics body = %s", body)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for metrics output")
	}
}

type metricsCaptureWriter struct {
	ch chan string
}

func (w metricsCaptureWriter) Write(p []byte) (int, error) {
	w.ch <- string(p)
	return len(p), nil
}

func clearRiidoAIServerEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		envAddr,
		envShutdownTimeoutSeconds,
		envAuthzTokensJSON,
		envExternalAuthzURL,
		envExternalAuthzAudience,
		envExternalAuthzTimeout,
		envReviewAccountTokenHash,
		envMetricsLogInterval,
		envPprofAddr,
		envWebAllowedOrigins,
		envAssignmentActiveLease,
		envAIAgentClientDev,
		envAIAgentClientTable,
		envAWSRegion,
		envDynamoDBEndpoint,
		envAgentProfileThumbnailBucket,
		envAgentProfileThumbnailPrefix,
		envAgentProfileThumbnailCDNBase,
		envAgentProfileThumbnailMaxBytes,
		envAgentProfileThumbnailExpires,
		envAgentProfileThumbnailS3Endpoint,
		envTaskContextBaseURL,
		envTaskContextWorkspaceID,
		envTaskContextTeamID,
		envTaskContextAPIKey,
		envTaskContextTimeout,
		envExternalAuthzAPIKey,
		envAWSContainerCredentialsFullURI,
		envAWSContainerCredentialsRelativeURI,
		envAWSContainerAuthorizationToken,
	} {
		t.Setenv(key, "")
	}
}

func testTokenSHA256(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
