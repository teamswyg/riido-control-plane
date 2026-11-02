package riidoaiserver

import (
	"net/http"
	"testing"
	"time"
)

func TestMemberReactionNotificationRetriesTransientFailures(t *testing.T) {
	previous := memberReactionNotificationRetryDelay
	memberReactionNotificationRetryDelay = time.Millisecond
	t.Cleanup(func() { memberReactionNotificationRetryDelay = previous })
	for _, tc := range []struct {
		name     string
		statuses []int
		want     int
	}{
		{name: "unavailable then delivered", statuses: []int{http.StatusServiceUnavailable}, want: 2},
		{name: "always unavailable", statuses: []int{503, 503, 503, 503, 503}, want: memberReactionNotificationMaxAttempts},
		{name: "not found is final", statuses: []int{http.StatusNotFound}, want: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := &reactionNotificationRecorder{statuses: tc.statuses}
			server := reactionNotificationServer(t, recorder)
			server.queueMemberReactionNotification(memberReactionNotification{key: "thread/message/user-2", revision: 1, body: map[string]string{"emoji": "1f44d"}})
			server.commentReactions.notificationWG.Wait()
			if got := recorder.sent(); len(got) != tc.want {
				t.Fatalf("sent %d times, want %d", len(got), tc.want)
			}
		})
	}
}
