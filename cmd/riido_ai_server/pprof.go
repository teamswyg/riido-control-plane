package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"
)

func parsePprofAddr(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	switch strings.ToLower(raw) {
	case "0", "false", "off", "disabled":
		return "", nil
	}
	host, port, err := net.SplitHostPort(raw)
	if err != nil {
		if !strings.HasPrefix(raw, ":") {
			return "", fmt.Errorf("%s must be a loopback host:port", envPprofAddr)
		}
		host = "127.0.0.1"
		port = strings.TrimPrefix(raw, ":")
	}
	if strings.TrimSpace(port) == "" {
		return "", fmt.Errorf("%s requires a port", envPprofAddr)
	}
	host = strings.TrimSpace(host)
	if host == "" {
		host = "127.0.0.1"
	}
	if !pprofHostIsLoopback(host) {
		return "", fmt.Errorf("%s must bind to localhost or a loopback address", envPprofAddr)
	}
	return net.JoinHostPort(host, port), nil
}

func pprofHostIsLoopback(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}

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
