package main

import "testing"

func TestLoopRefreshDispatchRunsAfterCommandProducer(t *testing.T) {
	root := repoRootForTest(t)
	if err := requireRefreshDispatchAfterProducer(root); err != nil {
		t.Fatal(err)
	}
}

func TestRequireRefreshDispatchAfterProducerRejectsEarlyDispatch(t *testing.T) {
	err := requireRefreshDispatchOrder(1200, 1199)
	if err == nil {
		t.Fatal("expected dispatch-before-producer rejection")
	}
}

func TestDailyCronMinuteRejectsUnsupportedCron(t *testing.T) {
	for _, expr := range []string{
		"17 20 * * 1",
		"*/5 * * * *",
		"17 bad * * *",
	} {
		if _, err := dailyCronMinute(expr); err == nil {
			t.Fatalf("expected unsupported cron rejection for %q", expr)
		}
	}
}
