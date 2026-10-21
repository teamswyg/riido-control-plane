package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLatestProviderDocumentLegacyEmptyJSONStringPreservesHTMLProjectionEvidence(t *testing.T) {
	for _, legacy := range []any{nil, ""} {
		htmlReads := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/documents/providers/team-a/component-a" {
				if r.URL.Query().Get("format") == "json" {
					_ = json.NewEncoder(w).Encode(map[string]any{"tiptapJson": legacy})
					return
				}
				htmlReads++
				_ = json.NewEncoder(w).Encode(map[string]any{"html": "<p>Latest agent HTML-only edit</p>"})
				return
			}
			writeProviderDocumentContextFixture(t, w, r)
		}))
		got := readProviderDocumentContextFixture(t, server.URL)
		server.Close()
		if htmlReads != 1 || got.Document.AlternateContent != "Latest agent HTML-only edit" || got.Document.AlternateSource != "provider_html_projection" {
			t.Fatalf("HTML-only document lost: %+v reads=%d", got.Document, htmlReads)
		}
	}
}
