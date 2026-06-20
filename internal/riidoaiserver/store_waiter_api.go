package riidoaiserver

import "context"

func (s *Store) registerWaiter(ctx context.Context, agentID string) (<-chan struct{}, func(), error) {
	reply := make(chan registerWaiterResult, 1)
	if err := s.send(ctx, registerWaiterCmd{agentID: agentID, reply: reply}); err != nil {
		return nil, nil, err
	}
	select {
	case res := <-reply:
		release := func() {
			done := make(chan struct{}, 1)
			if err := s.send(context.Background(), unregisterWaiterCmd{agentID: agentID, id: res.id, reply: done}); err != nil {
				return
			}
			<-done
		}
		return res.ch, release, nil
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}
}
