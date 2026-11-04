package riidoaiserver

import (
	"context"
	"strings"
	"time"
)

// scheduleAgentCommentCount pushes the task's current count in the background.
// It is a no-op unless enabled and the monolith comment-delivery URL and key are configured.
func (s Server) scheduleAgentCommentCount(ctx context.Context, workspaceID, taskID string) {
	d := s.agentCommentCounts
	config := s.config.CommentDelivery
	workspaceID, taskID = strings.TrimSpace(workspaceID), strings.TrimSpace(taskID)
	if d == nil || !s.config.AgentCommentCount || s.aiAgent == nil || config.EvidenceBaseURL == "" || config.Key == "" || workspaceID == "" || taskID == "" {
		return
	}
	key := workspaceID + "\x00" + taskID
	d.mu.Lock()
	if job := d.jobs[key]; job != nil {
		job.isDirty = true
		d.mu.Unlock()
		return
	}
	job := &agentCommentCountJob{}
	d.jobs[key] = job
	d.mu.Unlock()
	go func() {
		for {
			pushContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			s.pushAgentCommentCount(pushContext, key, workspaceID, taskID)
			cancel()
			d.mu.Lock()
			if job.isDirty {
				job.isDirty = false
				d.mu.Unlock()
				continue
			}
			delete(d.jobs, key)
			d.mu.Unlock()
			return
		}
	}()
}
