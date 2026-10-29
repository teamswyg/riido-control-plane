package riidoaiserver

import "testing"

// The prompt shows project, goal and parent keys with a "RIID-" prefix; a
// staging agent that reused the shown key for a goal was refused as out of
// scope. Shown, stored and ID forms must all resolve to the same reference.
func TestAssignmentContextTargetMatchesShownKeyForms(t *testing.T) {
	for _, target := range []string{"d7LcOdzu5RM5Djt9-tHFX", "0-14", "RIID-0-14", "riid-0-14"} {
		if !assignmentContextTargetMatches(target, "d7LcOdzu5RM5Djt9-tHFX", "0-14") {
			t.Fatalf("%q must match the goal shown as RIID-0-14", target)
		}
	}
	if !assignmentContextTargetMatches("RIID-1234", "id", "RIID-1234") {
		t.Fatal("an already prefixed key must still match itself")
	}
	for _, target := range []string{"RIID-0-13", "0-1", "RIID-", ""} {
		if assignmentContextTargetMatches(target, "d7LcOdzu5RM5Djt9-tHFX", "0-14") {
			t.Fatalf("%q must not match", target)
		}
	}
	if assignmentContextTargetMatches("RIID-", "id", "") {
		t.Fatal("an empty key must never match a bare prefix")
	}
}
