package riidoaiserver

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func (s *Store) send(ctx context.Context, cmd any) error {
	select {
	case s.commands <- cmd:
		return nil
	case <-s.done:
		return errors.New("riido-control-plane store closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Store) startStoreOperationTrace(ctx context.Context, operation StoreOperationName) (context.Context, TraceSpan) {
	return StartTraceSpan(ctx, s.traceRecorder, TraceSpanStart{
		Name: "store." + operation.String(),
		Kind: TraceSpanKindInternal,
		Attributes: []TraceAttribute{
			StringTraceAttribute(metadatakeys.RiidoStoreOperation.String(), operation.String()),
			StringTraceAttribute(metadatakeys.RiidoTraceSurface.String(), "assignment_store"),
		},
	})
}
