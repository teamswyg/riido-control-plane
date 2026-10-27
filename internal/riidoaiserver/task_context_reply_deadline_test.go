package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCurrentReplyMetadataHonorsRequestDeadline(t *testing.T) {
	release := make(chan struct{})
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	t.Cleanup(upstream.Close)
	t.Cleanup(func() { close(release) })
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL})
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	if got := client.readCurrentReplyMetadata(ctx, "team", "viewer", []AIAgentTaskContextComment{{ID: "parent", ReplyCount: 1}}); len(got) != 0 || ctx.Err() != context.DeadlineExceeded {
		t.Fatal("reply query did not fail closed on the original deadline")
	}
}
