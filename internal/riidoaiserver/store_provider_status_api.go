package riidoaiserver

import "context"

func (s *Store) SyncProviderStatus(ctx context.Context, agentID string, req ProviderStatusSyncRequest) (ProviderStatusSyncResponse, error) {
	reply := make(chan providerStatusResult, 1)
	if err := s.send(ctx, providerStatusCmd{agentID: agentID, req: req, reply: reply}); err != nil {
		return ProviderStatusSyncResponse{}, err
	}
	select {
	case res := <-reply:
		return res.response, res.err
	case <-ctx.Done():
		return ProviderStatusSyncResponse{}, ctx.Err()
	}
}

func (s *Store) GetProviderStatus(ctx context.Context, agentID string) (ProviderStatusSyncResponse, bool, error) {
	reply := make(chan getProviderStatusResult, 1)
	if err := s.send(ctx, getProviderStatusCmd{agentID: agentID, reply: reply}); err != nil {
		return ProviderStatusSyncResponse{}, false, err
	}
	select {
	case res := <-reply:
		return res.response, res.ok, res.err
	case <-ctx.Done():
		return ProviderStatusSyncResponse{}, false, ctx.Err()
	}
}
