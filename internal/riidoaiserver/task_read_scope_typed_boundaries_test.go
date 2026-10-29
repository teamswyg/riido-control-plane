package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestTaskReadScopeSelectorRejectsGraphQLIntegerOverflow(t *testing.T) {
	page, limit := int(9223372036854775807), 2
	if _, err := normalizeReadScopePartSelector(AssignmentContentComponent, "comments_page", &taskReadPartSelector{Page: &page, Limit: &limit}); err == nil {
		t.Fatal("overflowing signed GraphQL Int accepted")
	}
}

func TestTaskReadScopeRequiresSprintArray(t *testing.T) {
	rawScope, _ := readScopeWholeFixture()
	encoded, _ := json.Marshal(rawScope)
	var scope taskReadScope
	if json.Unmarshal(encoded, &scope) != nil {
		t.Fatal("fixture decode")
	}
	scope.Works[0].Sprints = nil
	if validateTaskReadScope(scope, scope.RootID, scope.TeamID) == nil {
		t.Error("missing required sprint list accepted")
	}
	work := taskReadScopeWork{ID: "work", Kind: AssignmentContentComponent}
	raw := []byte(`{"details":{"id":"work"},"links":[],"github":[],"repositories":[],"priorities":[],"checklists":[],"attachments":[]}`)
	if applyReadScopeWorkMetadata(&work, raw) == nil {
		t.Fatal("typed metadata lost required sprint list")
	}
}
