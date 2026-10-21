package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAbsentProjectionPreservesVerifiedSnapshotButValidEmptyJSONClearsIt(t *testing.T) {
	for _, structured := range []any{nil, `{"default":{"type":"doc","content":[]}}`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/documents/providers/team-a/component-a" {
				if r.URL.Query().Get("format") == "json" {
					writeTaskContextFixture(t, w, map[string]any{"tiptapJson": structured})
				} else {
					writeTaskContextFixture(t, w, map[string]any{"html": nil})
				}
				return
			}
			writeProviderDocumentContextFixture(t, w, r)
		}))
		snapshot := readProviderDocumentContextFixture(t, server.URL)
		server.Close()
		if structured == nil {
			if snapshot.Document.Source != "component_snapshot" || !strings.Contains(snapshot.Document.Content, "stale component document") {
				t.Fatalf("missing projection erased verified fallback: %+v", snapshot.Document)
			}
		} else if snapshot.Document.Content != "" || snapshot.Document.Status != "empty" {
			t.Fatalf("valid empty JSON resurrected old body: %+v", snapshot.Document)
		}
	}
}

func TestAbsentProjectionTransportIsNotReadFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeTaskContextFixture(t, w, map[string]any{"tiptapJson": nil, "html": nil})
	}))
	defer server.Close()
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL})
	document, err := client.getLatestProviderDocument(context.Background(), "team", "task", "fixture")
	if err != nil || document != nil {
		t.Fatalf("absence must remain distinct: %+v %v", document, err)
	}
}
