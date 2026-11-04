package riidoaiserver

import (
	"context"
	"strings"
	"sync"
)

type agentCommentCountJob struct{ isDirty bool }

// agentCommentCountDispatcher runs at most one push per task and folds changes
// that arrive meanwhile into one more push. sent skips unchanged counts.
type agentCommentCountDispatcher struct {
	mu   sync.Mutex
	jobs map[string]*agentCommentCountJob
	sent map[string]int
}

func newAgentCommentCountDispatcher() *agentCommentCountDispatcher {
	return &agentCommentCountDispatcher{jobs: map[string]*agentCommentCountJob{}, sent: map[string]int{}}
}

func (d *agentCommentCountDispatcher) isAlreadySent(key string, count int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	value, ok := d.sent[key]
	return ok && value == count
}

func (d *agentCommentCountDispatcher) markSent(key string, count int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.sent) >= 4096 {
		clear(d.sent)
	}
	d.sent[key] = count
}

// repairAgentCommentCount pushes once for a task this process has not pushed yet,
// so history reads fill tasks from before this change or after a failed push
// without recomputing on every read. Changes still call scheduleAgentCommentCount.
func (s Server) repairAgentCommentCount(ctx context.Context, workspaceID, taskID string) {
	d := s.agentCommentCounts
	if d == nil {
		return
	}
	d.mu.Lock()
	_, isSent := d.sent[strings.TrimSpace(workspaceID)+"\x00"+strings.TrimSpace(taskID)]
	d.mu.Unlock()
	if !isSent {
		s.scheduleAgentCommentCount(ctx, workspaceID, taskID)
	}
}
