package riidoaiserver

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestReadScopeJSONSourcesKeepAllRealOccurrencesAndCallerOrder(t *testing.T) {
	raw := `{"id":"same","url":"https://fixture.test"}`
	prefix, middle := `{"links":[`, `],"github":[`
	data := json.RawMessage(prefix + raw + middle + raw + `]}`)
	leaf := AssignmentContextSource{Kind: "link", ID: "same", Part: "metadata", Version: assignmentContextVersion(raw)}
	root := AssignmentContextSource{Kind: "related_task", ID: "root", Part: "metadata", Version: "root-version"}
	want := []AssignmentContextSource{leaf, root, leaf}
	sources := append([]AssignmentContextSource(nil), want...)
	selectSources, err := prepareReadScopeJSONSources(data, sources)
	if err != nil {
		t.Fatal(err)
	}
	sources[0].ID = "changed-after-prepare"
	for _, start := range []int{len(prefix), len(prefix) + len(raw) + len(middle)} {
		if got := selectSources(start, start+len(raw)); !reflect.DeepEqual(got, want) {
			t.Fatal("source order, duplicate policy or actual occurrence changed")
		}
	}
	got := selectSources(0, len(data))
	got[0].ID = "changed-result"
	if !reflect.DeepEqual(selectSources(0, len(data)), want) {
		t.Fatal("returned slice mutated prepared sources")
	}
}
