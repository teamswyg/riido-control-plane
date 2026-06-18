package riidoaiserver

import (
	"context"
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
