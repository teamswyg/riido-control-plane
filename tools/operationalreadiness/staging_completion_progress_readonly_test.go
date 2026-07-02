package main

import (
	"encoding/json"
	"os"
	"testing"
)

const stagingCompletionProgressReadOnlyEvidence = "docs/30-architecture/evidence/staging-completion-progress-readonly-2026-07-02.json"

func TestStagingCompletionProgressReadOnlyBaselineStaysPartial(t *testing.T) {
	body, err := os.ReadFile("../../" + stagingCompletionProgressReadOnlyEvidence)
	if err != nil {
		t.Fatal(err)
	}
	var evidence completionProgressBaselineEvidence
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	if !evidence.Redacted || evidence.Decision.Status != "partial" {
		t.Fatalf("baseline must be redacted partial evidence: redacted=%v status=%s", evidence.Redacted, evidence.Decision.Status)
	}
	assertNoActiveCompletionProgress(t, evidence)
}

type completionProgressBaselineEvidence struct {
	Redacted       bool `json:"redacted"`
	APIObservation struct {
		V3 struct {
			ActiveStreamPresent bool `json:"active_stream_present"`
			RunningThreads      int  `json:"running_threads"`
			QueuedMessages      int  `json:"queued_messages"`
		} `json:"v3"`
		V2 struct {
			ActiveStreamPresent bool `json:"active_stream_present"`
			RunningThreads      int  `json:"running_threads"`
			QueuedThreads       int  `json:"queued_threads"`
		} `json:"v2"`
		Subscription struct {
			ActiveThreadFilterCount int `json:"active_thread_filter_count"`
		} `json:"subscription"`
	} `json:"api_observation"`
	Decision struct {
		Status string `json:"status"`
	} `json:"decision"`
}

func assertNoActiveCompletionProgress(t *testing.T, evidence completionProgressBaselineEvidence) {
	t.Helper()
	if evidence.APIObservation.V3.ActiveStreamPresent || evidence.APIObservation.V2.ActiveStreamPresent {
		t.Fatal("read-only baseline must not expose active streams")
	}
	if evidence.APIObservation.V3.RunningThreads != 0 || evidence.APIObservation.V2.RunningThreads != 0 {
		t.Fatal("read-only baseline must not expose running threads")
	}
	if evidence.APIObservation.V3.QueuedMessages != 0 || evidence.APIObservation.V2.QueuedThreads != 0 {
		t.Fatal("read-only baseline must not expose queued client-visible state")
	}
	if evidence.APIObservation.Subscription.ActiveThreadFilterCount != 0 {
		t.Fatal("read-only baseline must not expose active thread filters")
	}
}
