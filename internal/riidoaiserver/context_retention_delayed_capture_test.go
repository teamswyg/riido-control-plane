package riidoaiserver

import (
	"context"
	"strconv"
	"testing"
	"time"
)

func TestContextRetentionDelayedCaptureKeepsFirstTerminalTime(t *testing.T) {
	f := newContextRetentionFixture(t)
	ctx := context.Background()
	record := AssignmentOperationRecord{AssignmentID: "assignment", Assignment: Assignment{ID: "assignment", State: AssignmentCompleted, UpdatedAt: f.now}, RecordedAt: f.now}
	credentials, _ := f.store.credentialsProvider.Credentials(ctx)
	if err := f.store.saveAssignmentProjection(ctx, record, credentials); err != nil {
		t.Fatal(err)
	}
	want := strconv.FormatInt(f.now.Add(24*time.Hour).Unix(), 10)
	f.now = f.now.Add(time.Hour)
	record.Assignment.UpdatedAt = f.now
	if err := f.store.saveAssignmentProjection(ctx, record, credentials); err != nil {
		t.Fatal(err)
	}
	if err := f.store.SaveAssignmentContextSnapshot(ctx, AssignmentContextSnapshot{AssignmentID: "assignment"}); err != nil {
		t.Fatal(err)
	}
	if got := f.fake.items[dynamoDBAssignmentContextPK("assignment")+"|SNAPSHOT"][assignmentContextTTLAttribute]["N"]; got != want {
		t.Fatalf("expiry extended: got=%s want=%s", got, want)
	}
}
