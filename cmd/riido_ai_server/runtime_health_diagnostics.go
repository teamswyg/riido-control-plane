package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
)

func observeRuntimeConnections() func(net.Conn, http.ConnState) {
	var first sync.Once
	return func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			first.Do(func() { fmt.Fprintln(os.Stderr, "riido_ai_server: http_connection_received") })
		}
	}
}

func observeRuntimeReadyProbe(next http.Handler) http.Handler {
	var first sync.Once
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/readyz" {
			first.Do(func() { fmt.Fprintln(os.Stderr, "riido_ai_server: ready_probe_received") })
		}
		next.ServeHTTP(w, r)
	})
}
