package main

import (
	"strings"
	"testing"
	"time"
)

func TestParseConfigNormalizesInputs(t *testing.T) {
	cfg, err := parseConfig([]string{
		"-base-url", " https://example.test/ ",
		"-token", " token ",
		"-workspace-id", "ws",
		"-scenario", " client-read ",
		"-duration", "2s",
		"-concurrency", "3",
		"-timeout", "4s",
		"-pprof-base-url", " http://127.0.0.1:6060/ ",
		"-pprof-profile-seconds", "2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != "https://example.test" || cfg.Token != "token" {
		t.Fatalf("normalized config = %+v", cfg)
	}
	if cfg.Duration != 2*time.Second || cfg.Concurrency != 3 || cfg.Timeout != 4*time.Second {
		t.Fatalf("timing config = %+v", cfg)
	}
	if cfg.PprofBaseURL != "http://127.0.0.1:6060" || cfg.PprofProfileSeconds != 2 {
		t.Fatalf("pprof config = %+v", cfg)
	}
}

func TestParseConfigRejectsInvalidInputs(t *testing.T) {
	cases := [][]string{
		{},
		{"-base-url", "https://example.test", "-duration", "0s"},
		{"-base-url", "https://example.test", "-concurrency", "0"},
		{"-base-url", "https://example.test", "-timeout", "0s"},
		{"-base-url", "https://example.test", "-pprof-profile-seconds", "-1"},
		{"-base-url", "https://example.test", "-scenario", "client-read"},
	}
	for _, args := range cases {
		if _, err := parseConfig(args); err == nil {
			t.Fatalf("parseConfig(%v) succeeded", args)
		}
	}
}

func TestParseConfigRejectsBadFlag(t *testing.T) {
	_, err := parseConfig([]string{"-unknown"})
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
		t.Fatalf("err = %v", err)
	}
}
