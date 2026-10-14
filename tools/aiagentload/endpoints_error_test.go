package main

import "testing"

func TestEndpointsForScenarios(t *testing.T) {
	public, err := endpointsFor(config{Scenario: "public"})
	if err != nil || len(public) != 2 || public[0].Path != "/healthz" {
		t.Fatalf("public endpoints = %+v err=%v", public, err)
	}
	client, err := endpointsFor(config{Scenario: "client-read", WorkspaceID: " ws "})
	if err != nil {
		t.Fatal(err)
	}
	if len(client) != 7 || !client[2].Auth || client[2].Path != "/v2/client/workspaces/ws/ai-agent/bootstrap" {
		t.Fatalf("client endpoints = %+v", client)
	}
	if _, err := endpointsFor(config{Scenario: "weird"}); err == nil {
		t.Fatal("expected unknown scenario error")
	}
}

func TestErrorCategory(t *testing.T) {
	cases := map[string]string{
		"context canceled":             "context_cancelled",
		"request deadline exceeded":    "timeout",
		"Client.Timeout while waiting": "timeout",
		"read: connection reset":       "connection_closed",
		"broken pipe":                  "connection_closed",
		"lookup host: no such host":    "dns",
		"other failure":                "other",
	}
	for message, want := range cases {
		if got := errorCategory(message); got != want {
			t.Fatalf("errorCategory(%q) = %q, want %q", message, got, want)
		}
	}
}
