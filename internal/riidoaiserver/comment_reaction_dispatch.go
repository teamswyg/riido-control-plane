package riidoaiserver

import (
	"context"
	"sync"
	"time"
)

type commentReactionJob struct {
	dirty bool
	done  chan struct{}
}
type commentReactionDispatcher struct {
	mu   sync.Mutex
	jobs map[string]*commentReactionJob
	sent map[string]string
}

func (d *commentReactionDispatcher) alreadySent(id, emoji string) bool {
	if d == nil {
		return false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	value, ok := d.sent[id]
	return ok && value == emoji
}
func (d *commentReactionDispatcher) markSent(id, emoji string) {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.sent) >= 2048 {
		clear(d.sent)
	}
	d.sent[id] = emoji
}
func (s Server) scheduleCommentStatusReactions(ctx context.Context, principal AuthorizationResult, taskID string) <-chan struct{} {
	d := s.commentReactions
	if d == nil || s.config.CommentDelivery.EvidenceBaseURL == "" {
		done := make(chan struct{})
		close(done)
		return done
	}
	key := principal.WorkspaceID + ":" + taskID
	d.mu.Lock()
	if job := d.jobs[key]; job != nil {
		job.dirty = true
		d.mu.Unlock()
		return job.done
	}
	job := &commentReactionJob{done: make(chan struct{})}
	d.jobs[key] = job
	d.mu.Unlock()
	go func() {
		for {
			deliveryContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
			s.deliverCommentStatusReactions(deliveryContext, principal, taskID)
			cancel()
			d.mu.Lock()
			if job.dirty {
				job.dirty = false
				d.mu.Unlock()
				continue
			}
			delete(d.jobs, key)
			close(job.done)
			d.mu.Unlock()
			return
		}
	}()
	return job.done
}
func (s Server) syncCommentStatusReactions(ctx context.Context, principal AuthorizationResult, taskID string) {
	done := s.scheduleCommentStatusReactions(ctx, principal, taskID)
	select {
	case <-done:
	case <-ctx.Done():
	case <-time.After(3 * time.Second):
	}
}
