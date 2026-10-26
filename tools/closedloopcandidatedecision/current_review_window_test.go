package main

import (
	"testing"
	"time"
)

// A reviewer may choose an earlier follow-up, but may never extend the
// maximum seven-day review window or claim a future review.
func validPendingReviewWindow(reviewed, due, today time.Time, interval int) bool {
	return interval >= 1 && interval <= 7 && !reviewed.After(today) &&
		!due.Before(today) && due.Equal(reviewed.AddDate(0, 0, interval))
}

func TestPendingReviewWindowAllowsEarlierFollowUpWithoutExtendingExpiry(t *testing.T) {
	today := time.Date(2026, time.October, 26, 0, 0, 0, 0, time.UTC)
	for _, check := range []struct {
		name        string
		reviewedDay int
		dueDay      int
		interval    int
		valid       bool
	}{
		{"one day", 0, 1, 1, true},
		{"user selected October 30", 0, 4, 4, true},
		{"seven-day maximum", 0, 7, 7, true},
		{"valid through deadline", -4, 0, 4, true},
		{"zero interval", 0, 0, 0, false},
		{"negative interval", 0, -1, -1, false},
		{"beyond seven days", 0, 8, 8, false},
		{"declared interval mismatch", 0, 7, 4, false},
		{"future review", 1, 5, 4, false},
		{"expired deadline", -5, -1, 4, false},
	} {
		t.Run(check.name, func(t *testing.T) {
			got := validPendingReviewWindow(today.AddDate(0, 0, check.reviewedDay),
				today.AddDate(0, 0, check.dueDay), today, check.interval)
			if got != check.valid {
				t.Fatalf("valid review window = %v, want %v", got, check.valid)
			}
		})
	}
}
