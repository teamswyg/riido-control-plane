package riidoaiserver

import (
	"context"
	"sync"
)

const dynamoDBAIAgentClientSnapshotPartWriteParallelism = 4

func (s *DynamoDBAIAgentClientSnapshot) putSnapshotItems(ctx context.Context, items []map[string]map[string]string, credentials AWSCredentials) error {
	plan := dynamoDBAIAgentClientSnapshotPlanWrites(items)
	if err := s.putSnapshotPartItems(ctx, plan.parts, credentials); err != nil {
		return err
	}
	if plan.manifest == nil {
		return nil
	}
	return s.putSnapshotItem(ctx, plan.manifest, credentials)
}

func (s *DynamoDBAIAgentClientSnapshot) putSnapshotPartItems(ctx context.Context, items []map[string]map[string]string, credentials AWSCredentials) error {
	if len(items) <= 1 {
		return s.putSnapshotPartItemsSequentially(ctx, items, credentials)
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errs := make(chan error, 1)
	sem := make(chan struct{}, dynamoDBAIAgentClientSnapshotPartWriteParallelism)
	var wg sync.WaitGroup
	for _, item := range items {
		if err := ctx.Err(); err != nil {
			break
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(item map[string]map[string]string) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := s.putSnapshotItem(ctx, item, credentials); err != nil {
				select {
				case errs <- err:
					cancel()
				default:
				}
			}
		}(item)
	}
	wg.Wait()
	select {
	case err := <-errs:
		return err
	default:
		return ctx.Err()
	}
}

func (s *DynamoDBAIAgentClientSnapshot) putSnapshotPartItemsSequentially(ctx context.Context, items []map[string]map[string]string, credentials AWSCredentials) error {
	for _, item := range items {
		if err := s.putSnapshotItem(ctx, item, credentials); err != nil {
			return err
		}
	}
	return nil
}
