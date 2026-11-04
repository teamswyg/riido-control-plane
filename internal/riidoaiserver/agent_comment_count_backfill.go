package riidoaiserver

import (
	"context"
	"log/slog"
	"time"
)

type agentCommentCountTask struct{ workspaceID, taskID string }

type agentCommentCountTaskLister interface {
	listAgentCommentCountTasks(ctx context.Context) ([]agentCommentCountTask, error)
}

// listAgentCommentCountTasks returns every task with Agent threads and its
// workspace, taken from the thread's agent snapshot or the current agent.
func (s *DevelopmentAIAgentClientStore) listAgentCommentCountTasks(context.Context) ([]agentCommentCountTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tasks := make([]agentCommentCountTask, 0, len(s.taskThreads))
	for taskID, threads := range s.taskThreads {
		for _, thread := range threads {
			workspaceID := ""
			if thread.AgentSnapshot != nil {
				workspaceID = thread.AgentSnapshot.WorkspaceID
			}
			if workspaceID == "" {
				workspaceID = s.agents[thread.AgentID].WorkspaceID
			}
			if workspaceID != "" {
				tasks = append(tasks, agentCommentCountTask{workspaceID: workspaceID, taskID: taskID})
				break
			}
		}
	}
	return tasks, nil
}

func (s *PersistentAIAgentClientStore) listAgentCommentCountTasks(ctx context.Context) ([]agentCommentCountTask, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return nil, err
	}
	return s.DevelopmentAIAgentClientStore.listAgentCommentCountTasks(ctx)
}

// backfillAgentCommentCounts fills tasks from before RIID-5842 once. It pushes
// one task at a time with a pause, since each read takes the store operation lock.
func (s Server) backfillAgentCommentCounts(ctx context.Context) {
	lister, ok := s.aiAgent.(agentCommentCountTaskLister)
	config := s.config.CommentDelivery
	if !ok || s.agentCommentCounts == nil || !s.config.AgentCommentCount || config.EvidenceBaseURL == "" || config.Key == "" {
		slog.WarnContext(ctx, "agent comment count backfill skipped: store or monolith delivery is not configured")
		return
	}
	tasks, err := lister.listAgentCommentCountTasks(ctx)
	if err != nil {
		slog.WarnContext(ctx, "agent comment count backfill failed to list tasks", "error", err)
		return
	}
	for _, task := range tasks {
		pushContext, cancel := context.WithTimeout(ctx, 5*time.Second)
		s.pushAgentCommentCount(pushContext, task.workspaceID+"\x00"+task.taskID, task.workspaceID, task.taskID)
		cancel()
		time.Sleep(100 * time.Millisecond)
	}
	slog.InfoContext(ctx, "agent comment count backfill finished", "tasks", len(tasks))
}
