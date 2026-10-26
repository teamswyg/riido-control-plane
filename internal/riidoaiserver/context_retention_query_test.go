package riidoaiserver

import (
	"context"
	"testing"
)

type contextRetentionQueryPage struct {
	Items            []map[string]map[string]string
	LastEvaluatedKey map[string]map[string]string
}

func contextCommentTestKey(page int) map[string]map[string]string {
	return map[string]map[string]string{"pk": {"S": dynamoDBAssignmentContextPK("assignment")}, "sk": {"S": dynamoDBAssignmentContextCommentSK(page)}}
}

func TestContextRetentionTraversesDynamoBytePages(t *testing.T) {
	f := newContextRetentionFixture(t)
	f.queryPages = []contextRetentionQueryPage{
		{Items: []map[string]map[string]string{contextCommentTestKey(0)}, LastEvaluatedKey: contextCommentTestKey(0)},
		{Items: []map[string]map[string]string{contextCommentTestKey(1), contextCommentTestKey(2)}},
	}
	items, err := f.store.contextRetentionPayloads(context.Background(), "assignment")
	if err != nil || len(items) != 3 || len(f.queryPages) != 0 {
		t.Fatalf("items=%d remaining=%d err=%v", len(items), len(f.queryPages), err)
	}
}

func TestContextRetentionRejectsInvalidOrRepeatedQueryKeys(t *testing.T) {
	for _, test := range []struct {
		name  string
		pages []contextRetentionQueryPage
	}{
		{"duplicate", []contextRetentionQueryPage{{Items: []map[string]map[string]string{contextCommentTestKey(0)}, LastEvaluatedKey: contextCommentTestKey(0)}, {Items: []map[string]map[string]string{contextCommentTestKey(0)}}}},
		{"cursor cycle", []contextRetentionQueryPage{{LastEvaluatedKey: contextCommentTestKey(0)}, {LastEvaluatedKey: contextCommentTestKey(0)}}},
		{"cursor backwards", []contextRetentionQueryPage{{LastEvaluatedKey: contextCommentTestKey(1)}, {LastEvaluatedKey: contextCommentTestKey(0)}}},
		{"outside pages", []contextRetentionQueryPage{{Items: []map[string]map[string]string{contextCommentTestKey(10)}}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := newContextRetentionFixture(t)
			f.queryPages = test.pages
			if _, err := f.store.contextRetentionPayloads(context.Background(), "assignment"); err == nil {
				t.Fatal("unsafe key/cursor accepted")
			}
		})
	}
}

func TestContextRetentionTerminalConflictIsNotAcknowledged(t *testing.T) {
	f := newContextRetentionFixture(t)
	record := f.projection(t, "assignment", AssignmentRunning)
	record.Assignment.State = AssignmentCompleted
	f.deny = true
	credentials, _ := f.store.credentialsProvider.Credentials(context.Background())
	err := f.store.saveAssignmentProjection(context.Background(), record, credentials)
	if err == nil || f.writes != 3 {
		t.Fatalf("uncommitted terminal err=%v writes=%d", err, f.writes)
	}
}
