package riidoaiserver

type dynamoDBPutPayload struct {
	TableName                 string                       `json:"TableName"`
	ConditionExpression       string                       `json:"ConditionExpression"`
	ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
	Item                      map[string]map[string]string `json:"Item"`
}

type dynamoDBTransactWritePayload struct {
	TransactItems []struct {
		Put dynamoDBTransactPut `json:"Put"`
	} `json:"TransactItems"`
}

type dynamoDBRepairTransactWritePayload struct {
	TransactItems []struct {
		Put            *dynamoDBTransactPut            `json:"Put,omitempty"`
		ConditionCheck *dynamoDBTransactConditionCheck `json:"ConditionCheck,omitempty"`
	} `json:"TransactItems"`
}

type dynamoDBTransactPut struct {
	TableName                 string                       `json:"TableName"`
	ConditionExpression       string                       `json:"ConditionExpression"`
	ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
	Item                      map[string]map[string]string `json:"Item"`
}

type dynamoDBTransactConditionCheck struct {
	TableName                 string                       `json:"TableName"`
	Key                       map[string]map[string]string `json:"Key"`
	ConditionExpression       string                       `json:"ConditionExpression"`
	ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
}

type dynamoDBDeletePayload struct {
	TableName                 string                       `json:"TableName"`
	Key                       map[string]map[string]string `json:"Key"`
	ConditionExpression       string                       `json:"ConditionExpression"`
	ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
}

type dynamoDBUpdatePayload struct {
	TableName                 string                       `json:"TableName"`
	Key                       map[string]map[string]string `json:"Key"`
	ConditionExpression       string                       `json:"ConditionExpression"`
	UpdateExpression          string                       `json:"UpdateExpression"`
	ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
}
