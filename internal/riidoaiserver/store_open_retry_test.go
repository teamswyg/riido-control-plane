package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func TestOpenStoreWithConfigRetriesTransientOperationReplayError(t *testing.T) {
	operations := &retryReplayOperationStore{
		errors: []error{errors.New("dynamodb api error: ThrottlingException")},
	}
	store, err := OpenStoreWithConfig(context.Background(), StoreConfig{OperationStore: operations})
	if err != nil {
		t.Fatalf("OpenStoreWithConfig: %v", err)
	}
	if store == nil {
		t.Fatal("OpenStoreWithConfig returned nil store")
	}
	if operations.loadCalls != 2 {
		t.Fatalf("loadCalls = %d, want 2", operations.loadCalls)
	}
}

func TestOpenStoreWithConfigDoesNotRetryPermanentOperationReplayError(t *testing.T) {
	operations := &retryReplayOperationStore{
		errors: []error{errors.New("dynamodb api error: ValidationException")},
	}
	if _, err := OpenStoreWithConfig(context.Background(), StoreConfig{OperationStore: operations}); err == nil {
		t.Fatal("OpenStoreWithConfig error = nil, want permanent failure")
	}
	if operations.loadCalls != 1 {
		t.Fatalf("loadCalls = %d, want 1", operations.loadCalls)
	}
}

type retryReplayOperationStore struct {
	errors    []error
	loadCalls int
}

func (s *retryReplayOperationStore) SaveAssignmentOperation(context.Context, AssignmentOperationRecord) error {
	return nil
}

func (s *retryReplayOperationStore) LoadAssignmentOperations(context.Context) ([]AssignmentOperationRecord, error) {
	call := s.loadCalls
	s.loadCalls++
	if call < len(s.errors) && s.errors[call] != nil {
		return nil, s.errors[call]
	}
	return nil, nil
}

func (s *retryReplayOperationStore) Close() error {
	return nil
}
