package riidoaiserver

import "time"

type taskThreadProgressMessageCache struct {
	assignmentID string
	runID        string
	lineCount    int
	lastSeq      int
	lastMessage  string
	lastObserved time.Time
	visibleLines []AIAgentTaskThreadHistoryMessage
}

func (s *DevelopmentAIAgentClientStore) cachedTaskThreadProgressMessagesLocked(
	thread AIAgentTaskThreadRecord,
) []AIAgentTaskThreadHistoryMessage {
	if len(thread.Lines) == 0 {
		return nil
	}
	if s.taskThreadProgressCache == nil {
		s.taskThreadProgressCache = map[string]taskThreadProgressMessageCache{}
	}
	if cached, ok := s.taskThreadProgressCache[thread.ThreadID]; ok && cached.matches(thread) {
		return cached.visibleLines
	}
	messages := taskThreadProgressMessages(thread)
	s.taskThreadProgressCache[thread.ThreadID] = newTaskThreadProgressMessageCache(thread, messages)
	return messages
}

func (s *DevelopmentAIAgentClientStore) dropTaskThreadProgressCacheLocked(threadID string) {
	if s.taskThreadProgressCache == nil {
		s.dropTaskThreadHistoryCacheLocked(threadID)
		return
	}
	delete(s.taskThreadProgressCache, threadID)
	s.dropTaskThreadHistoryCacheLocked(threadID)
}

func (s *DevelopmentAIAgentClientStore) dropTaskThreadHistoryCacheLocked(threadID string) {
	if s.taskThreadHistoryCache == nil {
		return
	}
	delete(s.taskThreadHistoryCache, threadID)
}

func newTaskThreadProgressMessageCache(
	thread AIAgentTaskThreadRecord,
	messages []AIAgentTaskThreadHistoryMessage,
) taskThreadProgressMessageCache {
	last := thread.Lines[len(thread.Lines)-1]
	return taskThreadProgressMessageCache{
		assignmentID: thread.AssignmentID,
		runID:        thread.RunID,
		lineCount:    len(thread.Lines),
		lastSeq:      last.Seq,
		lastMessage:  last.Message,
		lastObserved: last.ObservedAt,
		visibleLines: messages,
	}
}

func (cached taskThreadProgressMessageCache) matches(thread AIAgentTaskThreadRecord) bool {
	if len(thread.Lines) == 0 {
		return false
	}
	last := thread.Lines[len(thread.Lines)-1]
	return cached.assignmentID == thread.AssignmentID &&
		cached.runID == thread.RunID &&
		cached.lineCount == len(thread.Lines) &&
		cached.lastSeq == last.Seq &&
		cached.lastMessage == last.Message &&
		cached.lastObserved.Equal(last.ObservedAt)
}
