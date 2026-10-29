package riidoaiserver

import (
	"strconv"
	"testing"
	"time"
)

func TestAssignmentContentTerminalRetainsEveryCapturedPageFor24Hours(t *testing.T) {
	f := newContextRetentionFixture(t)
	f.projection(t, "assignment", AssignmentRunning)
	snapshot, pages := assignmentContentTestBundle()
	if err := f.store.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err != nil {
		t.Fatal(err)
	}
	stored, found, err := f.store.LoadAssignmentContextSnapshot(t.Context(), "assignment")
	if err != nil || !found || len(stored.ContentBlocks) == 0 {
		t.Fatalf("block manifest missing found=%v err=%v", found, err)
	}
	record := AssignmentOperationRecord{AssignmentID: "assignment", Assignment: Assignment{ID: "assignment", State: AssignmentCompleted, UpdatedAt: f.now}, RecordedAt: f.now}
	credentials, _ := f.store.credentialsProvider.Credentials(t.Context())
	if err := f.store.saveAssignmentProjection(t.Context(), record, credentials); err != nil {
		t.Fatal(err)
	}
	want := strconv.FormatInt(f.now.Add(24*time.Hour).Unix(), 10)
	f.fake.mu.Lock()
	for i := range stored.ContentBlocks {
		item := f.fake.items[dynamoDBAssignmentContextPK("assignment")+"|"+assignmentContentSK(i)]
		if item[assignmentContextTTLAttribute]["N"] != want {
			f.fake.mu.Unlock()
			t.Fatalf("page %d missing terminal TTL", i)
		}
	}
	f.fake.mu.Unlock()
	f.now = f.now.Add(24 * time.Hour)
	for i := range pages {
		if _, found, err := f.store.LoadAssignmentContentPage(t.Context(), "assignment", i); err != nil || found {
			t.Fatalf("expired logical page %d found=%v err=%v", i, found, err)
		}
	}
}

func TestAssignmentContentLegacyWritesCannotReplacePublishedBundle(t *testing.T) {
	for _, operation := range []string{"snapshot", "comments"} {
		t.Run(operation, func(t *testing.T) {
			m := NewMemoryAssignmentContextStore()
			snapshot, pages := assignmentContentTestBundle()
			if err := m.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err != nil {
				t.Fatal(err)
			}
			var err error
			if operation == "snapshot" {
				err = m.SaveAssignmentContextSnapshot(t.Context(), AssignmentContextSnapshot{AssignmentID: "assignment"})
			} else {
				err = m.SaveAssignmentContextComments(t.Context(), "assignment", [][]AIAgentTaskContextComment{{{ID: "late"}}})
			}
			if err == nil {
				t.Fatal("legacy writer changed scoped bundle")
			}
		})
	}
}
