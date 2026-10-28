package riidoaiserver

import (
	"strconv"
	"strings"
	"testing"
)

func (f *coverageLifecycleDynamo) validTTLUpdate(t *testing.T, update *dynamoDBTransactWriteUpdateAction) bool {
	t.Helper()
	if update.UpdateExpression != "SET assignment_context_expires_at_epoch = :expiry" || !strings.HasPrefix(update.Key["pk"]["S"], "ASSIGNMENT_CONTEXT#") {
		t.Error("unsupported fixture update")
		return false
	}
	expiry, err := strconv.ParseInt(update.ExpressionAttributeValues[":expiry"]["N"], 10, 64)
	if err != nil || expiry <= 0 {
		t.Error("invalid context expiry")
		return false
	}
	return true
}

func (f *coverageLifecycleDynamo) applyTTLUpdate(update *dynamoDBTransactWriteUpdateAction) {
	key := membershipDynamoKey(update.Key)
	item := f.items[key]
	if item == nil {
		item = map[string]map[string]string{"pk": update.Key["pk"], "sk": update.Key["sk"]}
	}
	item["assignment_context_expires_at_epoch"] = update.ExpressionAttributeValues[":expiry"]
	f.items[key] = item
}
