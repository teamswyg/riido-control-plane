package riidoaiserver

import (
	"encoding/json"
	"net/http"
)

func (f *contentBundleFixture) allowed(key map[string]map[string]string, expression string, names map[string]string, values map[string]map[string]string) bool {
	item := f.items[key["pk"]["S"]+"|"+key["sk"]["S"]]
	switch expression {
	case "":
		return true // Existing DynamoDB unconditional writes are intentionally modeled.
	case "attribute_not_exists(pk)":
		return item == nil
	case "attribute_not_exists(pk) OR #schema = :legacy":
		return names["#schema"] == "schema" && values[":legacy"]["S"] == "assignment_context_snapshot/v1" && (item == nil || item["schema"]["S"] == values[":legacy"]["S"])
	default:
		return false
	}
}

func (f *contentBundleFixture) putRequest(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var put dynamoDBTransactWritePutAction
	if json.NewDecoder(r.Body).Decode(&put) != nil || !f.allowed(put.Item, put.ConditionExpression, put.ExpressionAttributeNames, put.ExpressionAttributeValues) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"__type":"ConditionalCheckFailedException"}`))
		return
	}
	f.put(put.Item)
	_, _ = w.Write([]byte(`{}`))
}
