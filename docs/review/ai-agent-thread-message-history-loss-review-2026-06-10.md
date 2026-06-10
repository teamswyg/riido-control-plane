# AI Agent thread message history loss review (2026-06-10)

> Target: `riido-control-plane`, `riido-client`
> Status: root-cause review + fix plan
>
> **UPDATE 2026-06-10 — implementation plan moved.** The root-cause analysis
> below is accurate, but the "Fast Fix Plan" (inline `Messages[]`) is
> **superseded** by the locked decisions in
> [`ai-agent-thread-message-history-fix-plan-2026-06-10.md`](ai-agent-thread-message-history-fix-plan-2026-06-10.md):
> a **separate `thread_messages` collection** hydrated onto `thread.messages`, not
> an inline field. **Correction:** `assistant.partial` is **dead today**, not
> "supported" — the daemon never tags body deltas and
> `AgentThreadProgressLine.MarshalJSON` drops `message_key` on the wire, so
> repairing the live stream needs daemon tagging **and** a MarshalJSON change. The
> "Evidence for `assistant.partial` support" lines below (api.go:482/486) only
> point at a generic optional field.

## Summary

Agent replies disappear when a new follow-up reply is added to the same thread or
when the task participant agent is changed.

This is not primarily a rendering bug. The current control-plane client read
model uses a task thread as one mutable "current execution state" record, not as
an append-only conversation history. Because the previous agent response is only
kept in `thread.message`/`thread.lines`, later actions overwrite the only
visible fallback body.

## Confirmed Root Cause

### 1. Thread records do not have persisted message history

`AIAgentTaskThreadRecord` currently stores the thread-level execution state and
only these message-ish fields:

- `Message`
- `SourceCommentID`
- `SourceMessageID`
- `Lines`

There is no server-side `Messages []...` or `LastMessage` history field in the
Go read model.

Evidence:

- `internal/riidoaiserver/ai_agent_client_api.go:414`
- `internal/riidoaiserver/ai_agent_client_api.go:423`
- `internal/riidoaiserver/ai_agent_client_api.go:424`

The client, however, is already shaped as if the server may return message
history:

- `thread.messages`
- `thread.last_message`
- fallback to terminal `thread.message`

Evidence:

- `riido-client/src/lib/hooks/queries/aiAgent/useAiAgentTask.ts:126`
- `riido-client/src/lib/hooks/queries/aiAgent/useAiAgentTask.ts:146`
- `riido-client/src/components/domain/aiAgentTask/AgentThreadCard.tsx:282`

So the UI can render history, but the server never actually persists or returns
it.

### 2. Follow-up replies update the same thread instead of appending messages

The thread-message API creates a new assignment against the existing `thread_id`,
then updates the same thread record.

Evidence:

- `internal/riidoaiserver/server.go:1090`
- `internal/riidoaiserver/server.go:1095`
- `internal/riidoaiserver/server.go:1101`

The read-model update path finds the existing thread by `ThreadID` and
overwrites mutable fields:

- `RunID`
- `AssignmentID`
- `WorkStatus`
- `AssignmentState`
- `CommentKind`
- `Message`
- `SourceMessageID`

Evidence:

- `internal/riidoaiserver/ai_agent_client_development.go:2195`
- `internal/riidoaiserver/ai_agent_client_development.go:2215`
- `internal/riidoaiserver/ai_agent_client_development.go:2226`
- `internal/riidoaiserver/ai_agent_client_development.go:2228`

That means the previous assistant response is not archived anywhere before the
thread becomes the next execution.

### 3. The previous response is only visible as a fallback

`AgentThreadCard` renders `viewState.messages`. `viewState.messages` is built
from `thread.messages`, `thread.last_message`, and then, only for terminal
threads, the fallback `thread.message`.

Evidence:

- `riido-client/src/components/domain/aiAgentTask/AgentThreadCard.tsx:282`
- `riido-client/src/lib/hooks/queries/aiAgent/useAiAgentTask.ts:126`
- `riido-client/src/lib/hooks/queries/aiAgent/useAiAgentTask.ts:138`

Because the server does not send `messages[]`, the visible completed response is
usually just the fallback `thread.message`. When a follow-up starts, that field
is changed to a status/progress message for the new run. The prior response then
disappears from the rendered history.

### 4. Agent replacement also overwrites prior thread message

When a different agent is assigned, the old active thread is marked stopped and
its `Message` is overwritten with a participant-change/status message.

Evidence:

- `internal/riidoaiserver/ai_agent_client_development.go:816`
- `internal/riidoaiserver/ai_agent_client_development.go:817`
- `internal/riidoaiserver/ai_agent_client_development.go:2307`
- `internal/riidoaiserver/ai_agent_client_development.go:2317`

If the previous assistant response existed only as `thread.message`, this action
replaces it.

### 5. Current tests lock in the overwrite model

Existing tests expect a follow-up to keep one thread and update only the latest
`SourceMessageID`.

Evidence:

- `internal/riidoaiserver/ai_agent_client_http_test.go:1989`
- `internal/riidoaiserver/ai_agent_client_http_test.go:2034`
- `internal/riidoaiserver/ai_agent_client_http_test.go:2038`

There is no regression test asserting that previous agent responses survive a
follow-up or agent replacement.

## Fast Fix Plan (SUPERSEDED — see ai-agent-thread-message-history-fix-plan-2026-06-10.md; kept for history)

The fastest durable fix is to add append-only message history to the
control-plane read model and archive the current assistant response before any
thread-level overwrite.

### 1. Add server-side message records

Add a message history field to `AIAgentTaskThreadRecord`.

Suggested shape:

```go
type AIAgentTaskThreadMessageRecord struct {
	MessageID       string    `json:"message_id,omitempty"`
	SourceMessageID string    `json:"source_message_id,omitempty"`
	Role            string    `json:"role,omitempty"`
	AuthorType      string    `json:"author_type,omitempty"`
	Body            string    `json:"body,omitempty"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
}

type AIAgentTaskThreadRecord struct {
	// existing fields...
	Messages []AIAgentTaskThreadMessageRecord `json:"messages,omitempty"`
}
```

The client already has a compatible `TaskThreadMessageRecord` interface and
renders `thread.messages`.

Evidence:

- `riido-client/src/lib/apis/aiAgent/response.ts:172`
- `riido-client/src/lib/apis/aiAgent/response.ts:178`

### 2. Archive assistant output before overwriting thread state

Add a helper in `ai_agent_client_development.go` and call it before code paths
that overwrite `thread.Message` or replace the thread's active run fields.

Candidate helper:

```go
func archiveThreadAssistantMessageLocked(thread *AIAgentTaskThreadRecord, now time.Time) {
	body := latestAssistantPartialBody(thread.Lines)
	if body == "" {
		body = clientVisibleTaskThreadText(thread.Message)
	}
	if body == "" {
		return
	}
	id := "assistant-" + thread.ThreadID + "-" + thread.RunID
	for _, message := range thread.Messages {
		if message.MessageID == id || strings.TrimSpace(message.Body) == body {
			return
		}
	}
	thread.Messages = append(thread.Messages, AIAgentTaskThreadMessageRecord{
		MessageID:  id,
		Role:      "assistant",
		AuthorType: "agent",
		Body:      body,
		CreatedAt: firstNonZeroTime(thread.CompletedAt, thread.StartedAt, now),
	})
}
```

`latestAssistantPartialBody` should prefer the latest progress line whose
`MessageKey` is `assistant.partial`, because those lines carry the accumulated
assistant body.

Evidence for `assistant.partial` support:

- `internal/riidoaiserver/ai_agent_client_api.go:482`
- `internal/riidoaiserver/ai_agent_client_api.go:486`
- `riido-client/src/components/domain/aiAgentTask/AgentThreadCard.tsx:75`
- `riido-client/src/components/domain/aiAgentTask/AgentThreadCard.tsx:147`

Minimum overwrite call sites:

- `upsertTaskThreadMessageFromActionLocked` before updating an existing thread.
- `markTaskActiveThreadsStoppedLocked` before setting `threads[i].Message`.
- `markTaskAgentThreadsStoppedLocked` before setting `threads[i].Message`.
- `appendThreadProgressLocked` when transitioning or when a terminal event is
  projected, if terminal output is still only in lines.

### 3. Preserve and copy messages everywhere thread records are copied

Update copy/snapshot paths so `Messages` survives API responses and persistence:

- `copyTaskThread`
- `snapshot` / `restoreSnapshot`
- split snapshot task-thread DTOs if the split persistence work is active
- `retainLatestThreadProgressLines` should not delete message history

Evidence:

- `internal/riidoaiserver/ai_agent_client_development.go:3076`
- `internal/riidoaiserver/ai_agent_client_persistence.go:449`
- `internal/riidoaiserver/ai_agent_client_persistence.go:539`

### 4. Keep `Message` as status/summary, not conversation body

After this change:

- `thread.Message` should remain a status/summary/projection field.
- agent answers should be appended to `thread.Messages`.
- user follow-up source ids should not replace prior assistant output.

This avoids mixing "current execution status" with "conversation history".

## Optional Emergency Client Patch

If a server deployment cannot happen immediately, the client can temporarily
render the latest `assistant.partial` line from `thread.lines` as an additional
message fallback.

Patch area:

- `riido-client/src/lib/hooks/queries/aiAgent/useAiAgentTask.ts:126`

This is only a stopgap. It cannot reliably separate multiple runs because
`lines` is a progress stream, not a message history model. It also cannot
protect against server-side overwrites that remove or cap old progress lines.

## Required Tests

Add tests before or with the fix:

1. Completed agent response remains visible after posting a follow-up message to
   the same thread.
2. Completed agent response remains visible after replacing the task participant
   agent.
3. Follow-up can still resume the same provider session/thread context.
4. Snapshot save/restore preserves `Messages`.
5. Client renders multiple `thread.messages` in order and does not hide them as
   status messages.

Important existing tests that need updating:

- `internal/riidoaiserver/ai_agent_client_http_test.go:2034` currently expects
  a single thread and only checks latest `SourceMessageID`. Keep one thread if
  desired, but assert `Messages` contains the archived previous assistant body.

## Recommendation

Do not create a new task thread for every follow-up just to preserve display
history. The current resume flow depends on continuing the same thread identity.
Instead, keep one thread as the conversation container and make messages inside
that thread append-only.

The minimal correct model is:

- `thread`: task-level agent conversation container + current execution state
- `thread.messages`: append-only visible conversation history
- `thread.lines`: transient/progress replay tail
- `thread.message`: current status/summary fallback only
