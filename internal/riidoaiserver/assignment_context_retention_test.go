package riidoaiserver

import (
	"context"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestAssignmentContextExpiredPayloadCannotBeReadWhileDynamoTTLWaits(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC)
	fake := &assignmentContextFakeDynamoDB{}
	endpoint := httptest.NewServer(fake)
	defer endpoint.Close()
	credentials, _ := NewStaticAWSCredentialsProvider("test", "test", "")
	store, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{Region: "test", TableName: "test", Endpoint: endpoint.URL, CredentialsProvider: credentials, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.SaveAssignmentContextSnapshot(ctx, AssignmentContextSnapshot{AssignmentID: "assignment"}); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveAssignmentContextComments(ctx, "assignment", [][]AIAgentTaskContextComment{{{ID: "comment", Body: "private"}}}); err != nil {
		t.Fatal(err)
	}
	if err := store.AppendAssignmentContextReads(ctx, []AssignmentContextRead{{AssignmentID: "assignment", SourceKind: "task", SourceID: "task", ServedAt: now}}); err != nil {
		t.Fatal(err)
	}
	fake.mu.Lock()
	for _, sk := range []string{"SNAPSHOT", "COMMENTS#0000"} {
		fake.items["ASSIGNMENT_CONTEXT#assignment|"+sk]["assignment_context_expires_at_epoch"] = map[string]string{"N": strconv.FormatInt(now.Unix(), 10)}
	}
	fake.mu.Unlock()
	if _, found, err := store.LoadAssignmentContextSnapshot(ctx, "assignment"); err != nil || found {
		t.Fatalf("expired snapshot found=%v err=%v", found, err)
	}
	if _, found, err := store.LoadAssignmentContextCommentPage(ctx, "assignment", 0); err != nil || found {
		t.Fatalf("expired comments found=%v err=%v", found, err)
	}
	if reads, err := store.ListAssignmentContextReads(ctx, "assignment"); err != nil || len(reads) != 1 {
		t.Fatalf("read audit must remain: len=%d err=%v", len(reads), err)
	}
}
