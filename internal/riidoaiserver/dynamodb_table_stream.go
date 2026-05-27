package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type DynamoDBTableStreamARNConfig struct {
	Region              string
	TableName           string
	Endpoint            string
	HTTPClient          *http.Client
	Now                 func() time.Time
	CredentialsProvider AWSCredentialsProvider
}

type DynamoDBTableDescription struct {
	TableName       string                    `json:"table_name"`
	TableStatus     string                    `json:"table_status"`
	BillingMode     string                    `json:"billing_mode"`
	LatestStreamARN string                    `json:"latest_stream_arn,omitempty"`
	KeySchema       []DynamoDBKeySchemaMember `json:"key_schema,omitempty"`
}

type DynamoDBKeySchemaMember struct {
	AttributeName string `json:"attribute_name"`
	KeyType       string `json:"key_type"`
}

func LoadDynamoDBTableStreamARN(ctx context.Context, config DynamoDBTableStreamARNConfig) (string, error) {
	table, err := DescribeDynamoDBTable(ctx, config)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(table.LatestStreamARN) == "" {
		return "", fmt.Errorf("DynamoDB table %q has no LatestStreamArn", config.TableName)
	}
	return strings.TrimSpace(table.LatestStreamARN), nil
}

func DescribeDynamoDBTable(ctx context.Context, config DynamoDBTableStreamARNConfig) (DynamoDBTableDescription, error) {
	region := strings.TrimSpace(config.Region)
	if region == "" {
		return DynamoDBTableDescription{}, errors.New("riidoaiserver: DynamoDB table stream ARN region is required")
	}
	tableName := strings.TrimSpace(config.TableName)
	if tableName == "" {
		return DynamoDBTableDescription{}, errors.New("riidoaiserver: DynamoDB table stream ARN table name is required")
	}
	if config.CredentialsProvider == nil {
		return DynamoDBTableDescription{}, errors.New("riidoaiserver: DynamoDB table stream ARN credentials provider is required")
	}
	endpoint, endpointHost, err := normalizeDynamoDBEndpoint(region, strings.TrimSpace(config.Endpoint))
	if err != nil {
		return DynamoDBTableDescription{}, err
	}
	credentials, err := cachedAWSCredentials(ctx, config.Now, config.CredentialsProvider, nil)
	if err != nil {
		return DynamoDBTableDescription{}, err
	}
	payload, err := json.Marshal(struct {
		TableName string `json:"TableName"`
	}{TableName: tableName})
	if err != nil {
		return DynamoDBTableDescription{}, err
	}
	body, err := doDynamoDBJSON(ctx, dynamoDBRequest{
		endpoint:     endpoint,
		endpointHost: endpointHost,
		region:       region,
		target:       dynamoDBDescribeTableTarget,
		payload:      payload,
		credentials:  credentials,
		httpClient:   dynamoDBHTTPClient(config.HTTPClient),
		now:          dynamoDBClock(config.Now),
	})
	if err != nil {
		return DynamoDBTableDescription{}, fmt.Errorf("dynamodb describe table: %w", err)
	}
	var response struct {
		Table struct {
			TableName          string `json:"TableName"`
			TableStatus        string `json:"TableStatus"`
			LatestStreamARN    string `json:"LatestStreamArn"`
			BillingModeSummary struct {
				BillingMode string `json:"BillingMode"`
			} `json:"BillingModeSummary"`
			KeySchema []struct {
				AttributeName string `json:"AttributeName"`
				KeyType       string `json:"KeyType"`
			} `json:"KeySchema"`
		} `json:"Table"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return DynamoDBTableDescription{}, fmt.Errorf("decode DynamoDB describe table response: %w", err)
	}
	billingMode := strings.TrimSpace(response.Table.BillingModeSummary.BillingMode)
	if billingMode == "" {
		billingMode = "PROVISIONED"
	}
	description := DynamoDBTableDescription{
		TableName:       strings.TrimSpace(response.Table.TableName),
		TableStatus:     strings.TrimSpace(response.Table.TableStatus),
		BillingMode:     billingMode,
		LatestStreamARN: strings.TrimSpace(response.Table.LatestStreamARN),
	}
	if description.TableName == "" {
		description.TableName = tableName
	}
	for _, key := range response.Table.KeySchema {
		description.KeySchema = append(description.KeySchema, DynamoDBKeySchemaMember{
			AttributeName: strings.TrimSpace(key.AttributeName),
			KeyType:       strings.TrimSpace(key.KeyType),
		})
	}
	return description, nil
}
