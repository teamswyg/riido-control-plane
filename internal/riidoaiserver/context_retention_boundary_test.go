package riidoaiserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"math/rand/v2"
	"os"
	"strings"
	"testing"
)

func TestContextRetentionCaptureBudgetFailureReportsCommentsUnavailable(t *testing.T) {
	f := newContextRetentionFixture(t)
	data := make([]byte, 200<<10)
	_, _ = rand.NewChaCha8([32]byte{59}).Read(data)
	snapshot := AssignmentContextSnapshot{AssignmentID: "assignment", Context: AIAgentTaskContext{Extended: &AIAgentTaskContextExtended{CommentStatus: "available"}}}
	server := Server{config: ServerConfig{AssignmentContext: f.store}}
	pages := server.saveAssignmentContextComments(context.Background(), &snapshot, []AIAgentTaskContextComment{{ID: "large", Body: base64.StdEncoding.EncodeToString(data)}})
	if pages != 0 || snapshot.Context.Extended.CommentStatus != "unavailable" || len(f.fake.items) != 0 {
		t.Fatal("oversized comments were not marked unavailable before persistence")
	}
}

func TestContextRetentionRejectsUnbudgetedTerminalShape(t *testing.T) {
	f := newContextRetentionFixture(t)
	record := AssignmentOperationRecord{AssignmentID: "assignment", Assignment: Assignment{ID: "assignment", State: AssignmentCompleted, UpdatedAt: f.now}}
	item, err := assignmentProjectionDynamoDBItem(record)
	if err != nil {
		t.Fatal(err)
	}
	for _, count := range []int{2, 7} {
		actions := make([]dynamoDBTransactWriteItem, count)
		actions[0].Put = &dynamoDBTransactWritePutAction{TableName: "test", Item: item}
		if count == 2 {
			actions[1] = actions[0]
		}
		payload, _ := json.Marshal(struct{ TransactItems []dynamoDBTransactWriteItem }{actions})
		if _, err := f.store.writeAssignmentWithContextRetention(context.Background(), dynamoDBRequest{target: dynamoDBTransactWriteTarget, payload: payload}); err == nil {
			t.Fatal("unbudgeted terminal transaction accepted")
		}
	}
	if len(f.fake.calls) != 0 {
		t.Fatal("unbudgeted transaction reached storage")
	}
}

func TestContextRetentionAllTerminalProjectionWritersUseAtomicBoundary(t *testing.T) {
	for _, name := range []string{"dynamodb_assignment_mutation.go", "dynamodb_assignment_operation_store_save_assignment_projection.go", "dynamodb_assignment_projection_recovery.go", "dynamodb_assignment_operation_store_claim_next.go", "dynamodb_task_agent_publication_recovery.go"} {
		body, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "writeAssignmentWithContextRetention(ctx,") || strings.Contains(string(body), "doDynamoDBJSON(ctx,") {
			t.Fatalf("terminal writer bypasses context fence: %s", name)
		}
	}
}
