package riidoaiserver

import "context"

type ProviderStatusStore interface {
	SyncProviderStatus(ctx context.Context, agentID string, req ProviderStatusSyncRequest) (ProviderStatusSyncResponse, error)
}

type ProviderStatusReader interface {
	GetProviderStatus(ctx context.Context, agentID string) (ProviderStatusSyncResponse, bool, error)
}
