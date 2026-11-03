package riidoaiserver

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type dashboardTerminalCache struct {
	mu    sync.Mutex
	byRun map[[2]string]time.Time
}

// A late session or recovery update advances UpdatedAt even after completion.
// Read the immutable terminal event before computing page order and daily counts.
func (s *DynamoDBAssignmentOperationStore) firstDashboardTerminal(ctx context.Context, creds AWSCredentials, a Assignment, frontier int64) (time.Time, error) {
	cache := &s.dashboardTerminalTimes
	key := [2]string{a.AgentID, a.ID}
	cache.mu.Lock()
	first := cache.byRun[key]
	cache.mu.Unlock()
	if !first.IsZero() {
		return first, nil
	}
	for _, kind := range dashboardOperationPrefixes {
		err := s.dashboardQueryItems(ctx, creds, a.AgentID, kind+a.ID+":", "events_json", true, func(item map[string]map[string]string) error {
			var events []TaskEvent
			if err := json.Unmarshal([]byte(dynamoDBStringValue(item, "events_json")), &events); err != nil {
				return err
			}
			for _, e := range events {
				if e.AssignmentID == a.ID && e.AgentID == a.AgentID && e.TaskID == a.TaskID && e.Seq <= frontier && isDashboardTerminal(e.State) && !e.At.IsZero() && (first.IsZero() || e.At.Before(first)) {
					first = e.At
				}
			}
			return nil
		})
		if err != nil {
			return time.Time{}, err
		}
	}
	if !first.IsZero() {
		cache.mu.Lock()
		if cache.byRun == nil || len(cache.byRun) >= 1024 {
			cache.byRun = map[[2]string]time.Time{}
		}
		cache.byRun[key] = first
		cache.mu.Unlock()
	}
	return first, nil
}
