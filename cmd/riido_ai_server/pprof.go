package main

import (
	"net/http"
	"net/http/pprof"
	"strings"
	"time"
)

func newPprofServer(addr string) *http.Server {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return nil
	}
	return &http.Server{
		Addr:              addr,
		Handler:           newPprofHandler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func newPprofHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return mux
}
