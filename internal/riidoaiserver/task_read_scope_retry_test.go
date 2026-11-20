package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestTaskReadScopeRecoversTransientInitialRead(t *testing.T) {
	client, initial, pages := taskScopeRetryFixture(t, http.StatusServiceUnavailable, 1, 0)
	got, err := client.GetAIAgentTaskContextForRequest(t.Context(), taskScopeRetryRequest())
	if err != nil || got.ReadScope == nil || got.ReadScope.Status != "available" {
		t.Fatalf("transient scope read rejected assignment context: scope=%+v err=%v", got.ReadScope, err)
	}
	if initial.Load() != 2 || pages.Load() != 1 {
		t.Fatalf("initial=%d pages=%d, want 2/1", initial.Load(), pages.Load())
	}
}

func TestTaskReadScopeKeepsPrefixAfterTransientContinuationExhaustion(t *testing.T) {
	client, initial, pages := taskScopeRetryFixture(t, http.StatusServiceUnavailable, 0, 10)
	got, err := client.GetAIAgentTaskContextForRequest(t.Context(), taskScopeRetryRequest())
	if err != nil || got.ReadScope == nil || got.ReadScope.Status != "partial" {
		t.Fatalf("captured prefix discarded: scope=%+v err=%v", got.ReadScope, err)
	}
	if initial.Load() != 1 || pages.Load() != 2 {
		t.Fatalf("initial=%d pages=%d, want 1/2", initial.Load(), pages.Load())
	}
	for _, part := range got.readScopeCapture.Manifest {
		if part.Part == AssignmentContentEditor && (part.Complete || len(part.Cursors) != 1) {
			t.Fatalf("invented continuation or completeness: %+v", part)
		}
	}
}
