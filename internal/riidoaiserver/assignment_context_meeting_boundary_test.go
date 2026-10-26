package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestMeetingBodyFailuresNeverServeCapturedFallback(t *testing.T) {
	for _, tc := range []struct {
		status string
		code   int
	}{{"denied", 403}, {"deleted", 404}, {"unavailable", 503}, {"", 503}} {
		t.Run(tc.status, func(t *testing.T) {
			c := meetingContextFixture(t, tc.status)
			f := newAssignmentContextFixtureWith(t, NewMemoryAssignmentContextStore(), c)
			code, body := f.read("meeting", "meeting-1", f.leased.LeaseToken)
			if code != tc.code || strings.Contains(body, "MEETING_BODY_ONLY_MARKER") {
				t.Fatalf("status=%d", code)
			}
			code, _ = f.read("meeting", "not-related", f.leased.LeaseToken)
			if code != http.StatusForbidden {
				t.Fatal("outside snapshot meeting allowed")
			}
		})
	}
}

func TestMeetingBodiesFitSnapshotWithoutMutatingCapturedInput(t *testing.T) {
	c := meetingContextFixture(t, "available")
	original := strings.Repeat("한글본문", 30000)
	c.Extended.RelatedMeetings[0].Document.Content = original
	bounded, truncated := boundAssignmentContext(c)
	if !truncated || !assignmentContextFits(bounded) {
		t.Fatal("meeting body excluded from snapshot bound")
	}
	got := bounded.Extended.RelatedMeetings[0]
	if !got.Truncated || got.BodyStatus != "partial" || !utf8.ValidString(got.Document.Content) {
		t.Fatal("cut body is not explicitly partial UTF-8")
	}
	if c.Extended.RelatedMeetings[0].Document.Content != original {
		t.Fatal("snapshot bound mutated caller body")
	}
	content, alternate, cut := boundMeetingText(original, original)
	if !cut || len(content)+len(alternate) > taskContextMeetingBodyBytes || !utf8.ValidString(content+alternate) {
		t.Fatal("meeting capture byte limit failed")
	}
}

func TestMeetingBodyReadVisibilityRequiresCurrentBodyPermission(t *testing.T) {
	for _, status := range []string{"available", "empty", "partial", "denied", "deleted", "unavailable", ""} {
		current := meetingContextFixture(t, status)
		visible := visibleRunReadSources(current)
		want := status == "available" || status == "empty" || status == "partial"
		if visible[runReadSourceKey{"related_meeting", "meeting-1", "body"}] != want {
			t.Fatalf("body visibility status=%s", status)
		}
		current.Extended.RelatedStatus = "unavailable"
		if visibleRunReadSources(current)[runReadSourceKey{"related_meeting", "meeting-1", "body"}] {
			t.Fatal("failed relation section authorized a body")
		}
	}
}
