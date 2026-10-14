package main

import (
	"encoding/json"
	"os"
	"testing"
)

const stagingClientP0VisualEvidence = "docs/30-architecture/evidence/staging-client-p0-visual-retest-2026-07-02.json"

func TestStagingClientP0VisualRetestScreenshotsStayBound(t *testing.T) {
	evidence := loadStagingClientVisualEvidence(t)
	if evidence.Decision == "" || len(evidence.Screenshots) != 2 {
		t.Fatalf("unexpected visual evidence shape: decision=%q screenshots=%d", evidence.Decision, len(evidence.Screenshots))
	}
	for _, screenshot := range evidence.Screenshots {
		assertEvidenceFileHash(t, screenshot.Path, screenshot.SHA256)
		if screenshot.Bytes <= 0 {
			t.Fatalf("screenshot %s has no byte count", screenshot.ID)
		}
		assertNoStaleQueuedOrThinking(t, screenshot)
	}
}

func loadStagingClientVisualEvidence(t *testing.T) stagingClientVisualEvidence {
	t.Helper()
	body, err := os.ReadFile("../../" + stagingClientP0VisualEvidence)
	if err != nil {
		t.Fatal(err)
	}
	var evidence stagingClientVisualEvidence
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	return evidence
}

func assertNoStaleQueuedOrThinking(t *testing.T, screenshot stagingClientVisualScreenshot) {
	t.Helper()
	if screenshot.VisualCounts.QueuedCopy != 0 || screenshot.VisualCounts.Thinking != 0 {
		t.Fatalf("screenshot %s shows queued/thinking drift: %+v", screenshot.ID, screenshot.VisualCounts)
	}
	if screenshot.APISummary.V3.HasActiveStream || screenshot.APISummary.V3.QueuedMessages != 0 {
		t.Fatalf("screenshot %s v3 summary is not closed: %+v", screenshot.ID, screenshot.APISummary.V3)
	}
	if screenshot.APISummary.V2.ActiveThreadFilters != 0 {
		t.Fatalf("screenshot %s v2 subscription has active filters: %+v", screenshot.ID, screenshot.APISummary.V2)
	}
}

type stagingClientVisualEvidence struct {
	Decision    string                          `json:"decision"`
	Screenshots []stagingClientVisualScreenshot `json:"screenshots"`
}

type stagingClientVisualScreenshot struct {
	ID           string `json:"id"`
	Path         string `json:"path"`
	SHA256       string `json:"sha256"`
	Bytes        int    `json:"bytes"`
	VisualCounts struct {
		QueuedCopy int `json:"queued_copy"`
		Thinking   int `json:"thinking"`
	} `json:"visual_counts"`
	APISummary struct {
		V2 struct {
			ActiveThreadFilters int `json:"active_thread_filters"`
		} `json:"v2_stream_subscription"`
		V3 struct {
			HasActiveStream bool `json:"has_active_stream"`
			QueuedMessages  int  `json:"queued_messages"`
		} `json:"v3_threads"`
	} `json:"api_summary"`
}
