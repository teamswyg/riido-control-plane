package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (s Server) writeMonolithReaction(ctx context.Context, method, path string, body any) error {
	config := s.config.CommentDelivery
	if config.EvidenceBaseURL == "" || config.Key == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(config.EvidenceBaseURL, "/")+path, bytes.NewReader(payload))
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
		return fmt.Errorf("reaction delivery returned %d", response.StatusCode)
	}
	return nil
}

// Read the current projection rather than the arriving event, which can be a retry
// of an older state. Failed writes are repaired by the next history read.
func (s Server) deliverCommentStatusReactions(ctx context.Context, principal AuthorizationResult, taskID string) {
	if s.config.CommentDelivery.EvidenceBaseURL == "" {
		return
	}
	response, err := s.aiAgent.ListAIAgentTaskThreads(ctx, principal, taskID)
	if err != nil {
		return
	}
	for _, thread := range response.Threads {
		call := thread.CommentInvocation
		if call == nil || call.EvidenceEventID == "" || call.OriginMessageID != "" {
			continue
		}
		emoji, known := agentStatusReaction(thread)
		if !known || s.commentReactions.alreadySent(call.EvidenceEventID, emoji) {
			continue
		}
		if err := s.writeMonolithReaction(ctx, http.MethodPut, "/internal/agent-comment-deliveries/"+url.PathEscape(call.EvidenceEventID)+"/reaction", map[string]string{"emoji": emoji}); err != nil {
			slog.WarnContext(ctx, "agent comment reaction pending", "assignment_id", thread.AssignmentID, "error", err)
		} else {
			s.commentReactions.markSent(call.EvidenceEventID, emoji)
		}
	}
}

func (s Server) notifyMemberMessageReaction(ctx context.Context, principal AuthorizationResult, response AIAgentTaskThreadMessageReactionsResponse, emoji string, remove bool) {
	history, err := s.aiAgent.ListAIAgentTaskThreadHistory(ctx, principal, response.TaskID)
	if err != nil {
		return
	}
	for _, thread := range history.Threads {
		if thread.ThreadID == response.ThreadID {
			for _, message := range thread.Messages {
				if message.MessageID != response.MessageID || message.Role != AIAgentTaskThreadMessageRoleUser || message.AuthorPrincipalID == "" || message.AuthorPrincipalID == principal.PrincipalID {
					continue
				}
				if remove {
					emoji = ""
					for _, reaction := range message.Reactions {
						for _, actor := range reaction.Actors {
							if actor.Kind == "member" && actor.ID == principal.PrincipalID {
								emoji = reaction.Emoji
								break
							}
						}
					}
				}
				body := map[string]string{"workspace_id": principal.WorkspaceID, "task_id": response.TaskID, "thread_id": response.ThreadID, "message_id": response.MessageID, "recipient_principal_id": message.AuthorPrincipalID, "actor_principal_id": principal.PrincipalID, "emoji": emoji, "excerpt": message.Body}
				if err := s.writeMonolithReaction(ctx, http.MethodPost, "/internal/agent-thread-reaction-notifications", body); err != nil {
					slog.WarnContext(ctx, "thread reaction notification pending", "message_id", response.MessageID, "error", err)
				}
			}
		}
	}
}
