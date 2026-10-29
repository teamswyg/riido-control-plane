package riidoaiserver

import "testing"

func TestAssignmentContentDynamoPublishesOneAtomicBundleAndRestarts(t *testing.T) {
	fixture := &contentBundleFixture{}
	store := newContentBundleDynamo(t, fixture)
	snapshot, pages := assignmentContentTestBundle()
	if err := store.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err != nil {
		t.Fatal(err)
	}
	if fixture.writes != 1 || len(fixture.items) != 2 {
		t.Fatalf("writes=%d items=%d", fixture.writes, len(fixture.items))
	}
	restarted := newContentBundleDynamo(t, fixture)
	got, found, err := restarted.LoadAssignmentContextSnapshot(t.Context(), "assignment")
	if err != nil || !found || len(got.Content) != 1 {
		t.Fatalf("manifest=%+v found=%v err=%v", got, found, err)
	}
	page, found, err := restarted.LoadAssignmentContentPage(t.Context(), "assignment", 1)
	if err != nil || !found || page.Cursor != "last" || !page.Complete {
		t.Fatalf("page=%+v found=%v err=%v", page, found, err)
	}
	if err := restarted.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err != nil {
		t.Fatalf("same bundle replay: %v", err)
	}
	if fixture.writes != 1 {
		t.Fatal("replay overwrote a captured bundle")
	}
}

func TestAssignmentContentDynamoRejectedPublicationLeavesNoPages(t *testing.T) {
	fixture := &contentBundleFixture{reject: true}
	store := newContentBundleDynamo(t, fixture)
	snapshot, pages := assignmentContentTestBundle()
	if err := store.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err == nil {
		t.Fatal("rejected transaction succeeded")
	}
	if len(fixture.items) != 0 {
		t.Fatal("partial bundle escaped a rejected transaction")
	}
	fixture.reject = false
	if err := store.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err != nil {
		t.Fatalf("same request after rejected transaction: %v", err)
	}
}
