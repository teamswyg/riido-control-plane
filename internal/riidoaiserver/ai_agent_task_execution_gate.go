package riidoaiserver

import (
	"context"
	"sync"
)

// Publication and membership removal share a task boundary, without executing
// client-store persistence inside the assignment actor.
type taskAgentExecutionGate interface {
	beginPublication(context.Context, string) (func(), error)
	beginRemoval(context.Context, string) (func(), error)
}

type taskAgentExecutionGates struct {
	mu    sync.Mutex
	tasks map[string]*taskAgentExecutionGateEntry
}

type taskAgentExecutionGateEntry struct {
	token chan struct{}
	users int
}

func newTaskAgentExecutionGates() *taskAgentExecutionGates {
	return &taskAgentExecutionGates{tasks: map[string]*taskAgentExecutionGateEntry{}}
}

func (g *taskAgentExecutionGates) beginPublication(ctx context.Context, taskID string) (func(), error) {
	return g.acquire(ctx, taskID)
}

func (g *taskAgentExecutionGates) beginRemoval(ctx context.Context, taskID string) (func(), error) {
	return g.acquire(ctx, taskID)
}

func (g *taskAgentExecutionGates) acquire(ctx context.Context, taskID string) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	g.mu.Lock()
	entry := g.tasks[taskID]
	if entry == nil {
		entry = &taskAgentExecutionGateEntry{token: make(chan struct{}, 1)}
		entry.token <- struct{}{}
		g.tasks[taskID] = entry
	}
	entry.users++
	g.mu.Unlock()
	select {
	case <-ctx.Done():
		g.releaseUser(taskID, entry)
		return nil, ctx.Err()
	case <-entry.token:
		return func() { entry.token <- struct{}{}; g.releaseUser(taskID, entry) }, nil
	}
}

func (g *taskAgentExecutionGates) releaseUser(taskID string, entry *taskAgentExecutionGateEntry) {
	g.mu.Lock()
	defer g.mu.Unlock()
	entry.users--
	if entry.users == 0 {
		delete(g.tasks, taskID)
	}
}
