package main

import (
	"sync"
	"sync/atomic"
	"time"
)

type sample struct {
	latency time.Duration
}

func run(cfg config) (pressureReport, error) {
	started := time.Now().UTC()
	report := pressureReport{SchemaVersion: evidenceSchema, StartedAt: started, Fixture: fixtureSummary{Threads: cfg.Threads, Lines: cfg.Lines}}
	for _, sc := range scenarios() {
		for _, concurrency := range cfg.Concurrencies {
			row, err := runOne(cfg, sc, concurrency)
			if err != nil {
				return pressureReport{}, err
			}
			report.Runs = append(report.Runs, row)
			report.Candidates = append(report.Candidates, row.Candidate)
		}
	}
	report.EndedAt = time.Now().UTC()
	report.DurationMs = report.EndedAt.Sub(started).Milliseconds()
	return report, nil
}

func runOne(cfg config, sc scenario, concurrency int) (pressureRun, error) {
	op, err := sc.build(cfg)
	if err != nil {
		return pressureRun{}, err
	}
	runtimeGC()
	before := sampleResources()
	samples, ops, errors := runWorkers(op, concurrency, cfg.Duration)
	after := sampleResources()
	return pressureRun{
		Scenario: sc.name, Concurrency: concurrency, Operations: ops, Errors: errors,
		OpsPerSec: float64(ops) / cfg.Duration.Seconds(), Latency: summarize(samples),
		LatencyUS: summarizeMicros(samples),
		Resources: diffResources(before, after, ops),
		Candidate: candidateEntry{Scenario: sc.name, Risk: sc.risk, Next: sc.next},
	}, nil
}

func runWorkers(op func() error, concurrency int, duration time.Duration) ([]time.Duration, int64, int64) {
	deadline := time.Now().Add(duration)
	out := make(chan sample, concurrency*64)
	var wg sync.WaitGroup
	var ops atomic.Int64
	var errors atomic.Int64
	for range concurrency {
		wg.Add(1)
		go worker(op, deadline, out, &ops, &errors, &wg)
	}
	go func() { wg.Wait(); close(out) }()
	var values []time.Duration
	for sample := range out {
		values = append(values, sample.latency)
	}
	return values, ops.Load(), errors.Load()
}
