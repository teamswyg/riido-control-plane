package riidoaiserver

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
)

func TestMeetingCaptureUsesRequesterPermissionAndPreservesEmpty(t *testing.T) {
	cases := []meetingReadCase{
		{name: "missing-projection", status: "unavailable", providerMissing: true},
		{name: "body", status: "available"},
		{name: "empty", status: "empty", empty: true},
		{name: "deleted", status: "deleted", deleted: true},
		{name: "denied", status: "denied", denied: true},
		{name: "provider-denied", status: "denied", providerDenied: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			upstream := meetingReadUpstream(t, tc, &calls)
			defer upstream.Close()
			client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL, GraphQLURL: upstream.URL})
			got := client.readMeetingBody(context.Background(), "team", "synthetic-requester", AIAgentTaskContextRelatedMeeting{ID: "meeting-1"})
			if got.BodyStatus != tc.status {
				t.Fatalf("status=%s", got.BodyStatus)
			}
			if tc.deleted || tc.denied {
				if calls.Load() != 1 || got.Document != nil {
					t.Fatal("inaccessible meeting fetched or kept a body")
				}
				return
			}
			if tc.providerDenied {
				if got.Document != nil {
					t.Fatal("provider denial fell back to stored body")
				}
				return
			}
			if got.Document == nil {
				t.Fatal("body missing")
			}
			if tc.providerMissing {
				if got.Document.Status != "unavailable" {
					t.Fatal("missing provider projection became authoritative empty")
				}
				return
			}
			if tc.empty {
				if got.Document.Content != "" || got.Document.AlternateContent != "" {
					t.Fatal("empty provider body resurrected fallback")
				}
				return
			}
			if !strings.Contains(got.Document.Content, "REQUESTER_MEETING_BODY") {
				t.Fatal("provider body missing")
			}
		})
	}
}
