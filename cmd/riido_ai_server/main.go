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
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

const (
	envAddr                    = "RIIDO_AI_SERVER_ADDR"
	envShutdownTimeoutSeconds  = "RIIDO_AI_SERVER_SHUTDOWN_TIMEOUT_SECONDS"
	envAuthzTokensJSON         = "RIIDO_AI_SERVER_AUTHZ_TOKENS_JSON"
	envExternalAuthzURL        = "RIIDO_AI_SERVER_EXTERNAL_AUTHZ_URL"
	envExternalAuthzAudience   = "RIIDO_AI_SERVER_EXTERNAL_AUTHZ_AUDIENCE"
	envExternalAuthzTimeout    = "RIIDO_AI_SERVER_EXTERNAL_AUTHZ_TIMEOUT_SECONDS"
	envReviewAccountTokenHash  = "RIIDO_AI_SERVER_REVIEW_ACCOUNT_TOKEN_SHA256"
	envMetricsLogInterval      = "RIIDO_AI_SERVER_METRICS_LOG_INTERVAL_SECONDS"
	envWebAllowedOrigins       = "RIIDO_AI_SERVER_WEB_ALLOWED_ORIGINS"
	envAIAgentClientTable      = "RIIDO_AI_SERVER_AI_AGENT_CLIENT_DYNAMODB_TABLE"
	envDynamoDBAssignmentTable = "RIIDO_AI_SERVER_DYNAMODB_ASSIGNMENT_TABLE"
	envDynamoDBOutboxTable     = "RIIDO_AI_SERVER_DYNAMODB_OUTBOX_TABLE"
	envDynamoDBEndpoint        = "RIIDO_AI_SERVER_DYNAMODB_ENDPOINT"
	envAWSRegion               = "RIIDO_AI_SERVER_AWS_REGION"
	envAssignmentActiveLease   = "RIIDO_AI_SERVER_ASSIGNMENT_ACTIVE_LEASE_SECONDS"
	envTaskContextBaseURL      = "RIIDO_AI_SERVER_TASK_CONTEXT_BASE_URL"
	envTaskContextWorkspaceID  = "RIIDO_AI_SERVER_TASK_CONTEXT_WORKSPACE_ID"
	envTaskContextTeamID       = "RIIDO_AI_SERVER_TASK_CONTEXT_TEAM_ID"
	envTaskContextAPIKey       = "RIIDO_AI_SERVER_TASK_CONTEXT_WORKSPACE_API_KEY"
	envTaskContextTimeout      = "RIIDO_AI_SERVER_TASK_CONTEXT_TIMEOUT_SECONDS"
)

type runtimeConfig struct {
	Addr               string
	ShutdownTimeout    time.Duration
	Authorizer         riidoaiserver.RequestAuthorizer
	ReviewProvision    *riidoaiserver.ReviewAccountProvisioning
	MetricsLogInterval time.Duration
	WebAllowedOrigins  []string
	AIAgentClientTable string
	AssignmentTable    string
	OutboxTable        string
	DynamoDBEndpoint   string
	AWSRegion          string
	ActiveLease        time.Duration
	TaskContextReader  riidoaiserver.AIAgentTaskContextReader
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
	awsRuntime, err := openDynamoDBRuntime(config)
	if err != nil {
		return err
	}
	defer awsRuntime.close()
	if awsRuntime.aiAgentClient == nil {
		return fmt.Errorf("%s is required for the development AI Agent client store", envAIAgentClientTable)
	}
	aiAgentClient, err := riidoaiserver.NewDevelopmentAIAgentClientStore(context.Background(), riidoaiserver.DevelopmentAIAgentClientStoreConfig{
		Persistence: awsRuntime.aiAgentClient,
	})
	if err != nil {
		return fmt.Errorf("open AI Agent client store: %w", err)
	}
	defer closeAIAgentClient(aiAgentClient)
	agentRegistry := riidoaiserver.NewCompositeAgentRegistry(aiAgentClient)
	store, err := riidoaiserver.OpenStoreWithConfig(context.Background(), riidoaiserver.StoreConfig{
		ActiveLeaseDuration: config.ActiveLease,
		Outbox:              awsRuntime.outbox,
		SnapshotStore:       awsRuntime.assignmentSnapshot,
		OperationStore:      awsRuntime.assignmentOperations,
		AgentRegistry:       agentRegistry,
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
	activeLease, err := envDurationSeconds(envAssignmentActiveLease, time.Duration(riidoaiserver.DefaultAssignmentActiveLeaseSeconds)*time.Second)
	if err != nil {
		return runtimeConfig{}, err
	}
	taskContextReader, err := taskContextReaderFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	return runtimeConfig{
		Addr:               getenvDefault(envAddr, ":8080"),
		ShutdownTimeout:    shutdownTimeout,
		Authorizer:         authorizer,
		ReviewProvision:    reviewProvision,
		MetricsLogInterval: metricsLogInterval,
		WebAllowedOrigins:  webAllowedOrigins,
		AIAgentClientTable: firstNonEmpty(os.Getenv(envAIAgentClientTable), os.Getenv(envDynamoDBAssignmentTable)),
		AssignmentTable:    strings.TrimSpace(os.Getenv(envDynamoDBAssignmentTable)),
		OutboxTable:        strings.TrimSpace(os.Getenv(envDynamoDBOutboxTable)),
		DynamoDBEndpoint:   strings.TrimSpace(os.Getenv(envDynamoDBEndpoint)),
		AWSRegion:          firstNonEmpty(os.Getenv(envAWSRegion), os.Getenv("AWS_REGION")),
		ActiveLease:        activeLease,
		TaskContextReader:  taskContextReader,
	}, nil
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
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
	if baseURL == "" || workspaceID == "" || teamID == "" || apiKey == "" {
		return nil, fmt.Errorf("%s, %s, %s, and %s must be set together", envTaskContextBaseURL, envTaskContextWorkspaceID, envTaskContextTeamID, envTaskContextAPIKey)
	}
	timeout, err := envDurationSeconds(envTaskContextTimeout, 0)
	if err != nil {
		return nil, err
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

type dynamoDBRuntime struct {
	aiAgentClient        riidoaiserver.AIAgentClientPersistence
	assignmentSnapshot   riidoaiserver.SnapshotStore
	assignmentOperations riidoaiserver.AssignmentOperationStore
	outbox               riidoaiserver.EventSink
	closers              []func() error
}

func (r dynamoDBRuntime) close() {
	for i := len(r.closers) - 1; i >= 0; i-- {
		_ = r.closers[i]()
	}
}

func openDynamoDBRuntime(config runtimeConfig) (dynamoDBRuntime, error) {
	region := strings.TrimSpace(config.AWSRegion)
	if region == "" && (config.AIAgentClientTable != "" || config.AssignmentTable != "" || config.OutboxTable != "") {
		return dynamoDBRuntime{}, fmt.Errorf("%s or AWS_REGION is required when DynamoDB stores are configured", envAWSRegion)
	}
	provider, err := awsCredentialsProviderFromEnv()
	if err != nil {
		if config.AIAgentClientTable == "" && config.AssignmentTable == "" && config.OutboxTable == "" {
			return dynamoDBRuntime{}, nil
		}
		return dynamoDBRuntime{}, err
	}
	var runtime dynamoDBRuntime
	if table := strings.TrimSpace(config.AIAgentClientTable); table != "" {
		store, err := riidoaiserver.NewDynamoDBAIAgentClientSnapshot(riidoaiserver.DynamoDBAIAgentClientSnapshotConfig{
			Region:              region,
			TableName:           table,
			Endpoint:            config.DynamoDBEndpoint,
			CredentialsProvider: provider,
		})
		if err != nil {
			return dynamoDBRuntime{}, err
		}
		runtime.aiAgentClient = store
		runtime.closers = append(runtime.closers, store.Close)
	}
	if table := strings.TrimSpace(config.AssignmentTable); table != "" {
		snapshot, err := riidoaiserver.NewDynamoDBStoreSnapshot(riidoaiserver.DynamoDBStoreSnapshotConfig{
			Region:              region,
			TableName:           table,
			Endpoint:            config.DynamoDBEndpoint,
			CredentialsProvider: provider,
		})
		if err != nil {
			runtime.close()
			return dynamoDBRuntime{}, err
		}
		operations, err := riidoaiserver.NewDynamoDBAssignmentOperationStore(riidoaiserver.DynamoDBAssignmentOperationStoreConfig{
			Region:              region,
			TableName:           table,
			Endpoint:            config.DynamoDBEndpoint,
			ActiveLeaseDuration: config.ActiveLease,
			CredentialsProvider: provider,
		})
		if err != nil {
			runtime.close()
			return dynamoDBRuntime{}, err
		}
		runtime.assignmentSnapshot = snapshot
		runtime.assignmentOperations = operations
		runtime.closers = append(runtime.closers, snapshot.Close, operations.Close)
	}
	if table := strings.TrimSpace(config.OutboxTable); table != "" {
		outbox, err := riidoaiserver.NewDynamoDBOutbox(riidoaiserver.DynamoDBOutboxConfig{
			Region:              region,
			TableName:           table,
			Endpoint:            config.DynamoDBEndpoint,
			CredentialsProvider: provider,
		})
		if err != nil {
			runtime.close()
			return dynamoDBRuntime{}, err
		}
		runtime.outbox = outbox
		runtime.closers = append(runtime.closers, outbox.Close)
	}
	return runtime, nil
}

func awsCredentialsProviderFromEnv() (riidoaiserver.AWSCredentialsProvider, error) {
	fullURI := strings.TrimSpace(os.Getenv("AWS_CONTAINER_CREDENTIALS_FULL_URI"))
	relativeURI := strings.TrimSpace(os.Getenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI"))
	endpoint := fullURI
	if endpoint == "" && relativeURI != "" {
		endpoint = "http://169.254.170.2/" + strings.TrimPrefix(path.Clean(relativeURI), "/")
	}
	if endpoint == "" {
		return nil, errors.New("AWS_CONTAINER_CREDENTIALS_FULL_URI or AWS_CONTAINER_CREDENTIALS_RELATIVE_URI is required for DynamoDB stores")
	}
	token := strings.TrimSpace(os.Getenv("AWS_CONTAINER_AUTHORIZATION_TOKEN"))
	if token == "" {
		tokenFile := strings.TrimSpace(os.Getenv("AWS_CONTAINER_AUTHORIZATION_TOKEN_FILE"))
		if tokenFile != "" {
			body, err := os.ReadFile(tokenFile)
			if err != nil {
				return nil, fmt.Errorf("read AWS container authorization token file: %w", err)
			}
			token = strings.TrimSpace(string(body))
		}
	}
	return riidoaiserver.NewECSContainerCredentialsProvider(riidoaiserver.ECSContainerCredentialsProviderConfig{
		Endpoint:           endpoint,
		AuthorizationToken: token,
	})
}

func closeAIAgentClient(store interface{ Close() error }) {
	_ = store.Close()
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
	if endpoint := strings.TrimSpace(os.Getenv(envExternalAuthzURL)); endpoint != "" {
		timeout, err := envDurationSeconds(envExternalAuthzTimeout, 0)
		if err != nil {
			return nil, err
		}
		authorizer, err := riidoaiserver.NewExternalHTTPAuthorizer(riidoaiserver.ExternalHTTPAuthorizerConfig{
			Endpoint: endpoint,
			Audience: strings.TrimSpace(os.Getenv(envExternalAuthzAudience)),
			Timeout:  timeout,
		})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", envExternalAuthzURL, err)
		}
		authorizers = append(authorizers, authorizer)
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
