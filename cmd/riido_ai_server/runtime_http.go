package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/controlplanegraphql"
	"github.com/teamswyg/riido-control-plane/internal/controlplanehealth"
	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func newRuntimeHTTPServers(config runtimeConfig, store *riidoaiserver.Store, aiAgentClient riidoaiserver.AIAgentClientStore, httpTransactions *riidoaiserver.HTTPTransactionMetrics, traceRecorder riidoaiserver.TraceRecorder) ([]*http.Server, error) {
	graphqlHandler, graphqlErr := controlplanegraphql.NewHandler(controlplanehealth.NewService())
	handler, err := newRuntimeHandlerWithGraphQL(config, store, aiAgentClient, httpTransactions, traceRecorder, graphqlHandler, graphqlErr)
	if err != nil {
		return nil, err
	}
	servers := []*http.Server{
		{
			Addr:              config.Addr,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
	if err := validateControlPlaneGraphQLMTLSRuntimeConfig(config); err != nil {
		return nil, err
	}
	if config.ControlPlaneGraphQLMTLS.TLSConfig != nil {
		mux := http.NewServeMux()
		mux.Handle("/graphql", graphqlHandler)
		servers = append(servers, &http.Server{
			Addr:              config.ControlPlaneGraphQLMTLS.Addr,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
			TLSConfig:         config.ControlPlaneGraphQLMTLS.TLSConfig.Clone(),
		})
	}
	if pprofServer := newPprofServer(config.PprofAddr); pprofServer != nil {
		servers = append(servers, pprofServer)
	}
	return servers, nil
}

func validateControlPlaneGraphQLMTLSRuntimeConfig(config runtimeConfig) error {
	mtls := config.ControlPlaneGraphQLMTLS
	if mtls.Addr == "" && mtls.TLSConfig == nil {
		return nil
	}
	if mtls.Addr == "" || mtls.TLSConfig == nil {
		return fmt.Errorf("control-plane GraphQL mTLS address and TLS configuration must be provided together")
	}
	if mtls.Addr == config.Addr || mtls.Addr == config.PprofAddr {
		return fmt.Errorf("control-plane GraphQL mTLS address must use a dedicated listener")
	}
	if mtls.TLSConfig.MinVersion != tls.VersionTLS13 || mtls.TLSConfig.MaxVersion != tls.VersionTLS13 ||
		mtls.TLSConfig.ClientAuth != tls.RequireAndVerifyClientCert || mtls.TLSConfig.ClientCAs == nil ||
		len(mtls.TLSConfig.Certificates) != 1 || mtls.TLSConfig.VerifyConnection == nil {
		return fmt.Errorf("control-plane GraphQL listener requires the exact TLS 1.3 mutual-authentication policy")
	}
	return nil
}

func newRuntimeHandler(config runtimeConfig, store *riidoaiserver.Store, aiAgentClient riidoaiserver.AIAgentClientStore, httpTransactions *riidoaiserver.HTTPTransactionMetrics, traceRecorder riidoaiserver.TraceRecorder) (http.Handler, error) {
	graphqlHandler, err := controlplanegraphql.NewHandler(controlplanehealth.NewService())
	return newRuntimeHandlerWithGraphQL(config, store, aiAgentClient, httpTransactions, traceRecorder, graphqlHandler, err)
}

func newRuntimeHandlerWithGraphQL(config runtimeConfig, store *riidoaiserver.Store, aiAgentClient riidoaiserver.AIAgentClientStore, httpTransactions *riidoaiserver.HTTPTransactionMetrics, traceRecorder riidoaiserver.TraceRecorder, graphqlHandler http.Handler, graphqlErr error) (http.Handler, error) {
	if graphqlErr != nil {
		return nil, fmt.Errorf("open control-plane GraphQL receiver: %w", graphqlErr)
	}
	if graphqlHandler == nil {
		return nil, fmt.Errorf("open control-plane GraphQL receiver: handler is nil")
	}
	return riidoaiserver.NewServer(riidoaiserver.ServerConfig{
		Assignment:               store,
		AIAgentClient:            aiAgentClient,
		AIAgentProfileThumbnails: config.AIAgentProfileThumbnails,
		TaskContext:              config.TaskContextReader,
		Authorizer:               config.Authorizer,
		WebAllowedOrigins:        config.WebAllowedOrigins,
		HTTPTransactions:         httpTransactions,
		TraceRecorder:            traceRecorder,
		ControlPlaneGraphQL:      graphqlHandler,
		LongPollMaxHold:          config.LongPollMaxHold,
		LongPollTick:             config.LongPollTick,
	}).Handler(), nil
}
