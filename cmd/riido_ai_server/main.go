package main

import (
	"context"
	"fmt"
	"os"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

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
	traceRecorder, traceShutdown, err := openTracing(context.Background(), config.Tracing)
	if err != nil {
		return err
	}
	defer shutdownTracing(traceShutdown)
	defer closeRuntimeConfig(config)

	aiAgentClient, err := openAIAgentClient(context.Background(), config)
	if err != nil {
		return err
	}
	store, err := openAssignmentStore(context.Background(), config, aiAgentClient, traceRecorder)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := applyReviewAccountProvisioning(context.Background(), store, config); err != nil {
		return err
	}
	httpTransactions := riidoaiserver.NewHTTPTransactionMetrics()
	servers, err := newRuntimeHTTPServers(config, store, aiAgentClient, httpTransactions, traceRecorder)
	if err != nil {
		return err
	}
	metricsCancel, metricsErrCh := startMetricsPublisher(
		riidoaiserver.NewObservedMetricsReader(store, httpTransactions, config.AIAgentClientMetrics),
		config.MetricsLogInterval,
		os.Stdout,
	)
	defer stopMetricsPublisher(metricsCancel, metricsErrCh)
	return serveUntilSignal(servers, config.ShutdownTimeout, metricsErrCh)
}
