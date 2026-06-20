package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func shutdownAfterBackgroundError(servers []*http.Server, timeout time.Duration, errCh <-chan error, err error) error {
	if shutdownErr := shutdownServers(servers, timeout); shutdownErr != nil {
		return fmt.Errorf("shutdown after metrics publisher error: %w", shutdownErr)
	}
	if serverErr := waitForServers(errCh, len(servers)); serverErr != nil {
		return serverErr
	}
	return err
}

func shutdownServers(servers []*http.Server, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var shutdownErr error
	for _, server := range servers {
		if err := server.Shutdown(ctx); err != nil && shutdownErr == nil {
			shutdownErr = err
		}
	}
	return shutdownErr
}

func waitForServers(errCh <-chan error, count int) error {
	var firstErr error
	for range count {
		if err := <-errCh; err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
