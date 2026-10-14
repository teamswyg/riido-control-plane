package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

func parseConfig(args []string) (config, error) {
	fs := flag.NewFlagSet("aiagentload", flag.ContinueOnError)
	cfg := config{}
	fs.StringVar(&cfg.BaseURL, "base-url", "", "AI Agent base URL")
	fs.StringVar(&cfg.Token, "token", "", "AI Agent token")
	fs.StringVar(&cfg.WorkspaceID, "workspace-id", "workspace-smoke", "workspace id")
	fs.StringVar(&cfg.Scenario, "scenario", "client-read", "public or client-read")
	fs.DurationVar(&cfg.Duration, "duration", 30*time.Second, "load duration")
	fs.IntVar(&cfg.Concurrency, "concurrency", 16, "concurrent workers")
	fs.DurationVar(&cfg.Timeout, "timeout", 10*time.Second, "request timeout")
	fs.StringVar(&cfg.OutputPath, "out", "", "optional report JSON path")
	fs.StringVar(&cfg.OutputPath, "evidence-out", "", "optional report JSON path")
	fs.StringVar(&cfg.PprofBaseURL, "pprof-base-url", "", "optional loopback pprof base URL")
	fs.IntVar(&cfg.PprofProfileSeconds, "pprof-profile-seconds", 1, "CPU profile seconds when pprof is enabled")
	if err := fs.Parse(args); err != nil {
		return config{}, err
	}
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	cfg.Token = strings.TrimSpace(cfg.Token)
	cfg.Scenario = strings.TrimSpace(cfg.Scenario)
	cfg.PprofBaseURL = strings.TrimRight(strings.TrimSpace(cfg.PprofBaseURL), "/")
	if cfg.BaseURL == "" {
		return config{}, fmt.Errorf("base-url is required")
	}
	if cfg.Duration <= 0 || cfg.Concurrency <= 0 || cfg.Timeout <= 0 {
		return config{}, fmt.Errorf("duration, concurrency, and timeout must be positive")
	}
	if cfg.PprofProfileSeconds < 0 {
		return config{}, fmt.Errorf("pprof-profile-seconds must not be negative")
	}
	if cfg.Scenario == "client-read" && cfg.Token == "" {
		return config{}, fmt.Errorf("token is required for client-read scenario")
	}
	return cfg, nil
}
