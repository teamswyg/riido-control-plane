package main

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"
)

func run(ctx context.Context, cfg config) (report, error) {
	eps, err := endpointsFor(cfg)
	if err != nil {
		return report{}, err
	}
	startedAt := time.Now().UTC()
	done := stopAfter(ctx, cfg.Duration)
	results := make(chan result, cfg.Concurrency*8)
	var wg sync.WaitGroup
	client := newHTTPClient(cfg)
	for workerID := range cfg.Concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runWorker(ctx, done, client, cfg, eps, workerID, results)
		}()
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	agg := newAggregator()
	for res := range results {
		agg.add(res)
	}
	endedAt := time.Now().UTC()
	return agg.report(cfg, baseHost(cfg.BaseURL), startedAt, endedAt), nil
}

func stopAfter(ctx context.Context, duration time.Duration) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		timer := time.NewTimer(duration)
		defer timer.Stop()
		select {
		case <-ctx.Done():
		case <-timer.C:
		}
		close(done)
	}()
	return done
}

func newHTTPClient(cfg config) *http.Client {
	return &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.Concurrency * 4,
			MaxIdleConnsPerHost: cfg.Concurrency * 4,
			MaxConnsPerHost:     cfg.Concurrency * 4,
			IdleConnTimeout:     30 * time.Second,
		},
	}
}

func baseHost(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Host
}
