package riidoaiserver

import (
	"reflect"
	"strings"
	"testing"
)

func (f *coverageLifecycleDynamo) condition(t *testing.T, a *coverageLifecycleAction) bool {
	t.Helper()
	key := a.Key
	if key == nil {
		key = a.Item
	}
	current := f.items[membershipDynamoKey(key)]
	condition := strings.TrimSpace(a.ConditionExpression)
	if condition == "" {
		return true
	}
	if strings.Contains(condition, " OR ") || strings.Contains(condition, " <= ") {
		evaluator := membershipSharedDynamoFixture{items: f.items}
		return evaluator.condition(t, &membershipDynamoAction{Key: key, ConditionExpression: condition, ExpressionAttributeValues: a.ExpressionAttributeValues})
	}
	for _, clause := range strings.Split(condition, " AND ") {
		if strings.HasPrefix(clause, "attribute_not_exists(") && strings.HasSuffix(clause, ")") {
			field := strings.TrimSuffix(strings.TrimPrefix(clause, "attribute_not_exists("), ")")
			if current[field] != nil {
				return false
			}
			continue
		}
		if strings.HasPrefix(clause, "attribute_exists(") && strings.HasSuffix(clause, ")") {
			field := strings.TrimSuffix(strings.TrimPrefix(clause, "attribute_exists("), ")")
			if current[field] == nil {
				return false
			}
			continue
		}
		parts := strings.Split(clause, " = ")
		if len(parts) != 2 {
			t.Errorf("unsupported condition clause %s", clause)
			return false
		}
		field := parts[0]
		if name := a.ExpressionAttributeNames[field]; name != "" {
			field = name
		}
		expected := a.ExpressionAttributeValues[parts[1]]
		if expected == nil {
			t.Error("missing condition parameter")
			return false
		}
		if !reflect.DeepEqual(current[field], expected) {
			return false
		}
	}
	return true
}
