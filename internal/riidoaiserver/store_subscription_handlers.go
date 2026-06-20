package riidoaiserver

import (
	"errors"
	"strings"
)

func (s *Store) handleSubscribe(state *storeState, taskID string) ([]TaskEvent, <-chan TaskEvent, int64, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, nil, 0, errors.New("task_id is required")
	}
	state.nextSubscriberSeq++
	subID := state.nextSubscriberSeq
	ch := make(chan TaskEvent, 32)
	if state.subscribers[taskID] == nil {
		state.subscribers[taskID] = map[int64]chan TaskEvent{}
	}
	state.subscribers[taskID][subID] = ch
	history := append([]TaskEvent(nil), state.events[taskID]...)
	return history, ch, subID, nil
}

func (s *Store) handleUnsubscribe(state *storeState, taskID string, subID int64) {
	subs := state.subscribers[taskID]
	if subs == nil {
		return
	}
	if ch, ok := subs[subID]; ok {
		close(ch)
		delete(subs, subID)
	}
	if len(subs) == 0 {
		delete(state.subscribers, taskID)
	}
}
