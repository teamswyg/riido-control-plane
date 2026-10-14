package main

import (
	"strings"
	"testing"
)

func TestRenderListAndLineBudgetEdges(t *testing.T) {
	t.Parallel()
	var b strings.Builder
	renderList(&b, "Empty", nil)
	if b.Len() != 0 {
		t.Fatalf("empty list rendered: %q", b.String())
	}
	renderLineBudgetRatchet(&b, lineBudgetResult{})
	if b.Len() != 0 {
		t.Fatalf("empty ratchet rendered: %q", b.String())
	}
	if got := lineBudgetFilesSlack(lineBudgetResult{}); got != 0 {
		t.Fatalf("files slack without limit = %d", got)
	}
	if got := lineBudgetMaxLinesSlack(lineBudgetResult{}); got != 0 {
		t.Fatalf("max lines slack without limit = %d", got)
	}
	renderLineBudget(&b, lineBudgetResult{Target: 5})
	if !strings.Contains(b.String(), "files over target") {
		t.Fatalf("expected line budget summary: %q", b.String())
	}
}
