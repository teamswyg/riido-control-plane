package riidoaiserver

import (
	"testing"
	"time"
)

func TestTaskThreadProgressCacheMatchesLineIdentity(t *testing.T) {
	observed := time.Date(2026, 7, 7, 1, 2, 3, 0, time.UTC)
	thread := AIAgentTaskThreadRecord{
		ThreadID: "thread-cache", AssignmentID: "asn-1", RunID: "run-1",
		Lines: []AgentThreadProgressLine{{
			Seq: 7, Message: "thinking", ObservedAt: observed,
		}},
	}
	cache := newTaskThreadProgressMessageCache(thread, []AIAgentTaskThreadHistoryMessage{{
		MessageID: "cached-progress",
	}})
	if !cache.matches(thread) {
		t.Fatalf("cache should match unchanged thread: %+v", cache)
	}

	changed := thread
	changed.Lines = []AgentThreadProgressLine{{
		Seq: 7, Message: "changed", ObservedAt: observed,
	}}
	if cache.matches(changed) {
		t.Fatalf("cache matched changed progress line: %+v", changed.Lines)
	}
	changed.Lines = nil
	if cache.matches(changed) {
		t.Fatal("cache matched a thread with no progress lines")
	}
}
