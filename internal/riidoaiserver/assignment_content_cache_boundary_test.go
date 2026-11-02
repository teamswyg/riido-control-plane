package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestAssignmentContentCacheBoundaries(t *testing.T) {
	for _, name := range []string{"device", "revoked-device", "lease", "expired-lease", "assignment", "terminal", "scope", "cursor", "page", "page-version", "audit", "quota", "quota-error"} {
		t.Run(name, func(t *testing.T) {
			f := newAssignmentContentHTTPFixture(t)
			req := contentCacheRequest(f)
			code, cold := readContentCache(t, f, req, "1", "", nil)
			if code != http.StatusOK || cold.ContentDigest == "" {
				t.Fatal("initial negotiated read failed", code)
			}
			want, decorate := mutateContentCacheBoundary(t, f, &req, name, cold.ContentDigest)
			code, out := readContentCache(t, f, req, "1", cold.ContentDigest, decorate)
			if code != want || out.NotModified || out.ContentDigest != "" || len(out.Data) != 0 {
				t.Fatal("candidate bypassed current read gate", name, code)
			}
		})
	}
}
