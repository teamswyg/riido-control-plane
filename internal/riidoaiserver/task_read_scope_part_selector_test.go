package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
)

func TestTaskReadScopePartSelectorAndVersionPreserveFixedIdentity(t *testing.T) {
	parent := "comment"
	for _, tc := range []struct {
		kind       AssignmentContentKind
		part, want string
		input      *taskReadPartSelector
	}{
		{AssignmentContentComponent, "metadata", "{}", nil},
		{AssignmentContentBacklog, "comments_page", `{"page":1,"limit":20}`, nil},
		{AssignmentContentComponent, "comments_page", `{"parentCommentId":"comment","page":1,"limit":20}`, &taskReadPartSelector{ParentCommentID: &parent}},
		{AssignmentContentMeeting, "board_page", `{"limit":100}`, nil},
	} {
		selector, err := normalizeReadScopePartSelector(tc.kind, tc.part, tc.input)
		encoded, _ := json.Marshal(selector)
		if err != nil || string(encoded) != tc.want {
			t.Fatalf("normalized selector=%s err=%v", encoded, err)
		}
		canonical, _ := json.Marshal([]string{"riido-read-part.v1", string(tc.kind), "resource", "team", tc.part, tc.want, "riido_" + tc.part, "json", `{"full":"value"}`})
		digest := sha256.Sum256(canonical)
		if readScopePartVersion(tc.kind, "resource", "team", tc.part, "riido_"+tc.part, `{"full":"value"}`, selector) != "sha256:"+hex.EncodeToString(digest[:]) {
			t.Fatal("part version differs from independent canonical wire hash")
		}
	}
}

func TestTaskReadScopePartSelectorRejectsKindAndFieldConfusion(t *testing.T) {
	zero, large, parent, cursor := 0, 201, "comment", "cursor"
	for _, tc := range []struct {
		kind     AssignmentContentKind
		part     string
		selector *taskReadPartSelector
	}{
		{AssignmentContentMeeting, "comments_page", nil},
		{AssignmentContentComponent, "board_page", nil},
		{AssignmentContentComponent, "metadata", &taskReadPartSelector{Page: &zero}},
		{AssignmentContentComponent, "comments_page", &taskReadPartSelector{Page: &zero}},
		{AssignmentContentComponent, "comments_page", &taskReadPartSelector{BoardCursor: &cursor}},
		{AssignmentContentMeeting, "board_page", &taskReadPartSelector{ParentCommentID: &parent}},
		{AssignmentContentMeeting, "board_page", &taskReadPartSelector{Limit: &large}},
		{AssignmentContentComponent, "unrecognized", nil},
	} {
		if _, err := normalizeReadScopePartSelector(tc.kind, tc.part, tc.selector); err == nil {
			t.Fatalf("invalid %s/%s selector accepted", tc.kind, tc.part)
		}
	}
}
