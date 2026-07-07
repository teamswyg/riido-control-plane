package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func TestDynamoDBStreamRelayRunReturnsCredentialError(t *testing.T) {
	want := errors.New("stream credentials unavailable")
	relay, err := NewDynamoDBStreamRelay(DynamoDBStreamRelayConfig{
		Region:              "ap-northeast-2",
		StreamARN:           "arn:aws:dynamodb:ap-northeast-2:123456789012:table/events/stream/2026-07-07T00:00:00.000",
		CredentialsProvider: failingAWSCredentialsProvider{err: want},
		Publisher:           &fakeStreamRelayPublisher{},
	})
	if err != nil {
		t.Fatalf("NewDynamoDBStreamRelay: %v", err)
	}
	if err := relay.Run(context.Background()); !errors.Is(err, want) {
		t.Fatalf("Run error = %v, want %v", err, want)
	}
}

func TestDynamoDBStreamRelayRunRejectsNilRelay(t *testing.T) {
	if err := (*DynamoDBStreamRelay)(nil).Run(context.Background()); err == nil {
		t.Fatal("expected nil relay error")
	}
}
