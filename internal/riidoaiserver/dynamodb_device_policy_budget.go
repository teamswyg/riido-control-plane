package riidoaiserver

import (
	"reflect"
	"strings"
)

func isDevicePolicyAdvance(action dynamoDBTransactWriteItem) bool {
	if action.Update == nil || action.Put != nil || action.Delete != nil || action.ConditionCheck != nil {
		return false
	}
	update := action.Update
	device := strings.TrimPrefix(dynamoDBStringValue(update.Key, "pk"), "DEVICE_POLICY#")
	if !cliInstallationID(device) || dynamoDBStringValue(update.Key, "pk") == device {
		return false
	}
	store := DynamoDBAssignmentOperationStore{tableName: update.TableName}
	return reflect.DeepEqual(action, store.devicePolicyAdvanceAction(device))
}
