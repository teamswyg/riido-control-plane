package riidoaiserver

import "context"

func (s *Store) registerToolApprovalWaiter(ctx context.Context, key string) (<-chan struct{}, func(), error) {
	reply := make(chan registerToolApprovalWaiterResult, 1)
	if err := s.send(ctx, registerToolApprovalWaiterCmd{key: key, reply: reply}); err != nil {
		return nil, nil, err
	}
	var registered registerToolApprovalWaiterResult
	select {
	case registered = <-reply:
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}
	release := func() {
		done := make(chan struct{}, 1)
		cmd := unregisterToolApprovalWaiterCmd{key: key, id: registered.id, reply: done}
		if err := s.send(context.Background(), cmd); err == nil {
			<-done
		}
	}
	return registered.ch, release, nil
}
