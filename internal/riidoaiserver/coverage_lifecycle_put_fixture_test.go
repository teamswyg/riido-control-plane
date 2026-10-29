package riidoaiserver

import (
	"testing"
)

func (f *coverageLifecycleDynamo) putOrDelete(t *testing.T, req coverageLifecycleRequest, write func(any), reject func(), target string) {
	t.Helper()
	if !f.condition(t, &req.coverageLifecycleAction) {
		reject()
		return
	}
	if target == dynamoDBPutItemTarget {
		f.items[membershipDynamoKey(req.Item)] = req.Item
	} else {
		delete(f.items, membershipDynamoKey(req.Key))
	}
	write(map[string]any{})
}
