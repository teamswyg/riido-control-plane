package riidoaiserver

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestDynamoDBAssignmentContextCommentPages(t *testing.T) {
	fake := &assignmentContextFakeDynamoDB{}
	endpoint := httptest.NewServer(fake)
	defer endpoint.Close()
	credentials, _ := NewStaticAWSCredentialsProvider("test", "test", "")
	store, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{Region: "test", TableName: "test", Endpoint: endpoint.URL, CredentialsProvider: credentials})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	pages := pageAssignmentContextComments(extendedAssignmentContextForTest(25).Extended.Comments)
	if err := store.SaveAssignmentContextComments(context.Background(), "asg-1", pages); err != nil {
		t.Fatal(err)
	}
	second, found, err := store.LoadAssignmentContextCommentPage(context.Background(), "asg-1", 1)
	if err != nil || !found || len(second) != 5 || second[0].ID != "comment-20" {
		t.Fatalf("page=%+v found=%v err=%v", second, found, err)
	}
	if _, found, _ := store.LoadAssignmentContextCommentPage(context.Background(), "asg-1", 2); found {
		t.Fatal("missing page reported as found")
	}
}
