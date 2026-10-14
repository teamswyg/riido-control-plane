package riidoaiserver

import "time"

type taskThreadHistoryMessageCache struct {
	assignmentID    string
	runID           string
	assignmentState AgentAssignmentState
	commentKind     AgentTaskCommentKind
	message         string
	resultMessage   string
	lineCount       int
	lastSeq         int
	lastMessage     string
	lastObserved    time.Time
	startedAt       time.Time
	completedAt     time.Time
	messages        []AIAgentTaskThreadHistoryMessage
}

func (s *DevelopmentAIAgentClientStore) cachedTaskThreadHistoryMessagesLocked(
	thread AIAgentTaskThreadRecord,
) []AIAgentTaskThreadHistoryMessage {
	if s.taskThreadHistoryCache == nil {
		s.taskThreadHistoryCache = map[string]taskThreadHistoryMessageCache{}
	}
	if cached, ok := s.taskThreadHistoryCache[thread.ThreadID]; ok && cached.matches(thread) {
		return cached.messages
	}
	messages := buildTaskThreadHistoryMessages(nil, s.cachedTaskThreadProgressMessagesLocked(thread))
	if message, ok := taskThreadProjectionMessage(thread); ok {
		messages = append(messages, message)
	}
	sortTaskThreadHistoryMessages(messages)
	s.taskThreadHistoryCache[thread.ThreadID] = newTaskThreadHistoryMessageCache(thread, messages)
	return messages
}

func newTaskThreadHistoryMessageCache(
	thread AIAgentTaskThreadRecord,
	messages []AIAgentTaskThreadHistoryMessage,
) taskThreadHistoryMessageCache {
	lastSeq, lastMessage, lastObserved := lastTaskThreadProgressLineState(thread)
	return taskThreadHistoryMessageCache{
		assignmentID: thread.AssignmentID, runID: thread.RunID,
		assignmentState: thread.AssignmentState, commentKind: thread.CommentKind,
		message: thread.Message, resultMessage: thread.ResultMessage,
		lineCount: len(thread.Lines), lastSeq: lastSeq, lastMessage: lastMessage,
		lastObserved: lastObserved, startedAt: thread.StartedAt, completedAt: thread.CompletedAt,
		messages: messages,
	}
}

func (cached taskThreadHistoryMessageCache) matches(thread AIAgentTaskThreadRecord) bool {
	lastSeq, lastMessage, lastObserved := lastTaskThreadProgressLineState(thread)
	return cached.assignmentID == thread.AssignmentID && cached.runID == thread.RunID &&
		cached.assignmentState == thread.AssignmentState && cached.commentKind == thread.CommentKind &&
		cached.message == thread.Message && cached.resultMessage == thread.ResultMessage &&
		cached.lineCount == len(thread.Lines) && cached.lastSeq == lastSeq &&
		cached.lastMessage == lastMessage && cached.lastObserved.Equal(lastObserved) &&
		cached.startedAt.Equal(thread.StartedAt) && cached.completedAt.Equal(thread.CompletedAt)
}

func lastTaskThreadProgressLineState(thread AIAgentTaskThreadRecord) (int, string, time.Time) {
	if len(thread.Lines) == 0 {
		return 0, "", time.Time{}
	}
	last := thread.Lines[len(thread.Lines)-1]
	return last.Seq, last.Message, last.ObservedAt
}
