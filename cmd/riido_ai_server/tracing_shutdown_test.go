package main

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
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

func TestShutdownTracingLogsFailureWithoutDetails(t *testing.T) {
	var output bytes.Buffer
	previousWriter := log.Writer()
	log.SetOutput(&output)
	defer log.SetOutput(previousWriter)
	shutdownTracing(func(context.Context) error { return errors.New("sensitive detail") })
	if got := output.String(); !strings.Contains(got, "event=otel_internal_error") || strings.Contains(got, "sensitive detail") {
		t.Fatalf("shutdown log = %q", got)
	}
}
