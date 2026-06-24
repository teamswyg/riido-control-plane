package riidoaiserver

import "context"

type fakeStreamRelayPublisher struct {
	events []StreamRelayEvent
}

func (p *fakeStreamRelayPublisher) PublishStreamRelayEvent(_ context.Context, event StreamRelayEvent) error {
	p.events = append(p.events, event)
	return nil
}

type fakeStreamRelayCheckpointStore struct {
	checkpoint StreamRelayCheckpoint
	ok         bool
	saved      []StreamRelayCheckpoint
}

func (s *fakeStreamRelayCheckpointStore) LoadStreamRelayCheckpoint(_ context.Context, streamARN, shardID string) (StreamRelayCheckpoint, bool, error) {
	if s.checkpoint.StreamARN != "" && s.checkpoint.StreamARN != streamARN {
		return StreamRelayCheckpoint{}, false, nil
	}
	if s.checkpoint.ShardID != "" && s.checkpoint.ShardID != shardID {
		return StreamRelayCheckpoint{}, false, nil
	}
	return s.checkpoint, s.ok, nil
}

func (s *fakeStreamRelayCheckpointStore) SaveStreamRelayCheckpoint(_ context.Context, checkpoint StreamRelayCheckpoint) error {
	s.saved = append(s.saved, checkpoint)
	return nil
}
