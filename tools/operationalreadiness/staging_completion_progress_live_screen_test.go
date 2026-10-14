package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

const liveCompletionScreenEvidencePath = "docs/30-architecture/evidence/staging-completion-progress-live-screen-2026-07-02.json"

func TestStagingCompletionProgressLiveScreenStaysPartial(t *testing.T) {
	evidence := loadLiveCompletionScreenEvidence(t)
	if !evidence.Redacted || evidence.Decision.Status != "partial" {
		t.Fatalf("live screen evidence must stay redacted partial")
	}
	if evidence.APIObservation.Status != "not_captured" {
		t.Fatalf("screen-only evidence must not pretend API was captured")
	}
	if evidence.Browser.VisibleThinking != 0 || evidence.Browser.VisibleRunning != 0 {
		t.Fatalf("live screen evidence must not show active progress")
	}
	if evidence.Browser.VisibleQueued != 0 || evidence.Browser.VisibleRequest400 != 0 {
		t.Fatalf("live screen evidence must not show queued/error copy")
	}
	assertEvidenceFileHash(t, evidence.Browser.Screenshot, evidence.Browser.SHA256)
}

func loadLiveCompletionScreenEvidence(t *testing.T) liveCompletionScreenEvidence {
	t.Helper()
	body, err := os.ReadFile("../../" + liveCompletionScreenEvidencePath)
	if err != nil {
		t.Fatal(err)
	}
	var evidence liveCompletionScreenEvidence
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	return evidence
}

func assertEvidenceFileHash(t *testing.T, path, want string) {
	t.Helper()
	body, err := os.ReadFile("../../" + path)
	if err != nil {
		t.Fatal(err)
	}
	got := sha256.Sum256(body)
	if hex.EncodeToString(got[:]) != want {
		t.Fatalf("screenshot hash drifted for %s", path)
	}
}

type liveCompletionScreenEvidence struct {
	Redacted       bool                             `json:"redacted"`
	Browser        liveCompletionBrowserObservation `json:"browser_observation"`
	APIObservation struct {
		Status string `json:"status"`
	} `json:"api_observation"`
	Decision struct {
		Status string `json:"status"`
	} `json:"decision"`
}

type liveCompletionBrowserObservation struct {
	Screenshot        string `json:"screenshot"`
	SHA256            string `json:"screenshot_sha256"`
	VisibleThinking   int    `json:"visible_thinking_count"`
	VisibleRunning    int    `json:"visible_running_count"`
	VisibleQueued     int    `json:"visible_queued_copy_count"`
	VisibleRequest400 int    `json:"visible_request_failed_400_count"`
}
