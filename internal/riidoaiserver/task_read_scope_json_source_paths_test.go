package riidoaiserver

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestReadScopeJSONSourcesMatchOnlyKnownPathsAndCanonicalVersions(t *testing.T) {
	raw := json.RawMessage(`{ "id": "leaf", "body": "<p>한글</p>", "zero": 0 }`)
	canonical, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		field, kind string
		array       bool
	}{
		{"links", "link", true},
		{"github", "link", true},
		{"figma", "link", false},
		{"attachments", "attachment", true},
		{"items", "comment", true},
	} {
		t.Run(tc.field, func(t *testing.T) {
			prefix, suffix := "{\n \""+tc.field+"\" : ", "\n}"
			if tc.array {
				prefix, suffix = prefix+"[  ", "  ]"+suffix
			}
			version := assignmentContextVersion(string(canonical))
			if tc.kind == "comment" {
				version = assignmentContextVersion("<p>한글</p>")
			}
			source := AssignmentContextSource{Kind: tc.kind, ID: "leaf", Part: "metadata", Version: version}
			data := json.RawMessage(prefix + string(raw) + suffix)
			selectSources, err := prepareReadScopeJSONSources(data, []AssignmentContextSource{source})
			if err != nil {
				t.Fatal(err)
			}
			if got := selectSources(0, len(prefix)); len(got) != 0 {
				t.Fatal("envelope audited leaf")
			}
			if got := selectSources(len(prefix), len(prefix)+1); !reflect.DeepEqual(got, []AssignmentContextSource{source}) {
				t.Fatal("actual leaf byte was not audited")
			}
			if got := selectSources(len(prefix)+len(raw), len(data)); len(got) != 0 {
				t.Fatal("trailing envelope audited leaf")
			}
			source.Version = "wrong-version"
			if _, err = prepareReadScopeJSONSources(data, []AssignmentContextSource{source}); err == nil {
				t.Fatal("unknown source version accepted")
			}
		})
	}
}

func TestReadScopeJSONSourcesDoNotFindLeavesInNestedOrQuotedContent(t *testing.T) {
	raw := `{"id":"leaf"}`
	source := AssignmentContextSource{Kind: "link", ID: "leaf", Part: "metadata", Version: assignmentContextVersion(raw)}
	for _, data := range []string{`{"details":{"links":[` + raw + `]}}`, `{"text":"{\"links\":[{\"id\":\"leaf\"}]}"}`, `{"figma":null}`} {
		if _, err := prepareReadScopeJSONSources(json.RawMessage(data), []AssignmentContextSource{source}); err == nil {
			t.Fatal("unmatched leaf evidence accepted")
		}
	}
}
