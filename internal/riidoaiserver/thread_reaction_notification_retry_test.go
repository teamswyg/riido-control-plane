package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestThreadReactionNotificationRetriesTransientFailures(t *testing.T) {
	delay := threadReactionNotificationRetryDelay
	threadReactionNotificationRetryDelay = time.Millisecond
	t.Cleanup(func() { threadReactionNotificationRetryDelay = delay })
	tests := []struct {
		name     string
		statuses []int
		want     int32
	}{
		{name: "retries until monolith accepts", statuses: []int{503, 503, 204}, want: 3},
		{name: "stops after the last attempt", statuses: []int{503, 503, 503, 503, 503}, want: threadReactionNotificationMaxAttempts},
		{name: "never retries a permanent rejection", statuses: []int{400, 204}, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			monolith := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statuses[calls.Add(1)-1])
			}))
			t.Cleanup(monolith.Close)
			server := NewServer(ServerConfig{CommentDelivery: CommentDeliveryConfig{EvidenceBaseURL: monolith.URL, Key: strings.Repeat("k", 32)}})
			server.queueThreadReactionNotification(agentThreadReactionNotification{ThreadID: "thread", MessageID: "msg", ActorPrincipalID: "user-2", Emoji: "1f44d", revision: 1})
			server.commentReactions.wg.Wait()
			if got := calls.Load(); got != tt.want {
				t.Fatalf("monolith calls = %d, want %d", got, tt.want)
			}
		})
	}
}
