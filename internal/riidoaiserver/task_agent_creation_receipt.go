package riidoaiserver

import (
	"context"
	"sync"
)

type taskAgentCreationContextKey struct{}

type taskAgentCreationReceipt struct {
	mu         sync.Mutex
	assignment Assignment
	created    bool
}

func withTaskAgentCreationReceipt(ctx context.Context) (context.Context, *taskAgentCreationReceipt) {
	receipt := &taskAgentCreationReceipt{}
	return context.WithValue(ctx, taskAgentCreationContextKey{}, receipt), receipt
}

func taskAgentCreationReceiptFrom(ctx context.Context) *taskAgentCreationReceipt {
	receipt, _ := ctx.Value(taskAgentCreationContextKey{}).(*taskAgentCreationReceipt)
	return receipt
}

func (r *taskAgentCreationReceipt) record(assignment Assignment, created bool) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.assignment, r.created = assignment, created
}

func (r *taskAgentCreationReceipt) ownCreation(assignment Assignment) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.created && sameTaskAgentCreation(r.assignment, assignment)
}
