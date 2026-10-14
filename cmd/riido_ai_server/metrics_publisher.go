package main

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

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
