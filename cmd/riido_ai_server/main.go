package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

const (
	envAddr                   = "RIIDO_AI_SERVER_ADDR"
	envShutdownTimeoutSeconds = "RIIDO_AI_SERVER_SHUTDOWN_TIMEOUT_SECONDS"
	envAuthzTokensJSON        = "RIIDO_AI_SERVER_AUTHZ_TOKENS_JSON"
	envExternalAuthzURL       = "RIIDO_AI_SERVER_EXTERNAL_AUTHZ_URL"
	envExternalAuthzAudience  = "RIIDO_AI_SERVER_EXTERNAL_AUTHZ_AUDIENCE"
	envExternalAuthzAPIKey    = "RIIDO_AI_SERVER_EXTERNAL_AUTHZ_API_KEY"
	envExternalAuthzTimeout   = "RIIDO_AI_SERVER_EXTERNAL_AUTHZ_TIMEOUT_SECONDS"
	envReviewAccountTokenHash = "RIIDO_AI_SERVER_REVIEW_ACCOUNT_TOKEN_SHA256"
	envMetricsLogInterval     = "RIIDO_AI_SERVER_METRICS_LOG_INTERVAL_SECONDS"
	envWebAllowedOrigins      = "RIIDO_AI_SERVER_WEB_ALLOWED_ORIGINS"
	envAIAgentClientDev       = "RIIDO_AI_SERVER_AI_AGENT_CLIENT_DEVELOPMENT"
	envAIAgentClientTable     = "RIIDO_AI_SERVER_AI_AGENT_CLIENT_DYNAMODB_TABLE"
	envAWSRegion              = "RIIDO_AI_SERVER_AWS_REGION"
	envDynamoDBEndpoint       = "RIIDO_AI_SERVER_DYNAMODB_ENDPOINT"
	envTaskContextBaseURL     = "RIIDO_AI_SERVER_TASK_CONTEXT_BASE_URL"
	envTaskContextWorkspaceID = "RIIDO_AI_SERVER_TASK_CONTEXT_WORKSPACE_ID"
	envTaskContextTeamID      = "RIIDO_AI_SERVER_TASK_CONTEXT_TEAM_ID"
	envTaskContextAPIKey      = "RIIDO_AI_SERVER_TASK_CONTEXT_WORKSPACE_API_KEY"
	envTaskContextTimeout     = "RIIDO_AI_SERVER_TASK_CONTEXT_TIMEOUT_SECONDS"

	envAWSContainerCredentialsFullURI     = "AWS_CONTAINER_CREDENTIALS_FULL_URI"
	envAWSContainerCredentialsRelativeURI = "AWS_CONTAINER_CREDENTIALS_RELATIVE_URI"
	envAWSContainerAuthorizationToken     = "AWS_CONTAINER_AUTHORIZATION_TOKEN"
	awsECSCredentialsBaseURL              = "http://169.254.170.2"
)

type runtimeConfig struct {
	Addr                     string
	ShutdownTimeout          time.Duration
	Authorizer               riidoaiserver.RequestAuthorizer
	ReviewProvision          *riidoaiserver.ReviewAccountProvisioning
	MetricsLogInterval       time.Duration
	WebAllowedOrigins        []string
	AIAgentClientDev         bool
	AIAgentClientStore       riidoaiserver.AIAgentClientSnapshotStore
	AssignmentOperationStore riidoaiserver.AssignmentOperationStore
	TaskContextReader        riidoaiserver.AIAgentTaskContextReader
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "riido_ai_server:", err)
		os.Exit(1)
	}
}

func run() error {
	config, err := configFromEnv()
	if err != nil {
		return err
	}
	defer closeRuntimeConfig(config)
	var aiAgentClient riidoaiserver.AIAgentClientStore
	if config.AIAgentClientDev {
		aiAgentClient, err = riidoaiserver.OpenPersistentAIAgentClientStore(context.Background(), riidoaiserver.NewDevelopmentAIAgentClientStore(), config.AIAgentClientStore)
		if err != nil {
			return fmt.Errorf("open AI Agent development store: %w", err)
		}
	}
	store, err := riidoaiserver.OpenStoreWithConfig(context.Background(), riidoaiserver.StoreConfig{
		AgentRegistry:  riidoaiserver.NewCompositeAgentRegistry(agentRegistryFromAIAgentClient(aiAgentClient)),
		OperationStore: config.AssignmentOperationStore,
	})
	if err != nil {
		return fmt.Errorf("open assignment store: %w", err)
	}
	defer store.Close()
	if config.ReviewProvision != nil {
		if err := store.ApplyReviewAccountProvisioning(context.Background(), *config.ReviewProvision); err != nil {
			return fmt.Errorf("apply review account provisioning: %w", err)
		}
	}
	server := &http.Server{
		Addr:              config.Addr,
		Handler:           riidoaiserver.NewServer(riidoaiserver.ServerConfig{Assignment: store, AIAgentClient: aiAgentClient, TaskContext: config.TaskContextReader, Authorizer: config.Authorizer, WebAllowedOrigins: config.WebAllowedOrigins}).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	metricsCancel, metricsErrCh := startMetricsPublisher(store, config.MetricsLogInterval, os.Stdout)
	defer stopMetricsPublisher(metricsCancel, metricsErrCh)
	return serveUntilSignal(server, config.ShutdownTimeout, metricsErrCh)
}

func configFromEnv() (runtimeConfig, error) {
	shutdownTimeout, err := envDurationSeconds(envShutdownTimeoutSeconds, 10*time.Second)
	if err != nil {
		return runtimeConfig{}, err
	}
	metricsLogInterval, err := envOptionalDurationSeconds(envMetricsLogInterval)
	if err != nil {
		return runtimeConfig{}, err
	}
	reviewProvision, err := reviewAccountProvisioningFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	authorizer, err := authorizerFromEnvWithReview(reviewProvision)
	if err != nil {
		return runtimeConfig{}, err
	}
	webAllowedOrigins, err := webAllowedOriginsFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	aiAgentClientDev, err := aiAgentClientDevelopmentFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	aiAgentClientStore, err := aiAgentClientSnapshotStoreFromEnv(aiAgentClientDev)
	if err != nil {
		return runtimeConfig{}, err
	}
	assignmentOperationStore, err := assignmentOperationStoreFromEnv(aiAgentClientDev)
	if err != nil {
		return runtimeConfig{}, err
	}
	taskContextReader, err := taskContextReaderFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	return runtimeConfig{
		Addr:                     getenvDefault(envAddr, ":8080"),
		ShutdownTimeout:          shutdownTimeout,
		Authorizer:               authorizer,
		ReviewProvision:          reviewProvision,
		MetricsLogInterval:       metricsLogInterval,
		WebAllowedOrigins:        webAllowedOrigins,
		AIAgentClientDev:         aiAgentClientDev,
		AIAgentClientStore:       aiAgentClientStore,
		AssignmentOperationStore: assignmentOperationStore,
		TaskContextReader:        taskContextReader,
	}, nil
}

func closeRuntimeConfig(config runtimeConfig) {
	if closer, ok := config.AIAgentClientStore.(interface{ Close() error }); ok {
		_ = closer.Close()
	}
	if closer, ok := config.AssignmentOperationStore.(interface{ Close() error }); ok {
		_ = closer.Close()
	}
}

func agentRegistryFromAIAgentClient(client riidoaiserver.AIAgentClientStore) riidoaiserver.AgentRegistry {
	if registry, ok := client.(riidoaiserver.AgentRegistry); ok {
		return registry
	}
	return nil
}

func serveUntilSignal(server *http.Server, shutdownTimeout time.Duration, backgroundErrCh ...<-chan error) error {
	var bgErrCh <-chan error
	if len(backgroundErrCh) > 0 {
		bgErrCh = backgroundErrCh[0]
	}
	errCh := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigCh)

	select {
	case sig := <-sigCh:
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown after %s: %w", sig, err)
		}
		return <-errCh
	case err := <-bgErrCh:
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if shutdownErr := server.Shutdown(ctx); shutdownErr != nil {
			return fmt.Errorf("shutdown after metrics publisher error: %w", shutdownErr)
		}
		if serverErr := <-errCh; serverErr != nil {
			return serverErr
		}
		return err
	case err := <-errCh:
		return err
	}
}

func getenvDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envDurationSeconds(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return time.Duration(seconds) * time.Second, nil
}

func envOptionalDurationSeconds(key string) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return 0, nil
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return time.Duration(seconds) * time.Second, nil
}

func envOptionalBool(key string) (bool, error) {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if raw == "" {
		return false, nil
	}
	switch raw {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("%s must be a boolean", key)
	}
}

func aiAgentClientDevelopmentFromEnv() (bool, error) {
	return envOptionalBool(envAIAgentClientDev)
}

func aiAgentClientSnapshotStoreFromEnv(enabled bool) (riidoaiserver.AIAgentClientSnapshotStore, error) {
	if !enabled {
		return nil, nil
	}
	tableName := strings.TrimSpace(os.Getenv(envAIAgentClientTable))
	if tableName == "" {
		return nil, fmt.Errorf("%s is required when %s is enabled", envAIAgentClientTable, envAIAgentClientDev)
	}
	region := strings.TrimSpace(os.Getenv(envAWSRegion))
	if region == "" {
		return nil, fmt.Errorf("%s is required when %s is enabled", envAWSRegion, envAIAgentClientDev)
	}
	provider, err := awsContainerCredentialsProviderFromEnv()
	if err != nil {
		return nil, err
	}
	store, err := riidoaiserver.NewDynamoDBAIAgentClientSnapshot(riidoaiserver.DynamoDBAIAgentClientSnapshotConfig{
		Region:              region,
		TableName:           tableName,
		Endpoint:            strings.TrimSpace(os.Getenv(envDynamoDBEndpoint)),
		CredentialsProvider: provider,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", envAIAgentClientTable, err)
	}
	return store, nil
}

func assignmentOperationStoreFromEnv(enabled bool) (riidoaiserver.AssignmentOperationStore, error) {
	if !enabled {
		return nil, nil
	}
	tableName := strings.TrimSpace(os.Getenv(envAIAgentClientTable))
	if tableName == "" {
		return nil, fmt.Errorf("%s is required when %s is enabled", envAIAgentClientTable, envAIAgentClientDev)
	}
	region := strings.TrimSpace(os.Getenv(envAWSRegion))
	if region == "" {
		return nil, fmt.Errorf("%s is required when %s is enabled", envAWSRegion, envAIAgentClientDev)
	}
	provider, err := awsContainerCredentialsProviderFromEnv()
	if err != nil {
		return nil, err
	}
	store, err := riidoaiserver.NewDynamoDBAssignmentOperationStore(riidoaiserver.DynamoDBAssignmentOperationStoreConfig{
		Region:              region,
		TableName:           tableName,
		Endpoint:            strings.TrimSpace(os.Getenv(envDynamoDBEndpoint)),
		CredentialsProvider: provider,
	})
	if err != nil {
		return nil, fmt.Errorf("%s assignment operation store: %w", envAIAgentClientTable, err)
	}
	return store, nil
}

func awsContainerCredentialsProviderFromEnv() (riidoaiserver.AWSCredentialsProvider, error) {
	endpoint := strings.TrimSpace(os.Getenv(envAWSContainerCredentialsFullURI))
	if endpoint == "" {
		relativeURI := strings.TrimSpace(os.Getenv(envAWSContainerCredentialsRelativeURI))
		if relativeURI != "" {
			if !strings.HasPrefix(relativeURI, "/") {
				return nil, fmt.Errorf("%s must start with /", envAWSContainerCredentialsRelativeURI)
			}
			endpoint = awsECSCredentialsBaseURL + relativeURI
		}
	}
	if endpoint == "" {
		return nil, fmt.Errorf("%s or %s is required when %s is enabled", envAWSContainerCredentialsFullURI, envAWSContainerCredentialsRelativeURI, envAIAgentClientDev)
	}
	return riidoaiserver.NewECSContainerCredentialsProvider(riidoaiserver.ECSContainerCredentialsProviderConfig{
		Endpoint:           endpoint,
		AuthorizationToken: strings.TrimSpace(os.Getenv(envAWSContainerAuthorizationToken)),
	})
}

func webAllowedOriginsFromEnv() ([]string, error) {
	return parseWebAllowedOrigins(os.Getenv(envWebAllowedOrigins))
}

func taskContextReaderFromEnv() (riidoaiserver.AIAgentTaskContextReader, error) {
	baseURL := strings.TrimSpace(os.Getenv(envTaskContextBaseURL))
	workspaceID := strings.TrimSpace(os.Getenv(envTaskContextWorkspaceID))
	teamID := strings.TrimSpace(os.Getenv(envTaskContextTeamID))
	apiKey := strings.TrimSpace(os.Getenv(envTaskContextAPIKey))
	timeoutRaw := strings.TrimSpace(os.Getenv(envTaskContextTimeout))
	if baseURL == "" && workspaceID == "" && teamID == "" && apiKey == "" && timeoutRaw == "" {
		return nil, nil
	}
	if baseURL == "" {
		return nil, fmt.Errorf("%s is required when task context configuration is set", envTaskContextBaseURL)
	}
	timeout, err := envDurationSeconds(envTaskContextTimeout, 0)
	if err != nil {
		return nil, err
	}
	if workspaceID == "" && teamID == "" && apiKey == "" {
		client, err := riidoaiserver.NewAIAgentPrivateTaskContextClient(riidoaiserver.AIAgentPrivateTaskContextClientConfig{
			BaseURL: baseURL,
			Timeout: timeout,
		})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", envTaskContextBaseURL, err)
		}
		return client, nil
	}
	if workspaceID == "" || teamID == "" || apiKey == "" {
		return nil, fmt.Errorf("%s, %s, and %s must be set together for OpenAPI task context; omit all three to use private JWT task context", envTaskContextWorkspaceID, envTaskContextTeamID, envTaskContextAPIKey)
	}
	client, err := riidoaiserver.NewAIAgentTaskContextClient(riidoaiserver.AIAgentTaskContextClientConfig{
		BaseURL:         baseURL,
		WorkspaceID:     workspaceID,
		TeamID:          teamID,
		WorkspaceAPIKey: apiKey,
		Timeout:         timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", envTaskContextBaseURL, err)
	}
	return client, nil
}

func parseWebAllowedOrigins(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	seen := map[string]struct{}{}
	var origins []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		origin, err := normalizeWebOrigin(part)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", envWebAllowedOrigins, err)
		}
		if _, ok := seen[origin]; ok {
			continue
		}
		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}
	return origins, nil
}

func normalizeWebOrigin(raw string) (string, error) {
	if raw == "*" {
		return "", errors.New("wildcard origin is not supported")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse origin %q: %w", raw, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("origin %q must use http or https", raw)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("origin %q must include a host", raw)
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return "", fmt.Errorf("origin %q must not include path, query, fragment, or userinfo", raw)
	}
	return parsed.Scheme + "://" + parsed.Host, nil
}

func startMetricsPublisher(metrics riidoaiserver.MetricsReader, interval time.Duration, writer io.Writer) (context.CancelFunc, <-chan error) {
	if interval <= 0 {
		return func() {}, nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		err := riidoaiserver.RunCloudWatchEMFPublisher(ctx, metrics, interval, riidoaiserver.CloudWatchEMFConfig{Writer: writer})
		if errors.Is(err, context.Canceled) {
			err = nil
		}
		errCh <- err
		close(errCh)
	}()
	return cancel, errCh
}

func stopMetricsPublisher(cancel context.CancelFunc, errCh <-chan error) {
	cancel()
	if errCh != nil {
		<-errCh
	}
}

func authorizerFromEnv() (riidoaiserver.RequestAuthorizer, error) {
	reviewProvision, err := reviewAccountProvisioningFromEnv()
	if err != nil {
		return nil, err
	}
	return authorizerFromEnvWithReview(reviewProvision)
}

func authorizerFromEnvWithReview(reviewProvision *riidoaiserver.ReviewAccountProvisioning) (riidoaiserver.RequestAuthorizer, error) {
	var authorizers []riidoaiserver.RequestAuthorizer
	credentials, err := parseAuthzTokenCredentialsJSON(os.Getenv(envAuthzTokensJSON))
	if err != nil {
		return nil, err
	}
	if reviewProvision != nil {
		credentials = append(credentials, reviewProvision.Credential)
	}
	if len(credentials) > 0 {
		authorizer, err := riidoaiserver.NewStaticTokenAuthorizer(credentials)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", envAuthzTokensJSON, err)
		}
		authorizers = append(authorizers, authorizer)
	}
	externalAuthzAPIKey := strings.TrimSpace(os.Getenv(envExternalAuthzAPIKey))
	if endpoint := strings.TrimSpace(os.Getenv(envExternalAuthzURL)); endpoint != "" {
		timeout, err := envDurationSeconds(envExternalAuthzTimeout, 0)
		if err != nil {
			return nil, err
		}
		authorizer, err := riidoaiserver.NewExternalHTTPAuthorizer(riidoaiserver.ExternalHTTPAuthorizerConfig{
			Endpoint: endpoint,
			Audience: strings.TrimSpace(os.Getenv(envExternalAuthzAudience)),
			APIKey:   externalAuthzAPIKey,
			Timeout:  timeout,
		})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", envExternalAuthzURL, err)
		}
		authorizers = append(authorizers, authorizer)
	} else if externalAuthzAPIKey != "" {
		return nil, fmt.Errorf("%s requires %s", envExternalAuthzAPIKey, envExternalAuthzURL)
	}
	switch len(authorizers) {
	case 0:
		return nil, nil
	case 1:
		return authorizers[0], nil
	default:
		return riidoaiserver.NewFallbackAuthorizer(authorizers...)
	}
}

func parseAuthzTokensJSON(raw string) (*riidoaiserver.StaticTokenAuthorizer, error) {
	credentials, err := parseAuthzTokenCredentialsJSON(raw)
	if err != nil {
		return nil, err
	}
	if len(credentials) == 0 {
		return nil, nil
	}
	return riidoaiserver.NewStaticTokenAuthorizer(credentials)
}

func parseAuthzTokenCredentialsJSON(raw string) ([]riidoaiserver.StaticTokenCredential, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var credentials []riidoaiserver.StaticTokenCredential
	if err := strictDecodeJSON(raw, &credentials); err != nil {
		return nil, fmt.Errorf("%s: %w", envAuthzTokensJSON, err)
	}
	return credentials, nil
}

func reviewAccountProvisioningFromEnv() (*riidoaiserver.ReviewAccountProvisioning, error) {
	tokenHash := strings.TrimSpace(os.Getenv(envReviewAccountTokenHash))
	if tokenHash == "" {
		return nil, nil
	}
	seed, err := riidoaiserver.LoadReviewAccountSeed()
	if err != nil {
		return nil, fmt.Errorf("%s load seed: %w", envReviewAccountTokenHash, err)
	}
	provisioning, err := riidoaiserver.ProvisionReviewAccount(seed, riidoaiserver.ReviewAccountProvisionInput{
		TokenSHA256: tokenHash,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", envReviewAccountTokenHash, err)
	}
	return &provisioning, nil
}

func strictDecodeJSON(raw string, out any) error {
	dec := json.NewDecoder(strings.NewReader(raw))
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
