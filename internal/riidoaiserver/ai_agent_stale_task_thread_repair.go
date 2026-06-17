package riidoaiserver

import (
	"time"
)

const aiAgentClientStaleActiveThreadTTL = 24 * time.Hour

func (s *DevelopmentAIAgentClientStore) repairStaleActiveTaskThreads(now time.Time) bool {
	if s == nil {
		return false
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.repairStaleActiveTaskThreadsLocked(now.UTC())
}

func (s *DevelopmentAIAgentClientStore) repairStaleActiveTaskThreadsLocked(now time.Time) bool {
	changed := false
	for taskID, threads := range s.taskThreads {
		for i := range threads {
			if !staleActiveTaskThread(threads[i], now) {
				continue
			}
			applyStaleActiveTaskThreadProjection(&threads[i], now)
			changed = true
		}
		s.taskThreads[taskID] = threads
	}
	if changed {
		s.reprojectAgentsFromTaskThreadsLocked()
	}
	return changed
}

func staleActiveTaskThread(thread AIAgentTaskThreadRecord, now time.Time) bool {
	if !taskThreadHasActiveStream(thread) {
		return false
	}
	lastObservedAt := taskThreadLastObservedAt(thread)
	if lastObservedAt.IsZero() || now.Before(lastObservedAt) {
		return false
	}
	return now.Sub(lastObservedAt) > aiAgentClientStaleActiveThreadTTL
}

func taskThreadLastObservedAt(thread AIAgentTaskThreadRecord) time.Time {
	lastObservedAt := thread.StartedAt.UTC()
	for _, line := range thread.Lines {
		if line.ObservedAt.After(lastObservedAt) {
			lastObservedAt = line.ObservedAt.UTC()
		}
	}
	return lastObservedAt
}
