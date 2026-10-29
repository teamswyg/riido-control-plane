package riidoaiserver

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// blockingMonolith holds the first write until release is closed and records
// every write body in order.
func blockingMonolith(t *testing.T) (*httptest.Server, chan struct{}, chan struct{}, func() []string) {
	t.Helper()
	var mu sync.Mutex
	var bodies []string
	entered, release := make(chan struct{}), make(chan struct{})
	first := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		block := first
		first = false
		mu.Unlock()
		if block {
			close(entered)
			<-release
		}
		mu.Lock()
		bodies = append(bodies, string(body))
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	return server, entered, release, func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), bodies...)
	}
}

func TestCommentStatusReactionRewritesStopDuringInFlightWrite(t *testing.T) {
	store, _, _, taskID, threadID := seedReactionThread(t)
	monolith, entered, release, writes := blockingMonolith(t)
	server := NewServer(ServerConfig{AIAgentClient: store, CommentDelivery: CommentDeliveryConfig{EvidenceBaseURL: monolith.URL, Key: strings.Repeat("k", 32)}})
	if err := store.RecordCommentDeliveryEvent(t.Context(), defaultAIAgentClientWorkspaceID, threadID, "event"); err != nil {
		t.Fatal(err)
	}
	server.reconcileCommentStatusReactions(taskID)
	<-entered
	setReactionThreadState(store, taskID, threadID, AgentAssignmentStateStopped)
	server.reconcileCommentStatusReactions(taskID)
	close(release)
	server.commentReactions.wg.Wait()
	if got := strings.Join(writes(), " "); got != `{"emoji":"1f440"} {"emoji":""}` {
		t.Fatalf("writes = %s, want the started reaction removed after the stop", got)
	}
}

func TestThreadReactionNotificationsSendLatestInOrder(t *testing.T) {
	monolith, entered, release, writes := blockingMonolith(t)
	server := NewServer(ServerConfig{CommentDelivery: CommentDeliveryConfig{EvidenceBaseURL: monolith.URL, Key: strings.Repeat("k", 32)}})
	notification := agentThreadReactionNotification{ThreadID: "thread", MessageID: "msg", ActorPrincipalID: "actor", Emoji: "1f44d"}
	server.queueThreadReactionNotification(notification)
	<-entered
	notification.Emoji = "2764"
	server.queueThreadReactionNotification(notification)
	notification.Emoji = ""
	server.queueThreadReactionNotification(notification)
	close(release)
	server.commentReactions.wg.Wait()
	var emojis []string
	for _, body := range writes() {
		var sent agentThreadReactionNotification
		if err := json.Unmarshal([]byte(body), &sent); err != nil {
			t.Fatal(err)
		}
		emojis = append(emojis, sent.Emoji)
	}
	if strings.Join(emojis, ",") != "1f44d," {
		t.Fatalf("notifications = %q, want the first then only the latest removal", emojis)
	}
}

func TestStatusReactionTargetIsFirstMessageOfRun(t *testing.T) {
	thread := AIAgentTaskThreadRecord{AssignmentID: "a", RunID: "run", SourceMessageID: "follow-up"}
	messages := []AIAgentTaskThreadHistoryMessage{
		{MessageID: "old", Role: AIAgentTaskThreadMessageRoleUser, AssignmentID: "previous"},
		{MessageID: "first", Role: AIAgentTaskThreadMessageRoleUser, AssignmentID: "a", RunID: "run"},
		{MessageID: "second", Role: AIAgentTaskThreadMessageRoleUser, AssignmentID: "a", RunID: "run", SourceMessageID: "follow-up"},
	}
	if got := taskThreadStatusReactionTarget(thread, messages); got != "first" {
		t.Fatalf("target = %q, want the message that started the run", got)
	}
}
