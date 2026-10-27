package riidoaiserver

import "strings"

func (f *terminationDynamoFixture) conditionsHold(actions []dynamoDBTransactWriteItem) bool {
	for _, action := range actions {
		var key, values map[string]map[string]string
		var names map[string]string
		condition := ""
		if action.Put != nil {
			key, values, condition = action.Put.Item, action.Put.ExpressionAttributeValues, action.Put.ConditionExpression
		}
		if action.ConditionCheck != nil {
			key, values, condition = action.ConditionCheck.Key, action.ConditionCheck.ExpressionAttributeValues, action.ConditionCheck.ConditionExpression
			names = action.ConditionCheck.ExpressionAttributeNames
		}
		stored := f.items[terminationItemKey(key)]
		for _, part := range strings.Split(condition, " AND ") {
			if strings.HasPrefix(part, "attribute_not_exists(") {
				if len(stored) > 0 {
					return false
				}
				continue
			}
			pieces := strings.Split(part, " = ")
			if len(pieces) != 2 {
				return false
			}
			expected := values[pieces[1]]
			field := pieces[0]
			if name := names[field]; name != "" {
				field = name
			}
			actual := stored[field]
			if expected == nil || actual == nil || expected["S"] != actual["S"] || expected["N"] != actual["N"] {
				return false
			}
		}
	}
	return true
}
