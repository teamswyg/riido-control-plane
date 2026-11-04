package riidoaiserver

import (
	"context"
	"fmt"
)

// AssignmentNotificationStore never claims work or refreshes a lease. Signals
// are hints; durable state and the existing authorized poll remain authoritative.
type AssignmentNotificationStore interface {
	WatchAssignmentChanges(context.Context, []string) (<-chan struct{}, func(), error)
	AssignmentSignalVersions(context.Context, []string) (map[string]string, error)
}

type assignmentSignalsCmd struct {
	agentIDs []string
	reply    chan map[string]string
}

func (s *Store) WatchAssignmentChanges(ctx context.Context, ids []string) (<-chan struct{}, func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	ch, release := s.assignmentChanges.subscribe(ids)
	return ch, release, nil
}

func (s *Store) AssignmentSignalVersions(ctx context.Context, ids []string) (map[string]string, error) {
	if queue, ok := s.operationStore.(AssignmentQueueReader); ok {
		return s.durableAssignmentSignalVersions(ctx, queue, ids)
	}
	reply := make(chan map[string]string, 1)
	if err := s.send(ctx, assignmentSignalsCmd{agentIDs: ids, reply: reply}); err != nil {
		return nil, err
	}
	select {
	case versions := <-reply:
		return versions, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.done:
		return nil, fmt.Errorf("assignment store closed")
	}
}
