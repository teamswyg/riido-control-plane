package riidoaiserver

import "context"

func (s *Store) SubscribeTask(ctx context.Context, taskID string) ([]TaskEvent, <-chan TaskEvent, func(), error) {
	reply := make(chan subscribeResult, 1)
	if err := s.send(ctx, subscribeCmd{taskID: taskID, reply: reply}); err != nil {
		return nil, nil, nil, err
	}
	select {
	case res := <-reply:
		if res.err != nil {
			return nil, nil, nil, res.err
		}
		return res.history, res.events, s.unsubscribeTaskFunc(taskID, res.subID), nil
	case <-ctx.Done():
		return nil, nil, nil, ctx.Err()
	}
}

func (s *Store) unsubscribeTaskFunc(taskID string, subID int64) func() {
	return func() {
		unsub := make(chan struct{}, 1)
		if err := s.send(context.Background(), unsubscribeCmd{taskID: taskID, subID: subID, reply: unsub}); err != nil {
			return
		}
		<-unsub
	}
}
