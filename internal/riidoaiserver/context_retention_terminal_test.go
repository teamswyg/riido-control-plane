package riidoaiserver

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"
)

func TestContextRetentionTerminalCommitKeepsInitialExpiryAndAudit(t *testing.T) {
	f := newContextRetentionFixture(t)
	ctx := context.Background()
	f.projection(t, "assignment", AssignmentRunning)
	if err := f.store.SaveAssignmentContextSnapshot(ctx, AssignmentContextSnapshot{AssignmentID: "assignment"}); err != nil {
		t.Fatal(err)
	}
	if err := f.store.SaveAssignmentContextComments(ctx, "assignment", [][]AIAgentTaskContextComment{{{ID: "comment"}}}); err != nil {
		t.Fatal(err)
	}
	if err := f.store.AppendAssignmentContextReads(ctx, []AssignmentContextRead{{AssignmentID: "assignment", ServedAt: f.now}}); err != nil {
		t.Fatal(err)
	}
	record := AssignmentOperationRecord{AssignmentID: "assignment", Assignment: Assignment{ID: "assignment", State: AssignmentCompleted, UpdatedAt: f.now}, RecordedAt: f.now}
	credentials, _ := f.store.credentialsProvider.Credentials(ctx)
	if err := f.store.saveAssignmentProjection(ctx, record, credentials); err != nil {
		t.Fatal(err)
	}
	want := strconv.FormatInt(f.now.Add(24*time.Hour).Unix(), 10)
	record.Assignment.UpdatedAt = f.now.Add(time.Hour)
	if err := f.store.saveAssignmentProjection(ctx, record, credentials); err != nil {
		t.Fatal(err)
	}
	if err := f.store.SaveAssignmentContextSnapshot(ctx, AssignmentContextSnapshot{AssignmentID: "assignment"}); err != nil {
		t.Fatal(err)
	}
	f.fake.mu.Lock()
	for key, item := range f.fake.items {
		if item["sk"]["S"] == "SNAPSHOT" || item["sk"]["S"] == "COMMENTS#0000" {
			if item[assignmentContextTTLAttribute]["N"] != want {
				t.Errorf("%s expiry=%v", key, item[assignmentContextTTLAttribute])
			}
		}
		if item["schema"]["S"] == "assignment_context_read/v1" && item[assignmentContextTTLAttribute] != nil {
			t.Errorf("audit TTL added")
		}
	}
	f.fake.mu.Unlock()
	f.now = f.now.Add(24 * time.Hour)
	if _, found, err := f.store.LoadAssignmentContextSnapshot(ctx, "assignment"); err != nil || found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	if err := f.store.SaveAssignmentContextSnapshot(ctx, AssignmentContextSnapshot{AssignmentID: "assignment"}); !errors.Is(err, errAssignmentContextExpired) {
		t.Fatalf("late recapture=%v", err)
	}
}

func TestContextRetentionUnusedExecutionCreatesOnlyLifecycleMetadata(t *testing.T) {
	f := newContextRetentionFixture(t)
	ctx := context.Background()
	record := AssignmentOperationRecord{AssignmentID: "unused", Assignment: Assignment{ID: "unused", State: AssignmentFailed, UpdatedAt: f.now}, RecordedAt: f.now}
	credentials, _ := f.store.credentialsProvider.Credentials(ctx)
	if err := f.store.saveAssignmentProjection(ctx, record, credentials); err != nil {
		t.Fatal(err)
	}
	for _, item := range f.fake.items {
		if item["pk"]["S"] == dynamoDBAssignmentContextPK("unused") && item["sk"]["S"] != assignmentContextLifecycleSK {
			t.Fatalf("unnecessary row: %v", item)
		}
	}
}
