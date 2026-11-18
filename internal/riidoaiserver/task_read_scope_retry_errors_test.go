package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestTaskReadScopeRetriesStayBoundedAndFailClosed(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusServiceUnavailable} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			client, initial, _ := taskScopeRetryFixture(t, status, 10, 0)
			got, err := client.GetAIAgentTaskContextForRequest(t.Context(), taskScopeRetryRequest())
			if err == nil || got.ReadScope != nil {
				t.Fatalf("missing initial scope must fail closed: %+v %v", got.ReadScope, err)
			}
			want := int32(1)
			if status == http.StatusServiceUnavailable {
				want = 2
			}
			if initial.Load() != want {
				t.Fatalf("calls=%d want=%d", initial.Load(), want)
			}
		})
	}
}

func TestTaskReadScopeContinuationDenialDoesNotBecomePartial(t *testing.T) {
	client, _, pages := taskScopeRetryFixture(t, http.StatusForbidden, 0, 10)
	got, err := client.GetAIAgentTaskContextForRequest(t.Context(), taskScopeRetryRequest())
	if err == nil || got.ReadScope != nil || pages.Load() != 1 {
		t.Fatalf("denial was retried or hidden: scope=%+v calls=%d err=%v", got.ReadScope, pages.Load(), err)
	}
}

func TestTaskReadScopeRetryHonorsOverallDeadline(t *testing.T) {
	client, initial, _ := taskScopeRetryFixture(t, http.StatusServiceUnavailable, 10, 0)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()
	_, err := client.GetAIAgentTaskContextForRequest(ctx, taskScopeRetryRequest())
	if !errors.Is(err, context.DeadlineExceeded) || initial.Load() > 1 {
		t.Fatalf("deadline extended: calls=%d err=%v", initial.Load(), err)
	}
}
