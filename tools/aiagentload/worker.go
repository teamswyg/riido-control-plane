package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

func runWorker(
	ctx context.Context,
	done <-chan struct{},
	client *http.Client,
	cfg config,
	eps []endpoint,
	workerID int,
	results chan<- result,
) {
	next := workerID % len(eps)
	for {
		select {
		case <-done:
			return
		default:
		}
		ep := eps[next]
		next = (next + 1) % len(eps)
		results <- callEndpoint(ctx, client, cfg, ep)
	}
}

func callEndpoint(ctx context.Context, client *http.Client, cfg config, ep endpoint) result {
	start := time.Now()
	reqCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, ep.Method, cfg.BaseURL+ep.Path, strings.NewReader(ep.Body))
	if err != nil {
		return result{Endpoint: ep.Path, Latency: time.Since(start), Error: err.Error()}
	}
	req.Header.Set("Accept", "application/json")
	if ep.Body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if ep.Auth {
		req.Header.Set("X-Riido-Ai-Agent-Token", cfg.Token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return result{Endpoint: ep.Path, Latency: time.Since(start), Error: err.Error()}
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return result{Endpoint: ep.Path, Status: resp.StatusCode, Latency: time.Since(start)}
}
