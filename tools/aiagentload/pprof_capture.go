package main

import (
	"context"
	"net/http"
	"slices"
	"sync"
	"time"
)

func startPprofCapture(ctx context.Context, cfg config) func() pprofEvidence {
	evidence := pprofEvidence{
		Enabled:        cfg.PprofBaseURL != "",
		BaseHost:       baseHost(cfg.PprofBaseURL),
		ProfileSeconds: cfg.PprofProfileSeconds,
	}
	if cfg.PprofBaseURL == "" {
		return func() pprofEvidence { return evidence }
	}
	client := &http.Client{Timeout: cfg.Timeout + time.Duration(cfg.PprofProfileSeconds)*time.Second}
	samples := make(chan pprofSample, len(pprofTargets(cfg)))
	var wg sync.WaitGroup
	for _, target := range pprofTargets(cfg) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			samples <- fetchPprofSample(ctx, client, cfg.PprofBaseURL, target)
		}()
	}
	return func() pprofEvidence {
		wg.Wait()
		close(samples)
		for sample := range samples {
			evidence.Samples = append(evidence.Samples, sample)
		}
		slices.SortFunc(evidence.Samples, comparePprofSample)
		return evidence
	}
}

func comparePprofSample(a, b pprofSample) int {
	if a.Name < b.Name {
		return -1
	}
	if a.Name > b.Name {
		return 1
	}
	return 0
}
