package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"
)

func TestContextRetentionCaptureRacingTerminalRetriesWithOriginalExpiry(t *testing.T) {
	f := newContextRetentionFixture(t)
	record := f.projection(t, "assignment", AssignmentRunning)
	record.Assignment.State = AssignmentCancelled
	terminal, err := assignmentProjectionDynamoDBItem(record)
	if err != nil {
		t.Fatal(err)
	}
	f.beforeWrite = func() { f.fake.mu.Lock(); defer f.fake.mu.Unlock(); f.fake.put(terminal) }
	if err := f.store.SaveAssignmentContextSnapshot(context.Background(), AssignmentContextSnapshot{AssignmentID: "assignment"}); err != nil {
		t.Fatal(err)
	}
	if f.writes != 2 {
		t.Fatalf("writes=%d, expected one fenced retry", f.writes)
	}
	if got := f.fake.items[dynamoDBAssignmentContextPK("assignment")+"|SNAPSHOT"][assignmentContextTTLAttribute]["N"]; got != strconv.FormatInt(f.now.Add(24*time.Hour).Unix(), 10) {
		t.Fatalf("late capture ttl=%s", got)
	}
}

func TestContextRetentionTerminalRacingNewCommentPageIncludesIt(t *testing.T) {
	f := newContextRetentionFixture(t)
	ctx := context.Background()
	record := f.projection(t, "assignment", AssignmentRunning)
	if err := f.store.SaveAssignmentContextSnapshot(ctx, AssignmentContextSnapshot{AssignmentID: "assignment"}); err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal([]AIAgentTaskContextComment{{ID: "racing-comment"}})
	compressed, _ := gzipBase64(encoded)
	f.beforeWrite = func() {
		f.fake.mu.Lock()
		defer f.fake.mu.Unlock()
		f.fake.items[dynamoDBAssignmentContextPK("assignment")+"|LIFECYCLE"]["revision"] = map[string]string{"S": "new-capture"}
		f.fake.put(map[string]map[string]string{"pk": {"S": dynamoDBAssignmentContextPK("assignment")}, "sk": {"S": "COMMENTS#0000"}, "schema": {"S": "assignment_context_comments/v1"}, "payload_gzip": {"S": compressed}})
	}
	record.Assignment.State = AssignmentCompleted
	credentials, _ := f.store.credentialsProvider.Credentials(ctx)
	before := f.writes
	if err := f.store.saveAssignmentProjection(ctx, record, credentials); err != nil {
		t.Fatal(err)
	}
	if f.writes != before+2 {
		t.Fatalf("writes=%d before=%d", f.writes, before)
	}
	if got := f.fake.items[dynamoDBAssignmentContextPK("assignment")+"|COMMENTS#0000"][assignmentContextTTLAttribute]["N"]; got != strconv.FormatInt(f.now.Add(24*time.Hour).Unix(), 10) {
		t.Fatalf("racing page ttl=%s", got)
	}
}

func TestContextRetentionContentionRetryIsBounded(t *testing.T) {
	f := newContextRetentionFixture(t)
	f.deny = true
	err := f.store.SaveAssignmentContextSnapshot(context.Background(), AssignmentContextSnapshot{AssignmentID: "assignment"})
	if !errors.Is(err, errAssignmentProjectionConflict) || f.writes != 3 {
		t.Fatalf("err=%v writes=%d", err, f.writes)
	}
	if len(f.fake.items) != 0 {
		t.Fatal("failed capture left payload rows")
	}
}
