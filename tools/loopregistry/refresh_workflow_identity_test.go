package main

import "testing"

func TestRefreshWorkflowMustDeclareLoopIdentity(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	loop := findLoopIndexForTest(t, m, "ai_thread_history")
	m.Loops[loop].ID = "missing_loop_identity"
	if _, err := verifyAll("../..", m, hashes); err == nil {
		t.Fatal("expected missing loop identity declaration to fail")
	}
}

func TestRefreshWorkflowLoopIDParser(t *testing.T) {
	text := "env:\n  RIIDO_LOOP_IDS: ai_thread_history closed_loop_candidate\n"
	ids := workflowLoopIDs(text)
	if !ids["ai_thread_history"] || !ids["closed_loop_candidate"] {
		t.Fatalf("loop ids = %+v", ids)
	}
}
