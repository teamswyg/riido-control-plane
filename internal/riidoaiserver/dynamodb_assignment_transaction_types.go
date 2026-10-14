package riidoaiserver

import (
	"encoding/json"
	"errors"
	"strings"
)

type dynamoDBTransactWritePut struct {
	Put struct {
		TableName                 string                       `json:"TableName"`
		ConditionExpression       string                       `json:"ConditionExpression"`
		ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues,omitempty"`
		Item                      map[string]map[string]string `json:"Item"`
	} `json:"Put"`
}

type dynamoDBTransactWriteItem struct {
	Put            *dynamoDBTransactWritePutAction            `json:"Put,omitempty"`
	ConditionCheck *dynamoDBTransactWriteConditionCheckAction `json:"ConditionCheck,omitempty"`
}

type dynamoDBTransactWritePutAction struct {
	TableName                 string                       `json:"TableName"`
	ConditionExpression       string                       `json:"ConditionExpression"`
	ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues,omitempty"`
	Item                      map[string]map[string]string `json:"Item"`
}

type dynamoDBTransactWriteConditionCheckAction struct {
	TableName                 string                       `json:"TableName"`
	Key                       map[string]map[string]string `json:"Key"`
	ConditionExpression       string                       `json:"ConditionExpression"`
	ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues,omitempty"`
}

func isDynamoDBTransactionContention(err error) bool {
	var apiErr dynamoDBAPIError
	if !errors.As(err, &apiErr) {
		return false
	}
	var dynamoErr struct {
		Type       string `json:"__type"`
		Code       string `json:"Code"`
		Message    string `json:"message"`
		MessageAlt string `json:"Message"`
	}
	_ = json.Unmarshal(apiErr.body, &dynamoErr)
	errorText := dynamoErr.Type + " " + dynamoErr.Code + " " +
		dynamoErr.Message + " " + dynamoErr.MessageAlt + " " + string(apiErr.body)
	return strings.Contains(errorText, "TransactionCanceledException") ||
		strings.Contains(errorText, "ConditionalCheckFailedException")
}
