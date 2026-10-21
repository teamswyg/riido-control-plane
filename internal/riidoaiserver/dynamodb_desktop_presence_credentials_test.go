package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

type presenceCredentialsFunc func(context.Context) (AWSCredentials, error)

func (f presenceCredentialsFunc) Credentials(ctx context.Context) (AWSCredentials, error) {
	return f(ctx)
}

func TestDesktopPresenceCredentialWaitHonorsCancellation(t *testing.T) {
	entered := make(chan struct{})
	store := &DynamoDBAIAgentClientSnapshot{done: make(chan struct{}), credentialsProvider: presenceCredentialsFunc(func(ctx context.Context) (AWSCredentials, error) {
		close(entered)
		<-ctx.Done()
		return AWSCredentials{}, ctx.Err()
	})}
	first, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { _, err := store.desktopPresenceCredentials(first); result <- err }()
	<-entered
	waiting, stop := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer stop()
	waitResult := make(chan error, 1)
	go func() { _, err := store.desktopPresenceCredentials(waiting); waitResult <- err }()
	select {
	case err := <-waitResult:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("credential waiter lost deadline")
		}
	case <-time.After(time.Second):
		t.Fatal("cancelled poll blocked on credential refresh")
	}
	cancel()
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatal("refresh lost cancellation")
	}
}

func TestDesktopPresenceRefreshFailureDoesNotUseExpiredCache(t *testing.T) {
	now := fixedSnapshotTestNow()
	calls := 0
	store := &DynamoDBAIAgentClientSnapshot{done: make(chan struct{}), now: func() time.Time { return now }}
	store.credentialsProvider = presenceCredentialsFunc(func(context.Context) (AWSCredentials, error) {
		calls++
		if calls > 1 {
			return AWSCredentials{}, errors.New("private credential failure")
		}
		return AWSCredentials{AccessKeyID: "AKID", SecretAccessKey: "SECRET", ExpiresAt: now.Add(10 * time.Minute)}, nil
	})
	for range 2 {
		if _, err := store.desktopPresenceCredentials(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 1 {
		t.Fatal("valid cached credentials fetched repeatedly")
	}
	now = now.Add(6 * time.Minute)
	if _, err := store.desktopPresenceCredentials(context.Background()); err == nil {
		t.Fatal("refresh failure silently reused expiring credentials")
	}
	if calls != 2 {
		t.Fatal("expiry safety margin did not refresh credentials")
	}
}
