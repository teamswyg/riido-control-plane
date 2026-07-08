package riidoaiserver

import "testing"

func TestDropTaskThreadProgressCacheLockedDropsHistoryWithoutProgressCache(t *testing.T) {
	store := &DevelopmentAIAgentClientStore{
		taskThreadHistoryCache: map[string]taskThreadHistoryMessageCache{
			"thread-a": {},
			"thread-b": {},
		},
	}

	store.dropTaskThreadProgressCacheLocked("thread-a")

	if _, ok := store.taskThreadHistoryCache["thread-a"]; ok {
		t.Fatal("thread-a history cache was not dropped")
	}
	if _, ok := store.taskThreadHistoryCache["thread-b"]; !ok {
		t.Fatal("unrelated history cache was dropped")
	}
}

func TestDropTaskThreadProgressCacheLockedDropsProgressAndHistory(t *testing.T) {
	store := &DevelopmentAIAgentClientStore{
		taskThreadProgressCache: map[string]taskThreadProgressMessageCache{
			"thread-a": {},
			"thread-b": {},
		},
		taskThreadHistoryCache: map[string]taskThreadHistoryMessageCache{
			"thread-a": {},
			"thread-b": {},
		},
	}

	store.dropTaskThreadProgressCacheLocked("thread-a")

	if _, ok := store.taskThreadProgressCache["thread-a"]; ok {
		t.Fatal("thread-a progress cache was not dropped")
	}
	if _, ok := store.taskThreadHistoryCache["thread-a"]; ok {
		t.Fatal("thread-a history cache was not dropped")
	}
	if _, ok := store.taskThreadProgressCache["thread-b"]; !ok {
		t.Fatal("unrelated progress cache was dropped")
	}
	if _, ok := store.taskThreadHistoryCache["thread-b"]; !ok {
		t.Fatal("unrelated history cache was dropped")
	}
}
