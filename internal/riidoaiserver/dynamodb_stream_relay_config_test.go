package riidoaiserver

import "testing"

func TestDynamoDBStreamRelayRejectsInvalidConfig(t *testing.T) {
	provider := mustStaticAWSTestProvider(t, "AKID", "")
	publisher := &fakeStreamRelayPublisher{}
	if _, err := NewDynamoDBStreamRelay(DynamoDBStreamRelayConfig{
		StreamARN: "arn", CredentialsProvider: provider, Publisher: publisher,
	}); err == nil {
		t.Fatal("expected missing region error")
	}
	if _, err := NewDynamoDBStreamRelay(DynamoDBStreamRelayConfig{
		Region: "ap-northeast-2", CredentialsProvider: provider, Publisher: publisher,
	}); err == nil {
		t.Fatal("expected missing stream ARN error")
	}
	if _, err := NewDynamoDBStreamRelay(DynamoDBStreamRelayConfig{
		Region: "ap-northeast-2", StreamARN: "arn", Publisher: publisher,
	}); err == nil {
		t.Fatal("expected missing credentials provider error")
	}
	if _, err := NewDynamoDBStreamRelay(DynamoDBStreamRelayConfig{
		Region: "ap-northeast-2", StreamARN: "arn", CredentialsProvider: provider,
	}); err == nil {
		t.Fatal("expected missing publisher error")
	}
	if _, err := NewDynamoDBStreamRelay(DynamoDBStreamRelayConfig{
		Region: "ap-northeast-2", StreamARN: "arn", CredentialsProvider: provider,
		Publisher: publisher, Limit: 1001,
	}); err == nil {
		t.Fatal("expected invalid limit error")
	}
}
