package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCoverageLifecycleFixtureChecksAllConditionsBeforeWriting(t *testing.T) {
	fixture, _ := newCoverageLifecycleDynamo(t)
	original := map[string]map[string]string{"pk": {"S": "guard"}, "sk": {"S": "one"}, "revision": {"N": "2"}}
	fixture.items[membershipDynamoKey(original)] = original
	proposed := map[string]map[string]string{"pk": {"S": "new"}, "sk": {"S": "two"}}
	actions := []dynamoDBTransactWriteItem{
		{Put: &dynamoDBTransactWritePutAction{TableName: "synthetic", Item: proposed, ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)"}},
		{ConditionCheck: &dynamoDBTransactWriteConditionCheckAction{TableName: "synthetic", Key: map[string]map[string]string{"pk": original["pk"], "sk": original["sk"]}, ConditionExpression: "revision = :expected", ExpressionAttributeValues: map[string]map[string]string{":expected": {"N": "1"}}}},
	}
	payload, err := json.Marshal(struct{ TransactItems []dynamoDBTransactWriteItem }{actions})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))
	request.Header.Set("X-Amz-Target", dynamoDBTransactWriteTarget)
	response := httptest.NewRecorder()
	fixture.handle(t, response, request)
	if response.Code != http.StatusBadRequest || len(fixture.items[membershipDynamoKey(proposed)]) != 0 || len(fixture.items) != 1 {
		t.Fatal("fixture committed a partial transaction after failed CAS")
	}
}
