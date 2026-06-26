package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHarnessPromotionRejectsExpiredSummary(t *testing.T) {
	t.Setenv("RIIDO_HARNESS_PROMOTION_NOW", "2026-06-24T02:00:00Z")
	root := t.TempDir()
	summaryPath := filepath.Join(root, "summary.json")
	outPath := filepath.Join(root, "candidates.json")
	m := manifest{Sources: []promotionSource{{ID: "smoke"}}}
	if err := writeJSON(summaryPath, liveSummary{
		ID:          "smoke",
		LiveStatus:  "failure",
		GeneratedAt: "2026-06-24T00:00:00Z",
		ExpiresAt:   "2026-06-24T01:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
	err := promoteSummary(root, m, summaryPath, outPath)
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expired summary failure, got %v", err)
	}
}

func TestHarnessPromotionRejectsSummaryAtExpiryBoundary(t *testing.T) {
	err := verifySummaryFresh(liveSummary{
		ID:          "smoke",
		GeneratedAt: "2026-06-24T00:00:00Z",
		ExpiresAt:   "2026-06-24T01:00:00Z",
	}, mustParseTimeForTest(t, "2026-06-24T01:00:00Z"))
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected exact-expiry summary failure, got %v", err)
	}
}

func TestHarnessPromotionRejectsZeroLengthFreshnessWindow(t *testing.T) {
	err := verifySummaryFresh(liveSummary{
		ID:          "smoke",
		GeneratedAt: "2026-06-24T00:00:00Z",
		ExpiresAt:   "2026-06-24T00:00:00Z",
	}, mustParseTimeForTest(t, "2026-06-24T00:00:00Z"))
	if err == nil || !strings.Contains(err.Error(), "expires") {
		t.Fatalf("expected zero-length freshness failure, got %v", err)
	}
}

func TestHarnessPromotionRejectsMissingExpiry(t *testing.T) {
	err := verifySummaryFresh(liveSummary{
		ID:          "smoke",
		GeneratedAt: "2026-06-24T00:00:00Z",
	}, mustParseTimeForTest(t, "2026-06-24T00:30:00Z"))
	if err == nil || !strings.Contains(err.Error(), "expires_at") {
		t.Fatalf("expected missing expires_at failure, got %v", err)
	}
}

func mustParseTimeForTest(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
