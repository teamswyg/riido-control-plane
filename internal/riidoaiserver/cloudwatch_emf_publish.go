package riidoaiserver

import (
	"context"
	"errors"
	"time"
)

func PublishCloudWatchEMF(ctx context.Context, metrics MetricsReader, config CloudWatchEMFConfig) error {
	if metrics == nil {
		return errors.New("riidoaiserver: metrics reader is required")
	}
	snapshot, err := metrics.Metrics(ctx)
	if err != nil {
		return err
	}
	return WriteCloudWatchEMF(config.Writer, normalizeCloudWatchEMFConfig(config), snapshot)
}

func RunCloudWatchEMFPublisher(ctx context.Context, metrics MetricsReader, interval time.Duration, config CloudWatchEMFConfig) error {
	if interval <= 0 {
		return errors.New("riidoaiserver: metrics interval must be positive")
	}
	if err := PublishCloudWatchEMF(ctx, metrics, config); err != nil {
		return err
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := PublishCloudWatchEMF(ctx, metrics, config); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
