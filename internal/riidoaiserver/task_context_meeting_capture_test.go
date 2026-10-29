package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestMeetingCaptureBoundsCountConcurrencyAndDeadline(t *testing.T) {
	var active, peak, calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer requester" {
			t.Error("missing requester")
		}
		n := active.Add(1)
		defer active.Add(-1)
		for old := peak.Load(); n > old; old = peak.Load() {
			if peak.CompareAndSwap(old, n) {
				break
			}
		}
		calls.Add(1)
		time.Sleep(time.Millisecond * 3)
		writeJSON(w, 200, map[string]any{"id": path.Base(r.URL.Path), "description": "<p>LEGACY_MEETING_BODY</p>"})
	}))
	defer server.Close()
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL})
	var items []privateRelatedMeeting
	for i := range taskContextMaxMeetings + 5 {
		var item privateRelatedMeeting
		if json.Unmarshal([]byte(fmt.Sprintf(`{"meeting":{"id":"meeting-%d","title":"Meeting"}}`, i)), &item) != nil {
			t.Fatal("fixture")
		}
		items = append(items, item)
	}
	out := &AIAgentTaskContextExtended{}
	client.captureRelatedMeetings(t.Context(), "team", "requester", items, out)
	if len(out.RelatedMeetings) != taskContextMaxMeetings || out.MeetingTotal != len(items) || !out.MeetingTruncated || calls.Load() != taskContextMaxMeetings || peak.Load() > taskContextMeetingConcurrency {
		t.Fatal("meeting capture exceeded a bound")
	}
	if out.RelatedMeetings[0].Document == nil || out.RelatedMeetings[0].Document.Content != "LEGACY_MEETING_BODY" || out.RelatedMeetings[0].Document.ContentFormat != "markdown" {
		t.Fatal("legacy description was not normalized")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	cancelled := &AIAgentTaskContextExtended{}
	client.captureRelatedMeetings(ctx, "team", "requester", items, cancelled)
	if calls.Load() != taskContextMaxMeetings {
		t.Fatal("cancelled capture made another request")
	}
	for _, meeting := range cancelled.RelatedMeetings {
		if meeting.Document != nil {
			t.Fatal("cancelled capture returned a body")
		}
	}
}

func TestMeetingCaptureByteCutIsExplicit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"id": "meeting", "description": strings.Repeat("가나다", taskContextMeetingBodyBytes)})
	}))
	defer server.Close()
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL})
	got := client.readMeetingBody(t.Context(), "team", "requester", AIAgentTaskContextRelatedMeeting{ID: "meeting"})
	if got.Document == nil || !got.Truncated || got.BodyStatus != "partial" || len(got.Document.Content) > taskContextMeetingBodyBytes {
		t.Fatal("capture did not mark the byte cut")
	}
}
