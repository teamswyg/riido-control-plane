package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
	envAgentBindingsJSON      = "RIIDO_AI_SERVER_AGENT_BINDINGS_JSON"
	envAuthzTokensJSON        = "RIIDO_AI_SERVER_AUTHZ_TOKENS_JSON"
	envExternalAuthzURL       = "RIIDO_AI_SERVER_EXTERNAL_AUTHZ_URL"
	envExternalAuthzAudience  = "RIIDO_AI_SERVER_EXTERNAL_AUTHZ_AUDIENCE"
	envExternalAuthzTimeout   = "RIIDO_AI_SERVER_EXTERNAL_AUTHZ_TIMEOUT_SECONDS"
	envReviewAccountTokenHash = "RIIDO_AI_SERVER_REVIEW_ACCOUNT_TOKEN_SHA256"
	envMetricsLogInterval     = "RIIDO_AI_SERVER_METRICS_LOG_INTERVAL_SECONDS"
)

type runtimeConfig struct {
	Addr               string
	ShutdownTimeout    time.Duration
	AgentRegistry      riidoaiserver.AgentRegistry
	Authorizer         riidoaiserver.RequestAuthorizer
	ReviewProvision    *riidoaiserver.ReviewAccountProvisioning
	MetricsLogInterval time.Duration
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
	store := riidoaiserver.NewStoreWithConfig(riidoaiserver.StoreConfig{AgentRegistry: config.AgentRegistry})
	defer store.Close()
	if config.ReviewProvision != nil {
		if err := store.ApplyReviewAccountProvisioning(context.Background(), *config.ReviewProvision); err != nil {
			return fmt.Errorf("apply review account provisioning: %w", err)
		}
	}
	server := &http.Server{
		Addr:              config.Addr,
		Handler:           riidoaiserver.NewServer(riidoaiserver.ServerConfig{Assignment: store, Authorizer: config.Authorizer}).Handler(),
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
	registry, err := agentRegistryFromEnv()
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
	return runtimeConfig{
		Addr:               getenvDefault(envAddr, ":8080"),
		ShutdownTimeout:    shutdownTimeout,
		AgentRegistry:      registry,
		Authorizer:         authorizer,
		ReviewProvision:    reviewProvision,
		MetricsLogInterval: metricsLogInterval,
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

func agentRegistryFromEnv() (riidoaiserver.AgentRegistry, error) {
	raw := strings.TrimSpace(os.Getenv(envAgentBindingsJSON))
	if raw == "" {
		return nil, nil
	}
	return parseAgentRegistryJSON(raw)
}

func parseAgentRegistryJSON(raw string) (*riidoaiserver.StaticAgentRegistry, error) {
	var bindings []riidoaiserver.AgentRuntimeBinding
	if err := strictDecodeJSON(raw, &bindings); err != nil {
		return nil, fmt.Errorf("%s: %w", envAgentBindingsJSON, err)
	}
	return riidoaiserver.NewStaticAgentRegistry(bindings)
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
