package riidoaiserver

import (
	"context"
	"strings"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

const unknownTaskContextOperation = "task_context_unknown"

type TaskContextOperationName string

const (
	TaskContextOperationResolve TaskContextOperationName = "task_context_resolve"
)

func (op TaskContextOperationName) String() string {
	value := strings.TrimSpace(string(op))
	if value == "" {
		return unknownTaskContextOperation
	}
	return value
}

func startTaskContextOperationTrace(ctx context.Context, operation TaskContextOperationName) (context.Context, TraceSpan) {
	return StartTraceSpan(ctx, nil, TraceSpanStart{
		Name: "task_context." + operation.String(),
		Kind: TraceSpanKindClient,
		Attributes: []TraceAttribute{
			StringTraceAttribute(metadatakeys.RiidoTaskContextOperation.String(), operation.String()),
			StringTraceAttribute(metadatakeys.RiidoTraceSurface.String(), "task_context"),
		},
	})
}
