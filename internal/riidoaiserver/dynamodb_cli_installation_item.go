package riidoaiserver

import "encoding/json"

const (
	cliInstallationIntentItemSchema = "cli_installation_intent/v1"
	cliInstallationLinkItemSchema   = "cli_installation_link/v1"
)

func cliInstallationIntentKey(id string) map[string]map[string]string {
	return map[string]map[string]string{"pk": {"S": "CLI_INSTALLATION_INTENT#" + id}, "sk": {"S": "INTENT"}}
}

func cliInstallationLinkKey(id cliInstallationIdentity) map[string]map[string]string {
	return map[string]map[string]string{"pk": {"S": "CLI_INSTALLATION_LINK#" + id.key()}, "sk": {"S": "LINK"}}
}

func cliInstallationItem(key map[string]map[string]string, schema string, value any) (map[string]map[string]string, error) {
	data, err := json.Marshal(value)
	if err != nil || len(data) > 3<<10 {
		return nil, errCLIInstallationInvalid
	}
	return map[string]map[string]string{"pk": key["pk"], "sk": key["sk"], "schema": {"S": schema}, "payload": {"S": string(data)}}, nil
}

func decodeCLIInstallationItem(item, key map[string]map[string]string, schema string, value any) error {
	if dynamoDBStringValue(item, "pk") != key["pk"]["S"] || dynamoDBStringValue(item, "sk") != key["sk"]["S"] || dynamoDBStringValue(item, "schema") != schema || len(dynamoDBStringValue(item, "payload")) > 3<<10 {
		return errCLIInstallationUnavailable
	}
	if json.Unmarshal([]byte(dynamoDBStringValue(item, "payload")), value) != nil {
		return errCLIInstallationUnavailable
	}
	return nil
}

func (s *DynamoDBAssignmentOperationStore) installationItemCondition(item, key map[string]map[string]string) *dynamoDBTransactWriteConditionCheckAction {
	check := &dynamoDBTransactWriteConditionCheckAction{TableName: s.tableName, Key: key, ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)"}
	if len(item) > 0 {
		check.ConditionExpression = "#schema = :schema AND payload = :payload"
		check.ExpressionAttributeNames = map[string]string{"#schema": "schema"}
		check.ExpressionAttributeValues = map[string]map[string]string{":schema": item["schema"], ":payload": item["payload"]}
	}
	return check
}

func (s *DynamoDBAssignmentOperationStore) installationItemPut(item, old map[string]map[string]string) dynamoDBTransactWriteItem {
	put := &dynamoDBTransactWritePutAction{TableName: s.tableName, Item: item, ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)"}
	if len(old) > 0 {
		put.ConditionExpression = "#schema = :expected_schema AND payload = :expected_payload"
		put.ExpressionAttributeNames = map[string]string{"#schema": "schema"}
		put.ExpressionAttributeValues = map[string]map[string]string{":expected_schema": old["schema"], ":expected_payload": old["payload"]}
	}
	return dynamoDBTransactWriteItem{Put: put}
}
