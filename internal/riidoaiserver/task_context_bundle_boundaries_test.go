package riidoaiserver

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTaskContextBundleDeadlineKeepsSnapshotAndCancelsReads(t *testing.T) {
	f := newContextBundleFixture(t, false)
	graph := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		select {
		case <-r.Context().Done():
		case <-time.After(250 * time.Millisecond):
		}
	}))
	defer graph.Close()
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: f.rest.URL, GraphQLURL: graph.URL, Timeout: 40 * time.Millisecond})
	recorder := &recordingTraceRecorder{}
	ctx := WithTraceRecorder(context.Background(), recorder)
	started := time.Now()
	got, err := client.GetAIAgentTaskContextForRequest(ctx, AIAgentTaskContextRequest{ComponentID: "task", WorkspaceID: "workspace", BearerToken: "SYNTHETIC_VIEWER"})
	if err != nil || time.Since(started) > 500*time.Millisecond || !strings.Contains(got.Document.Content, "COMPONENT_SNAPSHOT") {
		t.Fatal("budget discarded snapshot or failed to cancel", err)
	}
	for _, span := range recorder.snapshot() {
		if span.Name == "task_context.workspace_location" || span.Name == "task_context.component_detail" {
			continue
		}
		if span.Attributes["riido.task_context.outcome"] != "budget_exhausted" {
			t.Fatal("timeout hidden by overall success")
		}
	}
}

func TestTaskContextBundleInvalidResponsesRemainUnavailable(t *testing.T) {
	for _, body := range []string{`null`, `{"data":null}`, `{"data":{}} trailing`, `{"data":{"structured":{"tiptapJson":42}}}`, strings.Repeat(" ", (1<<20)+1) + `{"data":{}}`} {
		t.Run("bounded response", func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(body)) }))
			defer srv.Close()
			client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: srv.URL, GraphQLURL: srv.URL})
			if doc := client.readProviderProjection(context.Background(), "team", "task", "SYNTHETIC_VIEWER"); doc != nil {
				t.Fatal("invalid data became a document")
			}
		})
	}
}

func TestTaskContextProjectionDoesNotTreatMissingOrMalformedAsEmpty(t *testing.T) {
	for _, p := range []privateDocumentProjections{{}, {JSON: taskContextTestString(`null`)}, {JSON: taskContextTestString(`{"type":"doc","content":[null]}`)}} {
		if p.document() != nil {
			t.Fatal("missing/invalid projection became empty")
		}
	}
}
