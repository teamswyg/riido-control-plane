package riidoaiserver

import "testing"

func TestDynamoDBOutboxRejectsInvalidConfig(t *testing.T) {
	provider := mustStaticAWSTestProvider(t, "AKID", "")
	if _, err := NewDynamoDBOutbox(DynamoDBOutboxConfig{
		TableName:           "events",
		CredentialsProvider: provider,
	}); err == nil {
		t.Fatal("expected missing region error")
	}
	if _, err := NewDynamoDBOutbox(DynamoDBOutboxConfig{
		Region:              "ap-northeast-2",
		CredentialsProvider: provider,
	}); err == nil {
		t.Fatal("expected missing table error")
	}
	if _, err := NewDynamoDBOutbox(DynamoDBOutboxConfig{
		Region:    "ap-northeast-2",
		TableName: "events",
	}); err == nil {
		t.Fatal("expected missing credentials provider error")
	}
	if _, err := NewDynamoDBOutbox(DynamoDBOutboxConfig{
		Region:              "ap-northeast-2",
		TableName:           "events",
		Endpoint:            "http://dynamodb.local?debug=true",
		CredentialsProvider: provider,
	}); err == nil {
		t.Fatal("expected invalid endpoint error")
	}
	if _, err := NewStaticAWSCredentialsProvider("AKID", "", ""); err == nil {
		t.Fatal("expected missing secret key error")
	}
}
