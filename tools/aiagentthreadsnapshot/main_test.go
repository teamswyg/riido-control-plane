package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAIAgentThreadSnapshotBehaviorGolden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(fakeHandler))
	defer server.Close()
	t.Setenv("RIIDO_TEST_TOKEN", "secret-token")
	output := filepath.Join(t.TempDir(), "snapshot.json")
	err := runMain([]string{
		"-base-url", server.URL, "-workspace-id", "ws", "-task-id", "task",
		"-conversation-id", "conv", "-token-env", "RIIDO_TEST_TOKEN",
		"-output", output, "-sse-window", "10ms",
	})
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	assertNotContains(t, string(body), "secret-token")
	assertNotContains(t, string(body), "private body")
	var got report
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if !got.Redacted || got.V3.ThreadCount != 1 || len(got.V3.HighlightedThreads) != 1 {
		t.Fatalf("unexpected report: %+v", got)
	}
	if got.ConversationCount != 1 || len(got.Conversations) != 1 {
		t.Fatalf("missing candidate conversations: %+v", got)
	}
	got.CapturedAt = "<captured_at>"
	got.Source.BaseURL = "<base_url>"
	normalized, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(append(normalized, '\n')))
	const want = "d7be2ff40dacf8d3258b2a8154d3d48e785d1fc8b3ee4142a350240c000e3fa9"
	if hash != want {
		t.Fatalf("golden hash = %s", hash)
	}
}

func fakeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") != "Bearer secret-token" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	switch {
	case strings.Contains(r.URL.Path, "/v3/"):
		_, _ = w.Write([]byte(`{"threads":[{"thread_id":"t","conversation_id":"conv","assignment_state":"completed","messages":[{"role":"agent","body":"private body"}]}]}`))
	case strings.HasSuffix(r.URL.Path, "/thread-stream-subscription"):
		_, _ = w.Write([]byte(`{"stream":{"href":"/events","event_type":"agent_thread_progress"},"active_thread_filters":[]}`))
	case strings.HasSuffix(r.URL.Path, "/events"):
		w.Header().Set("Content-Type", "text/event-stream")
	default:
		_, _ = w.Write([]byte(`{"threads":[{"thread_id":"t","assignment_state":"completed","lines":[{"message":"private body"}]}]}`))
	}
}
