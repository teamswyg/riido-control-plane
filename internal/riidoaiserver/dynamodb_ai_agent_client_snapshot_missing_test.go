package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func TestDynamoDBAIAgentClientSnapshotLoadsMissingSnapshot(t *testing.T) {
	fixture := newSnapshotDynamoDBFixture(t, fixedSnapshotTestNow(), nil, nil)
	defer fixture.close()
	_, ok, err := fixture.store.LoadAIAgentClientSnapshot(context.Background())
	if err != nil {
		t.Fatalf("LoadAIAgentClientSnapshot: %v", err)
	}
	if ok {
		t.Fatal("expected missing snapshot")
	}
}

func TestDynamoDBAIAgentClientSnapshotRejectsInvalidConfig(t *testing.T) {
	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	if _, err := NewDynamoDBAIAgentClientSnapshot(DynamoDBAIAgentClientSnapshotConfig{Region: "ap-northeast-2", CredentialsProvider: provider}); err == nil {
		t.Fatal("expected missing table error")
	}
	if _, err := NewDynamoDBAIAgentClientSnapshot(DynamoDBAIAgentClientSnapshotConfig{Region: "ap-northeast-2", TableName: "ai-agent"}); err == nil {
		t.Fatal("expected missing credentials provider error")
	}
}

func TestDynamoDBAIAgentClientSnapshotRepliesCredentialErrors(t *testing.T) {
	want := errors.New("credentials unavailable")
	store, err := NewDynamoDBAIAgentClientSnapshot(DynamoDBAIAgentClientSnapshotConfig{
		Region:              "ap-northeast-2",
		TableName:           "riido-ai-agent-development",
		CredentialsProvider: failingAWSCredentialsProvider{err: want},
	})
	if err != nil {
		t.Fatalf("NewDynamoDBAIAgentClientSnapshot: %v", err)
	}
	defer store.Close()
	if _, _, err := store.LoadAIAgentClientSnapshot(context.Background()); !errors.Is(err, want) {
		t.Fatalf("LoadAIAgentClientSnapshot() error = %v, want %v", err, want)
	}
	if err := store.SaveAIAgentClientSnapshot(context.Background(), AIAgentClientSnapshot{}); !errors.Is(err, want) {
		t.Fatalf("SaveAIAgentClientSnapshot() error = %v, want %v", err, want)
	}
}

type failingAWSCredentialsProvider struct {
	err error
}

func (p failingAWSCredentialsProvider) Credentials(context.Context) (AWSCredentials, error) {
	return AWSCredentials{}, p.err
}
