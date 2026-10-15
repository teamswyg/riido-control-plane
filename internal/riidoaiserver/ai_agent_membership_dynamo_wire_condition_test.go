package riidoaiserver

import "encoding/json"

// Dynamo distinguishes an omitted optional condition from an invalid empty expression.
func membershipDynamoHasEmptyCondition(payload json.RawMessage) bool {
	var request struct {
		TransactItems []map[string]map[string]json.RawMessage
	}
	if err := json.Unmarshal(payload, &request); err != nil {
		return true
	}
	for _, item := range request.TransactItems {
		for _, action := range item {
			if expression, exists := action["ConditionExpression"]; exists {
				var condition string
				if err := json.Unmarshal(expression, &condition); err != nil || condition == "" {
					return true
				}
			}
		}
	}
	return false
}
