package riidoaiserver

import (
	"bytes"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestAssignmentContentCacheNegotiationAndAudit(t *testing.T) {
	f := newAssignmentContentHTTPFixture(t)
	req := contentCacheRequest(f)
	code, legacy := readContentCache(t, f, req, "", "", nil)
	if code != http.StatusOK || legacy.cacheFieldsPresent {
		t.Fatal("legacy response changed", code)
	}
	code, cold := readContentCache(t, f, req, "1", "", nil)
	if code != http.StatusOK || !regexp.MustCompile(`^sha256:[0-9a-f]{64}$`).MatchString(cold.ContentDigest) || cold.NotModified || !bytes.Equal(legacy.Data, cold.Data) {
		t.Fatal("cold read did not negotiate a complete representation", code)
	}
	req.RequestID = "cache-hit-audit"
	code, hit := readContentCache(t, f, req, "1", cold.ContentDigest, nil)
	if code != http.StatusOK || !hit.NotModified || hit.ContentDigest != cold.ContentDigest || string(hit.Data) != "null" || !reflect.DeepEqual(hit.Sources, cold.Sources) || hit.CapturedAt != cold.CapturedAt {
		t.Fatal("matching read did not return current metadata and null data", code)
	}
	reads, err := f.contexts.ListAssignmentContextReads(t.Context(), f.leased.ID)
	if err != nil {
		t.Fatal(err)
	}
	var audited []AssignmentContextRead
	for _, read := range reads {
		if read.RequestID == req.RequestID {
			audited = append(audited, read)
		}
	}
	if len(audited) != len(hit.Sources) || audited[0].SourceID != hit.Sources[0].ID || audited[0].SourcePart != hit.Sources[0].Part || audited[0].Status != AssignmentContextReadServed {
		t.Fatal("hit skipped the current page audit")
	}
	code, miss := readContentCache(t, f, req, "1", "sha256:"+strings.Repeat("0", 64), nil)
	if code != http.StatusOK || miss.NotModified || miss.ContentDigest != cold.ContentDigest || !bytes.Equal(miss.Data, cold.Data) {
		t.Fatal("digest miss did not return full representation", code)
	}
}

func TestAssignmentContentCacheHeaderCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name, version, known, view string
		extra                      func(*http.Request)
		status                     int
	}{
		{"legacy-known-ignored", "", "invalid", "content", nil, 200},
		{"unknown-version", "2", "invalid", "content", nil, 200},
		{"non-content", "1", "invalid", "base", nil, 200},
		{"live", "1", "invalid", "agents", nil, 200},
		{"bad-prefix", "1", strings.Repeat("a", 64), "content", nil, 400},
		{"uppercase", "1", "sha256:" + strings.Repeat("A", 64), "content", nil, 400},
		{"empty", "1", "", "content", func(r *http.Request) { r.Header.Set(contentCacheKnownHeader, "") }, 400},
		{"duplicate", "1", "sha256:" + strings.Repeat("a", 64), "content", func(r *http.Request) { r.Header.Add(contentCacheKnownHeader, "sha256:"+strings.Repeat("a", 64)) }, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newAssignmentContentHTTPFixture(t)
			req := contentCacheRequest(f)
			req.View = tc.view
			if req.View != "content" {
				req.Cursor, req.Target = "", ""
			}
			code, out := readContentCache(t, f, req, tc.version, tc.known, tc.extra)
			if code != tc.status || out.cacheFieldsPresent {
				t.Fatal("negotiation compatibility failed", code)
			}
		})
	}
}
