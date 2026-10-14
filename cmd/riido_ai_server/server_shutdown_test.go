package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWaitForServersReturnsFirstError(t *testing.T) {
	want := errors.New("server failed")
	errCh := make(chan error, 2)
	errCh <- nil
	errCh <- want

	if got := waitForServers(errCh, 2); !errors.Is(got, want) {
		t.Fatalf("waitForServers error = %v, want %v", got, want)
	}
}

func TestShutdownServersStopsHTTPServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	if err := shutdownServers([]*http.Server{server.Config}, time.Second); err != nil {
		t.Fatalf("shutdownServers error = %v", err)
	}
}

func TestShutdownAfterBackgroundErrorReturnsServerErrorFirst(t *testing.T) {
	want := errors.New("listen failed")
	errCh := make(chan error, 1)
	errCh <- want

	got := shutdownAfterBackgroundError([]*http.Server{{}}, time.Second, errCh, errors.New("background failed"))
	if !errors.Is(got, want) {
		t.Fatalf("shutdownAfterBackgroundError error = %v, want %v", got, want)
	}
}

func TestShutdownAfterBackgroundErrorReturnsBackgroundError(t *testing.T) {
	want := errors.New("background failed")
	if got := shutdownAfterBackgroundError(nil, time.Second, nil, want); !errors.Is(got, want) {
		t.Fatalf("shutdownAfterBackgroundError error = %v, want %v", got, want)
	}
}
