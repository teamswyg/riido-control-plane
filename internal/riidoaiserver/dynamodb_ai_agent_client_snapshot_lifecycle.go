package riidoaiserver

import (
	"context"
	"errors"
)

func (s *DynamoDBAIAgentClientSnapshot) LoadAIAgentClientSnapshot(ctx context.Context) (AIAgentClientSnapshot, bool, error) {
	if s == nil {
		return AIAgentClientSnapshot{}, false, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	reply := make(chan dynamoDBAIAgentClientSnapshotLoadResult, 1)
	select {
	case s.commands <- dynamoDBAIAgentClientSnapshotCommand{ctx: ctx, load: true, loadDone: reply}:
	case <-s.done:
		return AIAgentClientSnapshot{}, false, errDynamoDBAIAgentClientSnapshotClosed()
	case <-ctx.Done():
		return AIAgentClientSnapshot{}, false, ctx.Err()
	}
	return s.waitLoadReply(ctx, reply)
}

func (s *DynamoDBAIAgentClientSnapshot) SaveAIAgentClientSnapshot(ctx context.Context, snapshot AIAgentClientSnapshot) error {
	if s == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	reply := make(chan error, 1)
	snapshotCopy := snapshot
	select {
	case s.commands <- dynamoDBAIAgentClientSnapshotCommand{ctx: ctx, save: &snapshotCopy, errDone: reply}:
	case <-s.done:
		return errDynamoDBAIAgentClientSnapshotClosed()
	case <-ctx.Done():
		return ctx.Err()
	}
	return s.waitErrorReply(ctx, reply)
}

func (s *DynamoDBAIAgentClientSnapshot) Close() error {
	if s == nil {
		return nil
	}
	reply := make(chan error, 1)
	select {
	case s.commands <- dynamoDBAIAgentClientSnapshotCommand{close: true, errDone: reply}:
		return <-reply
	case <-s.done:
		return nil
	}
}

func errDynamoDBAIAgentClientSnapshotClosed() error {
	return errors.New("riidoaiserver: DynamoDB AI Agent client snapshot store closed")
}
