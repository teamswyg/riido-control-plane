package riidoaiserver

import (
	"strings"
	"time"
)

// aiAgentClientAssistantPartialKey is the structured progress key the daemon
// attaches to streamed assistant body deltas. It mirrors the client
// ASSISTANT_PARTIAL_KEY so the web UI can render the live assistant body and so
// the control plane can reconstruct a completed answer from the progress lines.
const aiAgentClientAssistantPartialKey = "assistant.partial"

// aiAgentClientThreadMessagesPerThreadLimit caps the preserved conversation
// history per thread so the snapshot item stays under DynamoDB's 400 KB limit.
// Older turns beyond this bound are dropped (cursor pagination is future work).
const aiAgentClientThreadMessagesPerThreadLimit = 50

// threadRunBody accumulates the current run's assistant body from streamed
// assistant.partial deltas. It is transient (never persisted): a completed run's
// body is archived into taskThreadMessages at completion, and THAT is persisted.
type threadRunBody struct {
	runID string
	body  string
}

// accumulateAssistantRunBodyLocked appends one streamed assistant delta to the
// current run's body buffer for a thread, resetting when the run changes. The
// thread.Lines feed accumulates across runs, so this per-run buffer is what lets
// archiving capture exactly one run's answer.
func (s *DevelopmentAIAgentClientStore) accumulateAssistantRunBodyLocked(threadID, runID, delta string) {
	threadID = strings.TrimSpace(threadID)
	if threadID == "" || delta == "" {
		return
	}
	if s.taskThreadRunBodies == nil {
		s.taskThreadRunBodies = map[string]threadRunBody{}
	}
	current := s.taskThreadRunBodies[threadID]
	if current.runID != runID {
		current = threadRunBody{runID: runID}
	}
	current.body += delta
	s.taskThreadRunBodies[threadID] = current
}

// assistantRunBodyLocked returns the best available assistant answer body for a
// thread's current run: the live run accumulator, else the concatenated
// assistant.partial progress lines, else the scalar status Message.
func (s *DevelopmentAIAgentClientStore) assistantRunBodyLocked(thread AIAgentTaskThreadRecord) string {
	threadID := strings.TrimSpace(thread.ThreadID)
	runID := strings.TrimSpace(thread.RunID)
	if buf, ok := s.taskThreadRunBodies[threadID]; ok && buf.runID == runID {
		if body := clientVisibleTaskThreadText(buf.body); body != "" {
			return body
		}
	}
	var b strings.Builder
	for _, line := range thread.Lines {
		if line.MessageKey == aiAgentClientAssistantPartialKey {
			b.WriteString(line.Message)
		}
	}
	if body := clientVisibleTaskThreadText(b.String()); body != "" {
		return body
	}
	return clientVisibleTaskThreadText(thread.Message)
}

// archiveThreadAssistantMessageLocked preserves the thread's current assistant
// answer as an append-only history message BEFORE any path overwrites
// thread.Message or stops the run. Idempotent per (thread, run): a duplicate
// completion/guard call does not double-append.
func (s *DevelopmentAIAgentClientStore) archiveThreadAssistantMessageLocked(thread AIAgentTaskThreadRecord, now time.Time) {
	threadID := strings.TrimSpace(thread.ThreadID)
	if threadID == "" {
		return
	}
	body := s.assistantRunBodyLocked(thread)
	if body == "" {
		return
	}
	messageID := "assistant:" + threadID + ":" + strings.TrimSpace(thread.RunID)
	if s.threadMessageExistsLocked(threadID, messageID) {
		return
	}
	s.appendThreadMessageLocked(threadID, AIAgentTaskThreadMessageRecord{
		MessageID:  messageID,
		ThreadID:   threadID,
		Role:       "assistant",
		AuthorType: "agent",
		AgentID:    strings.TrimSpace(thread.AgentID),
		Body:       body,
		RunID:      strings.TrimSpace(thread.RunID),
		CreatedAt:  now.UTC(),
	})
}

// archiveThreadUserMessageLocked records a user follow-up turn so the rendered
// conversation alternates user/assistant. Idempotent per source message (or per
// run when no source id is supplied).
func (s *DevelopmentAIAgentClientStore) archiveThreadUserMessageLocked(thread AIAgentTaskThreadRecord, body, sourceMessageID string, now time.Time) {
	threadID := strings.TrimSpace(thread.ThreadID)
	body = strings.TrimSpace(body)
	sourceMessageID = strings.TrimSpace(sourceMessageID)
	if threadID == "" || body == "" {
		return
	}
	discriminator := sourceMessageID
	if discriminator == "" {
		discriminator = strings.TrimSpace(thread.RunID)
	}
	messageID := "user:" + threadID + ":" + discriminator
	if s.threadMessageExistsLocked(threadID, messageID) {
		return
	}
	s.appendThreadMessageLocked(threadID, AIAgentTaskThreadMessageRecord{
		MessageID:       messageID,
		ThreadID:        threadID,
		Role:            "user",
		AuthorType:      "user",
		Body:            clientVisibleTaskThreadText(body),
		SourceMessageID: sourceMessageID,
		CreatedAt:       now.UTC(),
	})
}

func (s *DevelopmentAIAgentClientStore) threadMessageExistsLocked(threadID, messageID string) bool {
	for _, m := range s.taskThreadMessages[threadID] {
		if m.MessageID == messageID {
			return true
		}
	}
	return false
}

func (s *DevelopmentAIAgentClientStore) appendThreadMessageLocked(threadID string, message AIAgentTaskThreadMessageRecord) {
	if s.taskThreadMessages == nil {
		s.taskThreadMessages = map[string][]AIAgentTaskThreadMessageRecord{}
	}
	messages := append(s.taskThreadMessages[threadID], message)
	if len(messages) > aiAgentClientThreadMessagesPerThreadLimit {
		messages = messages[len(messages)-aiAgentClientThreadMessagesPerThreadLimit:]
	}
	s.taskThreadMessages[threadID] = messages
}

// hydrateThreadMessagesLocked returns the client-visible conversation history for
// a thread (sanitized), to attach to the read response as thread.messages. The
// history is stored separately from the thread record and is never written into
// the task_threads snapshot.
func (s *DevelopmentAIAgentClientStore) hydrateThreadMessagesLocked(threadID string) []AIAgentTaskThreadMessageRecord {
	source := s.taskThreadMessages[strings.TrimSpace(threadID)]
	if len(source) == 0 {
		return nil
	}
	out := make([]AIAgentTaskThreadMessageRecord, len(source))
	for i, m := range source {
		m.Body = clientVisibleTaskThreadText(m.Body)
		out[i] = m
	}
	return out
}
