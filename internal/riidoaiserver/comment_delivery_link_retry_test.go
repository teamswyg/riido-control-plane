package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// flakyCommentLinkStore fails the first delivery link like a transient
// DynamoDB error.
type flakyCommentLinkStore struct {
	*DevelopmentAIAgentClientStore
	failures int
}

func (s *flakyCommentLinkStore) RecordCommentDeliveryEvent(ctx context.Context, workspaceID, threadID, eventID string) error {
	if s.failures > 0 {
		s.failures--
		return errors.New("transient")
	}
	return s.DevelopmentAIAgentClientStore.RecordCommentDeliveryEvent(ctx, workspaceID, threadID, eventID)
}

func TestCommentDeliveryLinkRetriesAfterTransientFailure(t *testing.T) {
	delay := threadReactionNotificationRetryDelay
	threadReactionNotificationRetryDelay = time.Millisecond
	t.Cleanup(func() { threadReactionNotificationRetryDelay = delay })
	base, _, _, taskID, threadID := seedReactionThread(t)
	monolith, _, release, writes := blockingMonolith(t)
	close(release)
	store := &flakyCommentLinkStore{DevelopmentAIAgentClientStore: base, failures: 1}
	server := NewServer(ServerConfig{AIAgentClient: store, CommentDelivery: CommentDeliveryConfig{EvidenceBaseURL: monolith.URL, Key: strings.Repeat("k", 32)}})

	server.trackCommentDeliveryReaction(context.Background(), defaultAIAgentClientWorkspaceID, taskID, threadID, "event")
	for i := 0; i < 100 && len(writes()) == 0; i++ {
		server.commentReactions.wg.Wait()
		time.Sleep(time.Millisecond)
	}
	if got := strings.Join(writes(), " "); got != `{"emoji":"1f440"}` {
		t.Fatalf("writes = %s, want the started reaction after the retried link", got)
	}
}
