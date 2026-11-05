package riidoaiserver

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFailureStageKeepsErrorIdentityAndMessage(t *testing.T) {
	cause := latestTaskDocumentError{withFailureStage("read_scope", taskContextReadError{"denied"})}
	err := withFailureStage("compose", cause)
	var latest latestTaskDocumentError
	if !errors.As(err, &latest) || err.Error() != cause.Error() {
		t.Fatalf("stage wrapper changed the error: %v", err)
	}
	if outer, inner := failureStages(err); outer != "compose" || inner != "read_scope" {
		t.Fatalf("stages = %q %q", outer, inner)
	}
	if withFailureStage("compose", nil) != nil {
		t.Fatal("nil error must stay nil")
	}
	for wrapped, want := range map[error]int{err: 503, withFailureStage("client_assignment", ErrAIAgentNotFound): 404} {
		recorder := httptest.NewRecorder()
		writeCommentDeliveryError(recorder, wrapped)
		if recorder.Code != want {
			t.Fatalf("status changed: %d want %d", recorder.Code, want)
		}
	}
}

func TestFailureKindIsBounded(t *testing.T) {
	expired, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	cases := map[string]struct {
		ctx context.Context
		err error
	}{
		"timeout":             {context.Background(), withFailureStage("assign", context.DeadlineExceeded)},
		"context_denied":      {context.Background(), latestTaskDocumentError{taskContextReadError{"denied"}}},
		"context_unknown":     {context.Background(), taskContextReadError{"something new"}},
		"snapshot_conflict":   {context.Background(), withFailureStage("client_assignment", errAIAgentSnapshotConflict)},
		"membership_changed":  {context.Background(), errTaskAgentSnapshotChanged},
		"thread_conflict":     {context.Background(), ErrAIAgentTaskThreadConflict},
		"denied":              {context.Background(), ErrAuthorizationForbidden},
		"not_found":           {context.Background(), ErrAIAgentNotFound},
		"context_unavailable": {context.Background(), latestTaskDocumentError{errors.New("x")}},
		"other":               {context.Background(), errors.New("x")},
	}
	for want, tc := range cases {
		if got := failureKind(tc.ctx, tc.err); got != want {
			t.Errorf("%s: got %q", want, got)
		}
	}
	if got := failureKind(expired, errors.New("wrapped deadline")); got != "timeout" {
		t.Errorf("expired context: got %q", got)
	}
}

func TestCommentDeliveryFailureLogOmitsContent(t *testing.T) {
	var buf bytes.Buffer
	previous, flags := log.Writer(), log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer func() { log.SetOutput(previous); log.SetFlags(flags) }()
	secret := "comment body with private text"
	err := withFailureStage("client_assignment", errors.New(secret))
	logCommentDeliveryFailure(context.Background(), "comment", "", "comment:event-1", time.Now(), err)
	line := buf.String()
	for _, want := range []string{"event=comment_delivery_failed", `stage="client_assignment"`, `failure_kind="other"`, `path="comment"`} {
		if !strings.Contains(line, want) {
			t.Fatalf("log missing %s: %s", want, line)
		}
	}
	if strings.Contains(line, secret) || strings.Contains(line, "event-1") {
		t.Fatalf("log leaked content: %s", line)
	}
	buf.Reset()
	logTaskContextUnavailable(context.Background(), latestTaskDocumentError{withFailureStage("read_scope", taskContextReadError{"denied"})})
	if line := buf.String(); !strings.Contains(line, `stage="read_scope"`) || !strings.Contains(line, `failure_kind="context_denied"`) {
		t.Fatalf("task context log: %s", line)
	}
}
