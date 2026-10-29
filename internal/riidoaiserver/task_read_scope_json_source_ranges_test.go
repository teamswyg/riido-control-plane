package riidoaiserver

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestReadScopeJSONSourcesOnlyAuditIntersectingLeafVersions(t *testing.T) {
	a, b := `{"id":"same","url":"first"}`, `{"id":"same","url":"second"}`
	prefix := `{"details":{"links":[{"id":"same"}]},"note":"{\"id\":\"same\",\"url\":\"second\"}","links":[`
	// The quoted note and a nested links path are never top-level leaf evidence.
	data := json.RawMessage(prefix + a + "," + b + "]}")
	root := AssignmentContextSource{Kind: "task", ID: "root", Part: "metadata", Version: "root-version"}
	first := AssignmentContextSource{Kind: "link", ID: "same", Part: "metadata", Version: assignmentContextVersion(a)}
	second := AssignmentContextSource{Kind: "link", ID: "same", Part: "metadata", Version: assignmentContextVersion(b)}
	sources := []AssignmentContextSource{root, first, second}
	selectSources, err := prepareReadScopeJSONSources(data, sources)
	if err != nil {
		t.Fatal(err)
	}
	aStart, aEnd := len(prefix), len(prefix)+len(a)
	for name, tc := range map[string]struct {
		start, end int
		want       []AssignmentContextSource
	}{
		"envelope": {0, aStart, []AssignmentContextSource{root}},
		"first":    {aStart, aEnd, []AssignmentContextSource{root, first}},
		"second":   {aEnd + 1, aEnd + 1 + len(b), []AssignmentContextSource{root, second}},
		"crossing": {aEnd - 1, aEnd + 2, sources},
		"whole":    {0, len(data), sources},
		"empty":    {aStart, aStart, nil},
		"invalid":  {-1, len(data), nil},
	} {
		t.Run(name, func(t *testing.T) {
			if got := selectSources(tc.start, tc.end); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("sources = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestReadScopeJSONSourceRangesRejectMalformedEnvelopes(t *testing.T) {
	for _, data := range []string{`[]`, `{"links":null} {}`, `{"links":[null]}`, `{"items":[{"id":"comment"}]}`, `{"links":[],"links":[]}`, string([]byte{'{', '"', 'a', '"', ':', '"', 0xff, '"', '}'})} {
		if _, err := prepareReadScopeJSONSources(json.RawMessage(data), nil); err == nil {
			t.Errorf("accepted malformed source envelope")
		}
	}
}
