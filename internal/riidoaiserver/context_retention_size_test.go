package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

func TestContextRetentionCaptureRejectsOversizedStoredItemsBeforeWriting(t *testing.T) {
	for _, test := range []struct {
		sk    string
		limit int
	}{{"SNAPSHOT", contextSnapshotItemBytes}, {"COMMENTS#0000", contextCommentItemBytes}} {
		t.Run(test.sk, func(t *testing.T) {
			f := newContextRetentionFixture(t)
			item := map[string]map[string]string{"pk": {"S": dynamoDBAssignmentContextPK("assignment")}, "sk": {"S": test.sk}, "payload_gzip": {"S": strings.Repeat("x", test.limit)}}
			if err := f.store.saveContextPayloads(context.Background(), "assignment", []map[string]map[string]string{item}); err == nil {
				t.Fatal("oversized context accepted")
			}
			if len(f.fake.calls) != 0 || len(f.fake.items) != 0 {
				t.Fatal("oversized payload reached storage")
			}
		})
	}
}

func TestContextRetentionAggregateBudgetIncludesUpdatedItemBytes(t *testing.T) {
	contextBytes := contextSnapshotItemBytes + assignmentContextMaxCommentPages*contextCommentItemBytes + contextLifecycleItemBytes + contextQuotaItemBytes
	if contextBytes >= 1280<<10 {
		t.Fatalf("context budget=%d", contextBytes)
	}
	if contextBytes+contextTerminalBaseActions*(400<<10) >= 4<<20 {
		t.Fatal("terminal transaction can exceed DynamoDB aggregate byte limit")
	}
	f := newContextRetentionFixture(t)
	if _, err := f.store.contextLifecycleFence(strings.Repeat("x", 1024), nil, 0); err == nil {
		t.Fatal("oversized lifecycle accepted")
	}
	if _, err := assignmentContextQuotaKey(strings.Repeat("x", 1024)); err == nil {
		t.Fatal("oversized quota accepted")
	}
}
