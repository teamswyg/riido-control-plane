package main

import (
	"errors"
	"testing"
	"time"
)

func TestMergeBackgroundErrorsReturnsNilWhenNoChannels(t *testing.T) {
	if got := mergeBackgroundErrors(nil, nil); got != nil {
		t.Fatalf("mergeBackgroundErrors(nil) = %v, want nil", got)
	}
}

func TestMergeBackgroundErrorsForwardsFirstError(t *testing.T) {
	want := errors.New("background failed")
	ch := make(chan error, 1)
	ch <- want
	close(ch)

	merged := mergeBackgroundErrors(nil, ch)
	if merged == nil {
		t.Fatal("mergeBackgroundErrors returned nil")
	}
	select {
	case got := <-merged:
		if !errors.Is(got, want) {
			t.Fatalf("merged error = %v, want %v", got, want)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for merged background error")
	}
}
