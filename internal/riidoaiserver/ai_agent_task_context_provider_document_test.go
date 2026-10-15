package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
)

func TestAIAgentPrivateTaskContextClientPrefersProviderDocumentHTML(t *testing.T) {
	gotPaths := []string{}
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		gotPaths = append(gotPaths, r.URL.RequestURI())
		mu.Unlock()
		if got := r.Header.Get("Authorization"); got != "Bearer user-jwt" {
			t.Fatalf("authorization header = %q", got)
		}
		writeProviderDocumentContextFixture(t, w, r)
	}))
	defer server.Close()

	contextSnapshot := readProviderDocumentContextFixture(t, server.URL)
	if contextSnapshot.Document.Content != "<p>fresh provider document says Bye World.</p>" ||
		contextSnapshot.Document.ContentFormat != "html" ||
		contextSnapshot.Document.TiptapDocumentID != "doc-a" {
		t.Fatalf("context snapshot document = %+v", contextSnapshot.Document)
	}
	wantPaths := []string{
		"/public/components/component-a/workspace",
		"/teams/team-a/components/component-a?getDocument=true",
		"/documents/providers/team-a/component-a?format=html",
		"/teams/team-a/components/component-a/files",
		"/teams/team-a/components/component-a/web-links",
	}
	sort.Strings(gotPaths)
	sort.Strings(wantPaths)
	if strings.Join(gotPaths, "\n") != strings.Join(wantPaths, "\n") {
		t.Fatalf("paths = %#v", gotPaths)
	}
}

func TestAIAgentPrivateTaskContextClientFallsBackWhenProviderDocumentUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/documents/providers/team-a/component-a" {
			http.NotFound(w, r)
			return
		}
		writeProviderDocumentContextFixture(t, w, r)
	}))
	defer server.Close()

	contextSnapshot := readProviderDocumentContextFixture(t, server.URL)
	if contextSnapshot.Document.Content != "<p>stale component document</p>" {
		t.Fatalf("context snapshot document = %+v", contextSnapshot.Document)
	}
}
