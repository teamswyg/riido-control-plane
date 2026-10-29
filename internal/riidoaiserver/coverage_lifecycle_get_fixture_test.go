package riidoaiserver

import (
	"testing"
)

func (f *coverageLifecycleDynamo) transactGet(t *testing.T, req coverageLifecycleRequest, write func(any), reject func()) {
	t.Helper()
	var responses []map[string]any
	for _, action := range req.TransactItems {
		if action.Get == nil {
			t.Error("missing transaction Get")
			reject()
			return
		}
		if !f.sameTable(t, action.Get.TableName) {
			reject()
			return
		}
		responses = append(responses, map[string]any{"Item": f.items[membershipDynamoKey(action.Get.Key)]})
	}
	write(map[string]any{"Responses": responses})
}
