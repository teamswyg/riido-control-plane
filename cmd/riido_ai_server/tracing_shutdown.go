package main

import (
	"context"
)

func shutdownTracing(shutdown tracingShutdownFunc) {
	if shutdown == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), tracingShutdownTimeout)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		logOTelInternalError(err)
	}
}
