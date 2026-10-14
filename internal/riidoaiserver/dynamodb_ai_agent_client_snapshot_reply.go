package riidoaiserver

import "context"

func (s *DynamoDBAIAgentClientSnapshot) waitLoadReply(ctx context.Context, reply <-chan dynamoDBAIAgentClientSnapshotLoadResult) (AIAgentClientSnapshot, bool, error) {
	select {
	case result := <-reply:
		return result.snapshot, result.ok, result.err
	case <-s.done:
		return AIAgentClientSnapshot{}, false, errDynamoDBAIAgentClientSnapshotClosed()
	case <-ctx.Done():
		return AIAgentClientSnapshot{}, false, ctx.Err()
	}
}

func (s *DynamoDBAIAgentClientSnapshot) waitErrorReply(ctx context.Context, reply <-chan error) error {
	select {
	case err := <-reply:
		return err
	case <-s.done:
		return errDynamoDBAIAgentClientSnapshotClosed()
	case <-ctx.Done():
		return ctx.Err()
	}
}
