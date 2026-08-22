package main

import (
	"net/http"
	"time"

	controlplaneowner "github.com/teamswyg/riido-control-plane/internal/ownergraphql/controlplane"
	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func newRuntimeHTTPServers(config runtimeConfig, store *riidoaiserver.Store, aiAgentClient riidoaiserver.AIAgentClientStore, httpTransactions *riidoaiserver.HTTPTransactionMetrics, traceRecorder riidoaiserver.TraceRecorder) []*http.Server {
	servers := []*http.Server{
		{
			Addr:              config.Addr,
			Handler:           newRuntimeHandler(config, store, aiAgentClient, httpTransactions, traceRecorder),
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
	if pprofServer := newPprofServer(config.PprofAddr); pprofServer != nil {
		servers = append(servers, pprofServer)
	}
	return servers
}

func newRuntimeHandler(config runtimeConfig, store *riidoaiserver.Store, aiAgentClient riidoaiserver.AIAgentClientStore, httpTransactions *riidoaiserver.HTTPTransactionMetrics, traceRecorder riidoaiserver.TraceRecorder) http.Handler {
	return riidoaiserver.NewServer(riidoaiserver.ServerConfig{
		Assignment:               store,
		AIAgentClient:            aiAgentClient,
		AIAgentProfileThumbnails: config.AIAgentProfileThumbnails,
		TaskContext:              config.TaskContextReader,
		Authorizer:               config.Authorizer,
		WebAllowedOrigins:        config.WebAllowedOrigins,
		HTTPTransactions:         httpTransactions,
		TraceRecorder:            traceRecorder,
		ControlPlaneOwnerGraphQL: controlplaneowner.NewGraphQLHandler(controlplaneowner.NewRuntimeUseCase()),
		LongPollMaxHold:          config.LongPollMaxHold,
		LongPollTick:             config.LongPollTick,
	}).Handler()
}
