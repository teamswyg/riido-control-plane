package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RIID-5842: Riido task lists count only monolith comments. After a task's
// Agent conversation changes, push the absolute number of visible Agent
// conversation messages to the monolith, which adds it to commentsCount.
// Failed pushes are repaired by the next change or history read of the task.

const agentCommentCountPath = "/internal/agent-comment-counts"

// Mirrors the client's visible-message rule (isVisibleThreadMessage):
// progress/tool rows and run-status kinds stay in the activity area.
var agentCommentCountHiddenKinds = map[AgentTaskCommentKind]bool{
	AgentTaskCommentQueuedByBusyAgent: true,
	AgentTaskCommentAssignmentStarted: true,
	AgentTaskCommentRuntimeProgress:   true,
}

// countVisibleAgentCommentMessages counts messages the comments tab shows for the
// task. User messages copied from Riido comments are already monolith comments;
// replies typed in the Agent thread carry a "thread-message:" source and count.
func countVisibleAgentCommentMessages(threads []AIAgentTaskThreadHistoryRecord) int {
	seen := map[string]bool{}
	for _, thread := range threads {
		for index, message := range thread.Messages {
			if message.Role != AIAgentTaskThreadMessageRoleUser && message.Role != AIAgentTaskThreadMessageRoleAgent {
				continue
			}
			if agentCommentCountHiddenKinds[message.CommentKind] {
				continue
			}
			body := message.ResultMessage
			if strings.TrimSpace(body) == "" {
				body = message.Body
			}
			if strings.TrimSpace(body) == "" {
				continue
			}
			source := strings.TrimSpace(message.SourceMessageID)
			if message.Role == AIAgentTaskThreadMessageRoleUser && source != "" && !strings.HasPrefix(source, "thread-message:") {
				continue
			}
			key := message.MessageID
			if key == "" {
				key = source
			}
			if key == "" {
				key = thread.ThreadID + ":" + strconv.Itoa(index)
			}
			seen[key] = true
		}
	}
	return len(seen)
}

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

// scheduleAgentCommentCount pushes the task's current count in the background.
// It is a no-op unless the monolith comment-delivery URL and key are configured.
func (s Server) scheduleAgentCommentCount(ctx context.Context, workspaceID, taskID string) {
	d := s.agentCommentCounts
	config := s.config.CommentDelivery
	workspaceID, taskID = strings.TrimSpace(workspaceID), strings.TrimSpace(taskID)
	if d == nil || s.aiAgent == nil || config.EvidenceBaseURL == "" || config.Key == "" || workspaceID == "" || taskID == "" {
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

func (s Server) pushAgentCommentCount(ctx context.Context, key, workspaceID, taskID string) {
	observedAt := time.Now().UTC()
	// The list count is task-wide, not per viewer, so read every thread of the task.
	principal := AuthorizationResult{WorkspaceID: workspaceID, Roles: []AgentCatalogRole{AgentCatalogRoleAdmin}}
	history, err := s.aiAgent.ListAIAgentTaskThreadHistory(ctx, principal, taskID)
	if err != nil {
		slog.WarnContext(ctx, "agent comment count pending", "task_id", taskID, "error", err)
		return
	}
	count := countVisibleAgentCommentMessages(history.Threads)
	if s.agentCommentCounts.isAlreadySent(key, count) {
		return
	}
	body := map[string]any{"workspace_id": workspaceID, "task_id": taskID, "message_count": count, "observed_at": observedAt}
	if err := s.writeAgentCommentCount(ctx, body); err != nil {
		slog.WarnContext(ctx, "agent comment count pending", "task_id", taskID, "error", err)
		return
	}
	s.agentCommentCounts.markSent(key, count)
}

func (s Server) writeAgentCommentCount(ctx context.Context, body any) error {
	config := s.config.CommentDelivery
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, strings.TrimRight(config.EvidenceBaseURL, "/")+agentCommentCountPath, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set(CommentDeliveryKeyHeader, config.Key)
	req.Header.Set("Content-Type", "application/json")
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("agent comment count returned %d", response.StatusCode)
	}
	return nil
}
