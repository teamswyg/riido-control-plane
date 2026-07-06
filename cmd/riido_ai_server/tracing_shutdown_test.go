package main

import (
	"context"
	"testing"
)

func TestShutdownTracingCallsConfiguredShutdown(t *testing.T) {
	called := false
	shutdownTracing(func(ctx context.Context) error {
		called = true
		if err := ctx.Err(); err != nil {
			t.Fatalf("shutdown context already done: %v", err)
		}
		return nil
	})
	if !called {
		t.Fatal("shutdown was not called")
	}
	shutdownTracing(nil)
}
