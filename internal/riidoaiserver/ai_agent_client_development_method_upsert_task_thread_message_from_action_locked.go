package riidoaiserver

import "time"

func (s *DevelopmentAIAgentClientStore) upsertTaskThreadMessageFromActionLocked(
	response AIAgentTaskActionResponse,
	sourceMessageID string,
	conversationID string,
	parentThreadID string,
) {
	now := time.Now().UTC()
	thread := taskThreadMessageFromAction(response, sourceMessageID, conversationID, parentThreadID, now)
	if !taskThreadHasActiveStream(thread) {
		thread.CompletedAt = now
	}
	s.ensureTaskThreadAgentSnapshotLocked(&thread, now)
	threads := s.taskThreads[response.TaskID]
	for i := range threads {
		if threads[i].ThreadID != response.ThreadID {
			continue
		}
		s.updateTaskThreadMessageFromActionLocked(&threads[i], response, sourceMessageID, conversationID, parentThreadID, now)
		s.taskThreads[response.TaskID] = threads
		return
	}
	s.taskThreads[response.TaskID] = append(threads, thread)
}
