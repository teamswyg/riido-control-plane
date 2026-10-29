package riidoaiserver

import (
	"testing"
)

func (f *coverageLifecycleDynamo) transactWrite(t *testing.T, req coverageLifecycleRequest, write func(any), reject func()) {
	t.Helper()
	seen := map[string]bool{}
	for _, action := range req.TransactItems {
		check := action.Put
		if action.Update != nil {
			if !f.validTTLUpdate(t, action.Update) {
				reject()
				return
			}
			check = &coverageLifecycleAction{TableName: action.Update.TableName, Key: action.Update.Key, ConditionExpression: action.Update.ConditionExpression, ExpressionAttributeValues: action.Update.ExpressionAttributeValues}
		}
		if check == nil {
			check = action.Delete
		}
		if check == nil {
			check = action.ConditionCheck
		}
		if check == nil || !f.sameTable(t, check.TableName) {
			reject()
			return
		}
		key := check.Key
		if key == nil {
			key = check.Item
		}
		identity := membershipDynamoKey(key)
		if seen[identity] {
			t.Error("duplicate transaction item")
			reject()
			return
		}
		seen[identity] = true
		if !f.condition(t, check) {
			reject()
			return
		}
	}
	for _, action := range req.TransactItems {
		if action.Update != nil {
			f.applyTTLUpdate(action.Update)
		}
		if action.Put != nil {
			f.items[membershipDynamoKey(action.Put.Item)] = action.Put.Item
		}
		if action.Delete != nil {
			delete(f.items, membershipDynamoKey(action.Delete.Key))
		}
	}
	write(map[string]any{})
}
