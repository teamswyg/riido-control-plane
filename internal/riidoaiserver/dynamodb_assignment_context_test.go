package riidoaiserver

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDynamoDBAssignmentContextRoundTrip(t *testing.T) {
	fake := &assignmentContextFakeDynamoDB{}
	endpoint := httptest.NewServer(fake)
	defer endpoint.Close()
	credentials, _ := NewStaticAWSCredentialsProvider("test", "test", "")
	store, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{Region: "test", TableName: "test", Endpoint: endpoint.URL, CredentialsProvider: credentials})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	at := time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC)
	snapshot := AssignmentContextSnapshot{AssignmentID: "asg-1", TaskID: "task-private", AgentID: "agent-1", TriggerKind: AssignmentContextTriggerComment, TriggerID: "comment-9", CapturedAt: at, Context: assignmentContextTestTaskContext()}
	if err := store.SaveAssignmentContextSnapshot(ctx, snapshot); err != nil {
		t.Fatal(err)
	}
	loaded, found, err := store.LoadAssignmentContextSnapshot(ctx, "asg-1")
	if err != nil || !found || loaded.TriggerID != "comment-9" || loaded.Context.Hierarchy.Project.Document.Content != "<p>project body</p>" {
		t.Fatalf("loaded=%+v found=%v err=%v", loaded, found, err)
	}
	if _, found, err := store.LoadAssignmentContextSnapshot(ctx, "asg-missing"); found || err != nil {
		t.Fatalf("missing snapshot found=%v err=%v", found, err)
	}
	reads := make([]AssignmentContextRead, 0, 101)
	for i := range 101 {
		reads = append(reads, AssignmentContextRead{AssignmentID: "asg-1", SourceKind: "task", SourceID: fmt.Sprint(i), Via: AssignmentContextViaCLI, Status: AssignmentContextReadServed, ServedAt: at})
	}
	if err := store.AppendAssignmentContextReads(ctx, reads); err != nil {
		t.Fatal(err)
	}
	listed, err := store.ListAssignmentContextReads(ctx, "asg-1")
	if err != nil || len(listed) != 101 {
		t.Fatalf("listed=%d err=%v", len(listed), err)
	}
	if got := strings.Count(strings.Join(fake.calls, ","), dynamoDBTransactWriteTarget); got != 2 {
		t.Fatalf("101 reads must use two transactions, got %d", got)
	}
	for key, item := range fake.items {
		if !strings.HasPrefix(key, dynamoDBAssignmentContextPKPrefix) || item["agent_id"] != nil || item["assignment_sort"] != nil {
			t.Fatalf("item %s must stay outside the agent_queue index: %v", key, item)
		}
	}
}

func TestBoundAssignmentContextKeepsSnapshotInOneItem(t *testing.T) {
	c := assignmentContextTestTaskContext()
	c.Document.Content = strings.Repeat("가", 400<<10)
	c.Document.AlternateContent = "alt"
	original := c.Hierarchy.Project.Document
	bounded, truncated := boundAssignmentContext(c)
	if !truncated || !assignmentContextFits(bounded) || bounded.Document.AlternateContent != "" || !strings.HasPrefix(bounded.Document.Content, "가") {
		t.Fatalf("truncated=%v fits=%v", truncated, assignmentContextFits(bounded))
	}
	if strings.ToValidUTF8(bounded.Document.Content, "?") != bounded.Document.Content || original.Content != "<p>project body</p>" || bounded.Hierarchy.Project.Document == original {
		t.Fatal("truncation must keep UTF-8 valid and leave the caller's context untouched")
	}
	if small, cut := boundAssignmentContext(assignmentContextTestTaskContext()); cut || small.Document.Content != "<p>task body</p>" {
		t.Fatal("a small context must be stored as is")
	}
}
