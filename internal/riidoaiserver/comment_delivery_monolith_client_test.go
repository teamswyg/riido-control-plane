package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestThreadReactionNotificationRequestShape(t *testing.T) {
	key := strings.Repeat("k", 32)
	var got map[string]string
	monolith := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/agent-thread-reaction-notifications" || r.Header.Get(CommentDeliveryKeyHeader) != key {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Error(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer monolith.Close()
	config := CommentDeliveryConfig{EvidenceBaseURL: monolith.URL, Key: key}
	err := config.postThreadReactionNotification(context.Background(), agentThreadReactionNotification{
		WorkspaceID: "ws", TaskID: "task", ThreadID: "thread", MessageID: "msg", RecipientPrincipalID: "author", ActorPrincipalID: "actor", Emoji: "1f44d", Excerpt: "hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	for field, want := range map[string]string{"workspace_id": "ws", "task_id": "task", "thread_id": "thread", "message_id": "msg", "recipient_principal_id": "author", "actor_principal_id": "actor", "emoji": "1f44d", "excerpt": "hello"} {
		if got[field] != want {
			t.Fatalf("%s = %q, want %q", field, got[field], want)
		}
	}
}

func TestMonolithStatusErrorSeparatesPermanentFailures(t *testing.T) {
	monolith := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }))
	defer monolith.Close()
	err := CommentDeliveryConfig{EvidenceBaseURL: monolith.URL, Key: strings.Repeat("k", 32)}.putCommentStatusReaction(context.Background(), "event", "")
	var statusErr monolithStatusError
	if !errors.As(err, &statusErr) || !statusErr.permanent() || (monolithStatusError{status: 503}).permanent() {
		t.Fatalf("status error = %v", err)
	}
}
