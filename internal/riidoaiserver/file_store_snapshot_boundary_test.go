package riidoaiserver

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFileStoreSnapshotNilAndCanceledBoundaries(t *testing.T) {
	var nilStore *FileStoreSnapshot
	var nilContext context.Context
	if snapshot, ok, err := nilStore.LoadStoreSnapshot(nilContext); err != nil || ok || snapshot.SchemaVersion != "" {
		t.Fatalf("nil LoadStoreSnapshot() = (%+v, %v, %v), want zero false nil", snapshot, ok, err)
	}
	if err := nilStore.SaveStoreSnapshot(nilContext, StoreSnapshot{}); err != nil {
		t.Fatalf("nil SaveStoreSnapshot() error = %v, want nil", err)
	}

	store := newFileStoreSnapshotForBoundary(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := store.LoadStoreSnapshot(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("LoadStoreSnapshot canceled error = %v", err)
	}
	if err := store.SaveStoreSnapshot(ctx, StoreSnapshot{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("SaveStoreSnapshot canceled error = %v", err)
	}
}

func TestFileStoreSnapshotSaveReportsFilesystemError(t *testing.T) {
	dir := t.TempDir()
	parentFile := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(parentFile, []byte("x"), 0o600); err != nil {
		t.Fatalf("write parent file: %v", err)
	}
	store, err := NewFileStoreSnapshot(filepath.Join(parentFile, "snapshot.json"))
	if err != nil {
		t.Fatalf("NewFileStoreSnapshot: %v", err)
	}
	err = store.SaveStoreSnapshot(context.Background(), StoreSnapshot{})
	if err == nil {
		t.Fatal("expected filesystem save error")
	}
}

func newFileStoreSnapshotForBoundary(t *testing.T) *FileStoreSnapshot {
	t.Helper()
	store, err := NewFileStoreSnapshot(filepath.Join(t.TempDir(), "snapshot.json"))
	if err != nil {
		t.Fatalf("NewFileStoreSnapshot: %v", err)
	}
	return store
}
