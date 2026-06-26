package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type config struct {
	Duration      time.Duration
	Concurrencies []int
	Threads       int
	Lines         int
	EvidenceOut   string
	PprofDir      string
	CandidateOut  string
}

func parseConfig(args []string) (config, error) {
	fs := flag.NewFlagSet("controlplanepressure", flag.ContinueOnError)
	cfg := config{}
	var concurrencyCSV string
	fs.DurationVar(&cfg.Duration, "duration", 500*time.Millisecond, "duration per scenario/concurrency")
	fs.StringVar(&concurrencyCSV, "concurrency", "1,8,32", "comma-separated worker counts")
	fs.IntVar(&cfg.Threads, "threads", 24, "fixture thread count")
	fs.IntVar(&cfg.Lines, "lines", 40, "fixture progress lines per thread")
	fs.StringVar(&cfg.EvidenceOut, "evidence-out", "", "optional evidence JSON path")
	fs.StringVar(&cfg.PprofDir, "pprof-dir", "", "optional directory for local CPU, heap, and goroutine profiles")
	fs.StringVar(&cfg.CandidateOut, "candidate-out", "", "optional closed-loop candidate JSON path")
	if err := fs.Parse(args); err != nil {
		return config{}, err
	}
	concurrencies, err := parseConcurrencies(concurrencyCSV)
	if err != nil {
		return config{}, err
	}
	cfg.Concurrencies = concurrencies
	if cfg.Duration <= 0 || cfg.Threads <= 0 || cfg.Lines <= 0 {
		return config{}, fmt.Errorf("duration, threads, and lines must be positive")
	}
	return cfg, nil
}

func parseConcurrencies(csv string) ([]int, error) {
	parts := strings.Split(csv, ",")
	values := make([]int, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || value <= 0 {
			return nil, fmt.Errorf("invalid concurrency %q", part)
		}
		values = append(values, value)
	}
	return values, nil
}
