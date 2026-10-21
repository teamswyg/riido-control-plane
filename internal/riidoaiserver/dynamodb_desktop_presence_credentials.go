package riidoaiserver

import "context"

// Refresh once outside the mutex so another poll can cancel while credentials
// are being fetched. Cached credentials keep the existing expiry safety margin.
func (s *DynamoDBAIAgentClientSnapshot) desktopPresenceCredentials(ctx context.Context) (AWSCredentials, error) {
	for {
		s.grantsReadMu.Lock()
		if pending := s.grantsReadPending; pending != nil {
			s.grantsReadMu.Unlock()
			select {
			case <-pending:
				continue
			case <-ctx.Done():
				return AWSCredentials{}, ctx.Err()
			case <-s.done:
				return AWSCredentials{}, errDynamoDBAIAgentClientSnapshotClosed()
			}
		}
		pending := make(chan struct{})
		s.grantsReadPending = pending
		cached := s.grantsReadCredentials
		s.grantsReadMu.Unlock()
		credentials, err := cachedAWSCredentials(ctx, s.now, s.credentialsProvider, &cached)
		s.grantsReadMu.Lock()
		if err == nil {
			s.grantsReadCredentials = cached
		}
		s.grantsReadPending = nil
		close(pending)
		s.grantsReadMu.Unlock()
		return credentials, err
	}
}
