package riidoaiserver

type dynamoDBTransactWriteUpdateAction struct {
	TableName                 string                       `json:"TableName"`
	Key                       map[string]map[string]string `json:"Key"`
	UpdateExpression          string                       `json:"UpdateExpression"`
	ConditionExpression       string                       `json:"ConditionExpression,omitempty"`
	ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues,omitempty"`
}
