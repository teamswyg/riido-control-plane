package riidoaiserver

import (
	"strings"
	"sync"
	"time"
)

type aiAgentGlobalReconcileGate struct {
	mu       sync.Mutex
	interval time.Duration
	last     map[string]time.Time
}

func newAIAgentGlobalReconcileGate(interval time.Duration) *aiAgentGlobalReconcileGate {
	if interval <= 0 {
		interval = time.Second
	}
	return &aiAgentGlobalReconcileGate{
		interval: interval,
		last:     map[string]time.Time{},
	}
}

func (g *aiAgentGlobalReconcileGate) reserve(principal AuthorizationResult, now time.Time) (string, bool) {
	if g == nil {
		return "", true
	}
	key := aiAgentGlobalReconcileKey(principal)
	g.mu.Lock()
	defer g.mu.Unlock()
	g.pruneLocked(now)
	if last := g.last[key]; !last.IsZero() && now.Before(last.Add(g.interval)) {
		return key, false
	}
	g.last[key] = now
	return key, true
}

func (g *aiAgentGlobalReconcileGate) forget(key string) {
	if g == nil || strings.TrimSpace(key) == "" {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.last, key)
}

func (g *aiAgentGlobalReconcileGate) pruneLocked(now time.Time) {
	for key, last := range g.last {
		if !now.Before(last.Add(2 * g.interval)) {
			delete(g.last, key)
		}
	}
}

func aiAgentGlobalReconcileKey(principal AuthorizationResult) string {
	return strings.TrimSpace(principal.WorkspaceID) + "\x00" + strings.TrimSpace(principal.PrincipalID)
}
