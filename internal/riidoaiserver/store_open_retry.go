package riidoaiserver

import (
	"context"
	"strings"
	"time"
)

const storeOpenRetryMaxRetries = 30

func loadAssignmentOperationsWithOpenRetry(
	ctx context.Context,
	loader AssignmentOperationLoader,
) ([]AssignmentOperationRecord, error) {
	var records []AssignmentOperationRecord
	err := retryStoreOpenTransient(ctx, func() error {
		var loadErr error
		records, loadErr = loader.LoadAssignmentOperations(ctx)
		return loadErr
	})
	return records, err
}

func loadReplayAssignmentProjectionsWithOpenRetry(
	ctx context.Context,
	state *storeState,
	reader AssignmentProjectionReader,
) ([]AssignmentProjection, error) {
	var projections []AssignmentProjection
	err := retryStoreOpenTransient(ctx, func() error {
		var loadErr error
		projections, loadErr = loadReplayAssignmentProjections(ctx, state, reader)
		return loadErr
	})
	return projections, err
}

func retryStoreOpenTransient(ctx context.Context, op func() error) error {
	for attempt := 0; ; attempt++ {
		err := op()
		if err == nil || !isTransientStoreOpenError(err) || attempt >= storeOpenRetryMaxRetries {
			return err
		}
		if waitErr := sleepStoreOpenRetry(ctx, storeOpenRetryDelay(attempt)); waitErr != nil {
			return waitErr
		}
	}
}

func sleepStoreOpenRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func isTransientStoreOpenError(err error) bool {
	if err == nil {
		return false
	}
	text := err.Error()
	for _, marker := range transientStoreOpenErrorMarkers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}
