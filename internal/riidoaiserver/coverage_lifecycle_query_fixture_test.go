package riidoaiserver

import (
	"sort"
	"strings"
	"testing"
)

func (f *coverageLifecycleDynamo) query(t *testing.T, req coverageLifecycleRequest, write func(any), reject func()) {
	t.Helper()
	if req.KeyConditionExpression != "pk = :pk" && req.KeyConditionExpression != "pk = :pk AND begins_with(sk, :read)" && (req.IndexName != dynamoDBAssignmentQueueIndex || req.KeyConditionExpression != "agent_id = :agent_id") {
		t.Error("unsupported query condition")
		reject()
		return
	}
	if req.ExclusiveStartKey != nil {
		t.Error("unexpected unimplemented query pagination")
		reject()
		return
	}
	var items []map[string]map[string]string
	for _, item := range f.items {
		match := req.KeyConditionExpression == "pk = :pk" && item["pk"]["S"] == req.ExpressionAttributeValues[":pk"]["S"]
		if req.IndexName == dynamoDBAssignmentQueueIndex && req.KeyConditionExpression == "agent_id = :agent_id" {
			match = item["agent_id"]["S"] == req.ExpressionAttributeValues[":agent_id"]["S"] && item["assignment_sort"]["S"] != ""
		}
		if req.KeyConditionExpression == "pk = :pk AND begins_with(sk, :read)" {
			match = item["pk"]["S"] == req.ExpressionAttributeValues[":pk"]["S"] && strings.HasPrefix(item["sk"]["S"], req.ExpressionAttributeValues[":read"]["S"])
		}
		if match {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool { return membershipDynamoKey(items[i]) < membershipDynamoKey(items[j]) })
	if req.Limit > 0 && len(items) > req.Limit {
		t.Error("fixture result exceeds query limit")
		reject()
		return
	}
	write(map[string]any{"Items": items})
}
