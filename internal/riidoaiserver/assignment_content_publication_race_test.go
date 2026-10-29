package riidoaiserver

import "testing"

func TestAssignmentContentPublicationRacesLegacyCommentAtomically(t *testing.T) {
	store := newContentBundleDynamo(t, &contentBundleFixture{})
	snapshot, pages := assignmentContentTestBundle()
	start := make(chan struct{})
	results := make(chan error, 2)
	go func() {
		<-start
		results <- store.SaveAssignmentContentBundle(t.Context(), snapshot, pages)
	}()
	go func() {
		<-start
		results <- store.SaveAssignmentContextComments(t.Context(), "assignment", [][]AIAgentTaskContextComment{{{ID: "legacy"}}})
	}()
	close(start)
	first, second := <-results, <-results
	if (first == nil) == (second == nil) {
		t.Fatalf("exactly one publication must win: %v / %v", first, second)
	}
	_, scoped, err := store.LoadAssignmentContextSnapshot(t.Context(), "assignment")
	if err != nil {
		t.Fatal(err)
	}
	_, legacy, err := store.LoadAssignmentContextCommentPage(t.Context(), "assignment", 0)
	if err != nil || scoped == legacy {
		t.Fatalf("mixed physical publication: scoped=%v legacy=%v err=%v", scoped, legacy, err)
	}
}

func TestAssignmentContentRetentionEnabledFencesLegacyPublication(t *testing.T) {
	f := newContextRetentionFixture(t)
	snapshot, pages := assignmentContentTestBundle()
	if err := f.store.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err != nil {
		t.Fatal(err)
	}
	if err := f.store.SaveAssignmentContextSnapshot(t.Context(), AssignmentContextSnapshot{AssignmentID: "assignment"}); err == nil {
		t.Fatal("retention-enabled legacy snapshot replaced scoped bundle")
	}
	if err := f.store.SaveAssignmentContextComments(t.Context(), "assignment", [][]AIAgentTaskContextComment{{{ID: "legacy"}}}); err == nil {
		t.Fatal("retention-enabled legacy comments joined scoped bundle")
	}
}
