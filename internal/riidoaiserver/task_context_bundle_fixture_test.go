package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type contextBundleFixture struct {
	mu           sync.Mutex
	reads        map[string]int
	rest, graph  *httptest.Server
	bundleStatus int
	graphqlError string
}

func newContextBundleFixture(t *testing.T, parents bool) *contextBundleFixture {
	t.Helper()
	f := &contextBundleFixture{reads: map[string]int{}}
	f.rest = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.count("rest:" + r.URL.Path)
		if r.Header.Get("Authorization") != "Bearer SYNTHETIC_VIEWER" {
			t.Error("viewer credential lost")
		}
		if r.URL.Path == "/public/components/task/workspace" {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "task", "team": map[string]any{"id": "team", "workspace": map[string]any{"id": "workspace"}}})
			return
		}
		id := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
		out := map[string]any{"id": id, "componentType": "task", "title": "synthetic", "document": map[string]any{"HTMLContent": "<p>COMPONENT_SNAPSHOT</p>"}}
		if id == "task" && parents {
			for _, key := range []string{"project", "milestone", "task"} {
				parent := key
				if key == "task" {
					parent = "parent"
				}
				out[key] = map[string]any{"id": parent}
			}
		}
		_ = json.NewEncoder(w).Encode(out)
	}))
	f.graph = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer SYNTHETIC_VIEWER" {
			t.Error("cross-principal bearer")
		}
		var in struct {
			Query     string            `json:"query"`
			Variables map[string]string `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		kind := "projection"
		if strings.Contains(in.Query, "AgentTaskResources") {
			kind = "resources"
		}
		f.count(kind + ":" + in.Variables["id"])
		if f.bundleStatus != 0 {
			http.Error(w, "PRIVATE_UPSTREAM_CANARY", f.bundleStatus)
			return
		}
		if f.graphqlError != "" {
			_ = json.NewEncoder(w).Encode(map[string]any{"errors": []any{map[string]any{"message": "PRIVATE_UPSTREAM_CANARY", "extensions": map[string]string{"code": f.graphqlError}}}})
			return
		}
		data := map[string]any{"files": []any{}, "links": []any{}}
		if kind == "projection" {
			data = map[string]any{"structured": map[string]any{"tiptapJson": `{"default":{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"STRUCTURED_BODY"}]}]}}`}, "rendered": map[string]any{"html": nil}}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
	}))
	t.Cleanup(f.rest.Close)
	t.Cleanup(f.graph.Close)
	return f
}
func (f *contextBundleFixture) count(path string) { f.mu.Lock(); defer f.mu.Unlock(); f.reads[path]++ }
