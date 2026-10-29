package riidoaiserver

import "testing"

func TestAssignmentContentDynamoLegacyWritersCannotReplaceBundle(t *testing.T) {
	for _, operation := range []string{"snapshot", "comments"} {
		t.Run(operation, func(t *testing.T) {
			store := newContentBundleDynamo(t, &contentBundleFixture{})
			snapshot, pages := assignmentContentTestBundle()
			if err := store.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err != nil {
				t.Fatal(err)
			}
			var err error
			if operation == "snapshot" {
				err = store.SaveAssignmentContextSnapshot(t.Context(), AssignmentContextSnapshot{AssignmentID: "assignment"})
			} else {
				err = store.SaveAssignmentContextComments(t.Context(), "assignment", [][]AIAgentTaskContextComment{{{ID: "late"}}})
			}
			if err == nil {
				t.Fatal("legacy writer replaced scoped publication")
			}
		})
	}
}

func TestAssignmentContentDynamoCannotPublishOverLegacyOrphanPayload(t *testing.T) {
	store := newContentBundleDynamo(t, &contentBundleFixture{})
	if err := store.SaveAssignmentContextComments(t.Context(), "assignment", [][]AIAgentTaskContextComment{{{ID: "orphan"}}}); err != nil {
		t.Fatal(err)
	}
	snapshot, pages := assignmentContentTestBundle()
	if err := store.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err == nil {
		t.Fatal("scoped capture published over an incomplete legacy capture")
	}
	if _, found, err := store.LoadAssignmentContextSnapshot(t.Context(), "assignment"); err != nil || found {
		t.Fatalf("failed publication left snapshot: found=%v err=%v", found, err)
	}
}
