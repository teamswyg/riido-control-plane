package riidoaiserver

import (
	"context"
	"testing"
)

func TestMemoryAssignmentContextQuotaCancelledRequestCannotConsume(t *testing.T) {
	m := NewMemoryAssignmentContextStore()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if ok, err := m.ConsumeAssignmentContextRead(ctx, "assignment"); ok || err == nil {
		t.Fatalf("cancelled request allowed=%v err=%v", ok, err)
	}
	for i := range 100 {
		if ok, err := m.ConsumeAssignmentContextRead(t.Context(), "assignment"); !ok || err != nil {
			t.Fatalf("attempt=%d allowed=%v err=%v", i, ok, err)
		}
	}
}
