package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

func TestTaskContextBundlesReduceHTTPReadsWithoutCrossTaskReuse(t *testing.T) {
	for _, parents := range []bool{false, true} {
		t.Run(map[bool]string{false: "current", true: "parents"}[parents], func(t *testing.T) {
			f := newContextBundleFixture(t, parents)
			client, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: f.rest.URL, GraphQLURL: f.graph.URL})
			if err != nil {
				t.Fatal(err)
			}
			got, err := client.GetAIAgentTaskContextForRequest(context.Background(), AIAgentTaskContextRequest{ComponentID: "task", WorkspaceID: "workspace", BearerToken: "SYNTHETIC_VIEWER"})
			if err != nil || !strings.Contains(got.Document.Content, "STRUCTURED_BODY") {
				t.Fatal("structured-only body missing", err)
			}
			want := 4
			if parents {
				want = 13
			}
			f.mu.Lock()
			defer f.mu.Unlock()
			count := 0
			for path, n := range f.reads {
				count += n
				if n != 1 {
					t.Fatalf("repeated read %s", path)
				}
			}
			if count != want {
				t.Fatalf("HTTP calls=%d want=%d", count, want)
			}
		})
	}
}

func TestTaskContextBundlesNeverReplayDenialAgainstREST(t *testing.T) {
	for _, code := range []int{401, 403, 500} {
		t.Run(map[int]string{401: "401", 403: "403", 500: "500"}[code], func(t *testing.T) {
			f := newContextBundleFixture(t, false)
			f.bundleStatus = code
			client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: f.rest.URL, GraphQLURL: f.graph.URL})
			got, err := client.GetAIAgentTaskContextForRequest(context.Background(), AIAgentTaskContextRequest{ComponentID: "task", WorkspaceID: "workspace", BearerToken: "SYNTHETIC_VIEWER"})
			if err == nil || got.Document.Content != "" {
				t.Fatal("latest document failure must prevent stale execution", err)
			}
			f.mu.Lock()
			defer f.mu.Unlock()
			for path := range f.reads {
				if strings.Contains(path, "documents/providers") || strings.Contains(path, "web-links") {
					t.Fatal("fallback widened read surface")
				}
			}
		})
	}
}

func TestTaskContextBundleGraphQLDenialsHaveSafeTrace(t *testing.T) {
	f := newContextBundleFixture(t, false)
	f.graphqlError = "FORBIDDEN"
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: f.rest.URL, GraphQLURL: f.graph.URL})
	recorder := &recordingTraceRecorder{}
	ctx := WithTraceRecorder(context.Background(), recorder)
	_, err := client.GetAIAgentTaskContextForRequest(ctx, AIAgentTaskContextRequest{ComponentID: "task", WorkspaceID: "workspace", BearerToken: "SYNTHETIC_VIEWER"})
	if err == nil {
		t.Fatal("denied latest document must not execute")
	}
	spans := recorder.snapshot()
	if len(spans) != 4 {
		t.Fatalf("read spans=%d", len(spans))
	}
	for _, span := range spans {
		if span.Name == "task_context.workspace_location" || span.Name == "task_context.component_detail" {
			continue
		}
		if span.Attributes["riido.task_context.outcome"] != "denied" {
			t.Fatal("denial not visible")
		}
		for _, v := range span.Attributes {
			if strings.Contains(v, "CANARY") || strings.Contains(v, "VIEWER") || strings.Contains(v, "http") {
				t.Fatal("private data in trace")
			}
		}
	}
}
