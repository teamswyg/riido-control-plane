package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLatestProviderDocumentJSONIsAuthoritative(t *testing.T) {
	for _, document := range []string{`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Fresh before history save"}]}]}`, `{"type":"doc","content":[]}`, `"{\"default\":{\"type\":\"doc\",\"content\":[{\"type\":\"paragraph\",\"content\":[{\"type\":\"text\",\"text\":\"Fresh before history save\"}]}]}}"`, `"{\"default\":{\"type\":\"doc\",\"content\":[]}}"`} {
		t.Run(document, func(t *testing.T) {
			htmlReads := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/documents/providers/team-a/component-a" {
					if r.URL.Query().Get("format") != "json" {
						htmlReads++
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"tiptapJson": json.RawMessage(document), "html": "stale provider HTML"})
					return
				}
				writeProviderDocumentContextFixture(t, w, r)
			}))
			defer server.Close()
			got := readProviderDocumentContextFixture(t, server.URL)
			if htmlReads != 0 || got.Document.Source != "provider_document_json" || strings.Contains(got.Document.Content, "stale") {
				t.Fatalf("document = %+v, html reads=%d", got.Document, htmlReads)
			}
			if strings.Contains(document, "Fresh") && !strings.Contains(got.Document.Content, "Fresh before history save") {
				t.Fatal("latest JSON omitted")
			}
			if !strings.Contains(document, "Fresh") && got.Document.Content != "" {
				t.Fatal("valid empty doc resurrected history")
			}
		})
	}
}

func TestLatestProviderDocumentErrorsDoNotFallBackToStaleHTML(t *testing.T) {
	for _, body := range []string{`{"tiptapJson":{"type":"wrong"}}`, `{"tiptapJson":"invalid"}`, `invalid`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/documents/providers/team-a/component-a" {
				_, _ = w.Write([]byte(body))
				return
			}
			writeProviderDocumentContextFixture(t, w, r)
		}))
		client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL})
		_, err := client.GetAIAgentTaskContextForRequest(context.Background(), AIAgentTaskContextRequest{ComponentID: "component-a", WorkspaceID: "workspace-a", BearerToken: "user-jwt"})
		server.Close()
		if err == nil {
			t.Fatalf("accepted unreadable latest document: %s", body)
		}
	}
}
