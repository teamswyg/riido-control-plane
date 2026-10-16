package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"sort"
	"testing"
)

func (f *membershipSharedDynamoFixture) handle(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()
	var p struct {
		membershipDynamoAction
		IndexName     string `json:"IndexName"`
		TransactItems []struct {
			Put            *membershipDynamoAction `json:"Put"`
			ConditionCheck *membershipDynamoAction `json:"ConditionCheck"`
			Delete         *membershipDynamoAction `json:"Delete"`
			Get            *membershipDynamoAction `json:"Get"`
		} `json:"TransactItems"`
	}
	var payload json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		t.Errorf("decode Dynamo payload: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if r.Header.Get("X-Amz-Target") == dynamoDBTransactWriteTarget && membershipDynamoHasEmptyCondition(payload) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"__type": "ValidationException", "message": "ConditionExpression must not be empty"})
		return
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		t.Errorf("decode Dynamo actions: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	f.mu.Lock()
	receiptHook, ownedReadHook, recoveryHook := f.beforeReceiptQuery, f.beforeOwnedSnapshotRead, f.beforePublicationRecovery
	generationHook := f.beforeGenerationRead
	f.mu.Unlock()
	if r.Header.Get("X-Amz-Target") == dynamoDBGetItemTarget && p.Key["sk"]["S"] == taskAgentGenerationSK && generationHook != nil {
		generationHook()
	}
	if r.Header.Get("X-Amz-Target") == dynamoDBQueryTarget && p.ExpressionAttributeValues[":exec"]["S"] != "" && receiptHook != nil {
		receiptHook()
	}
	if r.Header.Get("X-Amz-Target") == dynamoDBSnapshotTransactGetTarget && len(p.TransactItems) == 2+len(onboardingOwnedPartNames) && ownedReadHook != nil {
		ownedReadHook()
	}
	if r.Header.Get("X-Amz-Target") == dynamoDBTransactWriteTarget && recoveryHook != nil {
		for _, action := range p.TransactItems {
			if action.Put != nil && action.Put.ConditionExpression == "assignment_state = :observed_state AND last_event_seq = :observed_seq" {
				recoveryHook()
				break
			}
		}
	}
	if r.Header.Get("X-Amz-Target") == dynamoDBTransactWriteTarget {
		f.mu.Lock()
		hook, reject := f.beforeScopedCommit, f.rejectScopedPublication
		f.mu.Unlock()
		if hook != nil || reject {
			guarded := false
			var relation AIAgentClientSnapshot
			for _, action := range p.TransactItems {
				if action.ConditionCheck != nil && action.ConditionCheck.ConditionExpression == "generation = :generation" {
					guarded = true
				}
				if action.Put != nil && dynamoDBAIAgentClientSnapshotPartName(action.Put.Item) == dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships {
					if err := decodeDynamoDBAIAgentClientSnapshotPart(dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships, action.Put.Item, &relation); err != nil {
						t.Errorf("decode scoped commit membership: %v", err)
					}
				}
			}
			if guarded && reject {
				w.WriteHeader(http.StatusServiceUnavailable)
				_ = json.NewEncoder(w).Encode(map[string]string{"__type": "InternalServerError", "message": "injected scoped publication failure"})
				return
			}
			if guarded && hook != nil {
				hook(relation)
			}
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	write := func(value any) { _ = json.NewEncoder(w).Encode(value) }
	fail := func(kind string) {
		w.WriteHeader(http.StatusBadRequest)
		write(map[string]string{"__type": kind, "message": "conditional fixture rejection"})
	}
	switch r.Header.Get("X-Amz-Target") {
	case dynamoDBPutItemTarget:
		if !f.condition(t, &p.membershipDynamoAction) {
			fail("ConditionalCheckFailedException")
			return
		}
		f.items[membershipDynamoKey(p.Item)] = p.Item
		write(map[string]any{})
	case dynamoDBDeleteItemTarget:
		if !f.condition(t, &p.membershipDynamoAction) {
			fail("ConditionalCheckFailedException")
			return
		}
		delete(f.items, membershipDynamoKey(p.Key))
		write(map[string]any{})
	case dynamoDBGetItemTarget:
		write(map[string]any{"Item": f.items[membershipDynamoKey(p.Key)]})
	case dynamoDBQueryTarget:
		var out []map[string]map[string]string
		for _, item := range f.items {
			if p.IndexName != "" {
				if item["agent_id"]["S"] == p.ExpressionAttributeValues[":agent_id"]["S"] && item["assignment_sort"]["S"] != "" {
					out = append(out, item)
				}
			} else if item["pk"]["S"] == p.ExpressionAttributeValues[":pk"]["S"] {
				out = append(out, item)
			}
		}
		sort.Slice(out, func(i, j int) bool { return out[i]["sk"]["S"] < out[j]["sk"]["S"] })
		write(map[string]any{"Items": out})
	case dynamoDBSnapshotTransactGetTarget:
		var responses []map[string]any
		for _, action := range p.TransactItems {
			if action.Delete != nil {
				continue
			}
			responses = append(responses, map[string]any{"Item": f.items[membershipDynamoKey(action.Get.Key)]})
		}
		write(map[string]any{"Responses": responses})
	case dynamoDBTransactWriteTarget:
		if f.rejectCreation {
			for _, action := range p.TransactItems {
				if action.Put != nil && action.Put.Item["sk"]["S"] == taskAgentGenerationSK {
					fail("TransactionCanceledException")
					return
				}
			}
		}
		for _, action := range p.TransactItems {
			if action.Delete != nil {
				if !f.condition(t, action.Delete) {
					fail("TransactionCanceledException")
					return
				}
				continue
			}
			check := action.Put
			if check == nil {
				check = action.ConditionCheck
			}
			if check == nil || !f.condition(t, check) {
				fail("TransactionCanceledException")
				return
			}
		}
		for _, action := range p.TransactItems {
			if action.Put != nil {
				f.items[membershipDynamoKey(action.Put.Item)] = action.Put.Item
			}
			if action.Delete != nil {
				delete(f.items, membershipDynamoKey(action.Delete.Key))
			}
		}
		if f.ambiguousCreationCommit || f.loseCreationResponse {
			for _, action := range p.TransactItems {
				if action.Put != nil && action.Put.Item["sk"]["S"] == taskAgentGenerationSK {
					if f.loseCreationResponse {
						panic(http.ErrAbortHandler)
					}
					w.WriteHeader(http.StatusServiceUnavailable)
					write(map[string]string{"__type": "InternalServerError", "message": "creation committed response lost"})
					return
				}
			}
		}
		if f.ambiguousScopedCommit {
			for _, action := range p.TransactItems {
				if action.ConditionCheck != nil && action.ConditionCheck.ConditionExpression == "generation = :generation" {
					w.WriteHeader(http.StatusServiceUnavailable)
					write(map[string]string{"__type": "InternalServerError", "message": "committed response lost"})
					return
				}
			}
		}
		write(map[string]any{})
	default:
		t.Errorf("unsupported Dynamo target %q", r.Header.Get("X-Amz-Target"))
		w.WriteHeader(http.StatusBadRequest)
	}
}
