# AI Agent thread message-history loss — fix plan (2026-06-10)

> Targets: `riido-contracts`, `riido-control-plane`, `riido-daemon`, `riido-client`
> Status: **implementation plan (decisions locked)**. Supersedes the "Fast Fix
> Plan / inline `Messages[]`" section of
> `ai-agent-thread-message-history-loss-review-2026-06-10.md`.
> Root-cause analysis: see that review (verified accurate on the core thesis;
> two corrections folded in below).

## Locked decisions

1. **Storage = separate `thread_messages` collection** (NOT inline on the thread
   record). The thread record is also the client DTO; we keep history physically
   separate server-side and **hydrate** it onto `thread.messages` only on the
   read path.
2. **Live stream = repair `assistant.partial`** end to end (it is dead today).
3. **Contracts = contract-first** (author DSL → regenerate IR/OpenAPI → sync the
   control-plane copy → then code).
4. **Archive = at completion + a guard before every overwrite** of `thread.Message`.

Scope is the broadest option chosen: durable assistant answers **and** user
follow-up turns appended to history, **plus** the `assistant.partial` live-stream
repair.

## Verified current state (exact anchors)

- `AIAgentTaskThreadRecord` (`internal/riidoaiserver/ai_agent_client_api.go:414-433`)
  is **both** the server record and the client DTO. It has a single scalar
  `Message string` (:429) + `Lines []AgentThreadProgressLine` (:432). **No**
  `Messages[]`/`LastMessage`.
- The body is overwritten on every thread-level action:
  - follow-up: `upsertTaskThreadMessageFromActionLocked` sets `threads[i].Message`
    to a placeholder (`ai_agent_client_development.go:2226`, fed by :988/:999).
  - agent replace / stop: `markTaskActiveThreadsStoppedLocked` (:2317),
    `markAgentTaskThreadsStoppedLocked` (:2341), `markTaskAgentThreadsStoppedLocked`
    (:2358) each set `threads[i].Message = <stop reason>`. **Three** sites.
  - progress: `appendThreadProgressLocked` (:2244-2289) appends to `Lines` and
    mirrors the last line into `Message` (:2266).
- The accumulated answer body lives only in `thread.Lines` (live memory) +
  the scalar `Message`. `Lines` are pruned to the last 200 and threads to 50/task
  at snapshot (`retainLatestThreadProgressLines`; `ai_agent_client_persistence.go`
  snapshot at :486, caps ~:515-533) — so **archiving must happen at completion in
  live memory**, not by reconstructing persisted `Lines` later.
- **Correction #1 to the review:** `assistant.partial` is dead, not "supported."
  - There is no `assistant.partial` in Go. The constant lives only in the client
    (`AgentThreadCard.tsx:75`).
  - The daemon never tags body deltas: `EventTextDelta` →
    `req.EventType=EventRiidoLog; req.Message=ev.Text` with **no metadata**
    (`riido-daemon .../saasplane/saasplane.go:889-898`).
  - Even if a line had `message_key`, `AgentThreadProgressLine.MarshalJSON`
    (`ai_agent_client_api.go:512-523`) emits **only** `seq/message/observed_at`
    and drops `message_key/message_code/message_args` — for **both** REST and SSE.
    `UnmarshalJSON` (:491-510) reads them, and the ingestion path
    `progressLineMetadata(event.Metadata)` → `line.MessageKey`
    (`ai_agent_client_development.go:1512,1523`) sets them on the in-memory line.
    So the **only** wire-drop is `MarshalJSON`.
  - Metadata key convention is shared: the daemon `ProgressEventMetadata`
    (`riido-daemon .../agentbridge/progress_messages.go:85-104`) emits
    `ProgressMessageMetadataKey` etc.; the server reads the same keys.
- **Correction #2:** persistence mostly rides for free, **but the split DTO
  converters enumerate fields explicitly**, so a new collection is NOT automatic:
  - `AIAgentClientSnapshot.TaskThreads` (`ai_agent_client_persistence.go:29`),
    `AIAgentClientThreadsSnapshot.TaskThreads`
    (`dynamodb_ai_agent_client_snapshot_split.go:51-55`), and `splitFromCombined`
    (:77-101) / `combinedFromSplit` (:106-122) all assign fields by name — a new
    `TaskThreadMessages` map must be added to **all** of them.
  - `restoreCorePreservingRest` (`ai_agent_client_persistence_split.go`) does not
    touch threads (preserved) — safe; just extend the "preserved" comment.
- Client is already shaped for this: `TaskThreadRecord.messages?`
  (`riido-client/src/lib/apis/aiAgent/response.ts:172`), `TaskThreadMessageRecord`
  (:178-189), and `AgentThreadProgressLine.message_key?` (:191-202, with a comment
  describing `assistant.partial`). So **no client type change is required**.
- Contract today: `AgentThreadProgressLine` in the DSL fixture
  (`riido-contracts/apicontract/fixtures/control-plane-ai-agent-client.dsl.riido.json:1952-1973`)
  has only `seq/message/observed_at` (matches MarshalJSON). The thread record is
  at :2100-2174, collection at :2179-2207. No `messages` / `message_key`.
- **Repo rule:** `riido-daemon` is **SSOT-document-first** (`AGENTS.md`): update the
  owning doc under `docs/` in the same PR before changing daemon behavior.
  Likely owner: `docs/20-domain/provider-runtime.md`.

## Target architecture

### Data model (server-internal, separate collection)

```go
// New, in ai_agent_client_api.go (and mirrored in the contract).
type AIAgentTaskThreadMessageRecord struct {
    MessageID       string    `json:"message_id"`
    ThreadID        string    `json:"thread_id"`
    Role            string    `json:"role"`        // "assistant" | "user"
    AuthorType      string    `json:"author_type"` // "agent" | "user"
    AgentID         string    `json:"agent_id,omitempty"`
    Body            string    `json:"body"`
    SourceMessageID string    `json:"source_message_id,omitempty"`
    RunID           string    `json:"run_id,omitempty"`
    CreatedAt       time.Time `json:"created_at"`
}

// New store field on DevelopmentAIAgentClientStore, keyed by ThreadID
// (thread ids are globally unique), separate from taskThreads.
taskThreadMessages map[string][]AIAgentTaskThreadMessageRecord
```

The thread DTO gains a **hydration-only** field:

```go
// AIAgentTaskThreadRecord:
Messages []AIAgentTaskThreadMessageRecord `json:"messages,omitempty"`
```

It is populated **only on the read path** and is empty in storage, so the
"separate collection" invariant holds (history is never written into the
`task_threads` snapshot — it lives in its own `task_thread_messages` map).

### Three correctness pillars

1. **Append idempotency.** Deterministic `MessageID`:
   - assistant: `"assistant:" + ThreadID + ":" + RunID` (one final answer per run)
   - user: `"user:" + ThreadID + ":" + SourceMessageID`
   Append only if no record with that `MessageID` exists. A duplicated terminal
   event must not double-append.
2. **Hydrate boundary.** In the read/list path only
   (`ListAIAgentTaskThreads` → `visibleTaskThreadsLocked`, NOT in `copyTaskThread`
   which `snapshot()` reuses), set `thread.Messages = lastN(taskThreadMessages[
   thread.ThreadID])`, sanitized with `clientVisibleTaskThreadText`. Keeps client
   changes near-zero (client already reads `thread.messages`).
3. **Overwrite guard.** Before any site that overwrites `thread.Message`
   (upsert :2226; the three stop helpers :2317/:2341/:2358), call
   `archiveThreadAssistantMessageLocked(thread, now)` to capture the still-unsaved
   assistant body first.

### Body extraction for an archived assistant turn

`archiveThreadAssistantMessageLocked` computes the body from live memory in this
order:
1. concatenate `thread.Lines[i].Message` for lines tagged
   `MessageKey == "assistant.partial"` for the current `RunID` (post-repair path);
2. fallback: the accumulated non-status `Lines` body (legacy / untagged);
3. fallback: `thread.Message` if it still holds the answer.
Then sanitize via `clientVisibleTaskThreadText`, dedupe by `MessageID`, append.

## Implementation plan (ordered)

### Phase 0 — Docs (SSOT-document-first)

- This plan doc (done).
- Update `riido-control-plane/docs/review/ai-agent-thread-message-history-loss-review-2026-06-10.md`:
  mark the inline-`Messages[]` "Fast Fix Plan" **superseded**; point to this plan;
  correct the `assistant.partial` "support" claim (it is dead — see Correction #1).
- Update the control-plane SSOT API doc that owns the thread endpoints (e.g.
  `docs/.../ai-agent-client-api.md` if present) for the new `messages` field +
  `message_key` passthrough.
- Update `riido-daemon/docs/20-domain/provider-runtime.md` (owning doc) to specify
  that assistant text deltas are reported as `assistant.partial`-keyed progress.

### Phase 1 — Contracts (SSOT, contract-first)

In `riido-contracts/apicontract/fixtures/control-plane-ai-agent-client.dsl.riido.json`:
- Add `message_key` (string, optional) — and for fidelity `message_code` (integer,
  optional) + `message_args` (map, optional) — to `AgentThreadProgressLine`
  (:1952-1973).
- Add a new `AIAgentTaskThreadMessageRecord` schema (message_id, thread_id, role,
  author_type, agent_id?, body, source_message_id?, run_id?, created_at).
- Add optional `messages` array (ref → `AIAgentTaskThreadMessageRecord`) to
  `AIAgentTaskThreadRecord` (:2100-2174). Keep it **out of `required`** (additive,
  non-breaking per `contract-promotion-policy`).
- **Regenerate** IR + OpenAPI from the DSL via the apicontract tool (never hand-edit
  the IR/OpenAPI). Verify all three fixtures stay in lockstep.
- Sync the byte-identical copy under
  `riido-control-plane/contracts/ai-agent-client/*.json`.
- Contracts version tag (next after v0.3.5) is cut **only if** a repo must import
  new contract Go types. CP/client structs are hand-written and the daemon uses a
  `map[string]string` metadata channel, so a tag is likely **not** required to
  compile; cut it at the end only to publish the fixture change if needed.

### Phase 2 — control-plane

- `ai_agent_client_api.go`: add `AIAgentTaskThreadMessageRecord`; add hydration-only
  `Messages` field to `AIAgentTaskThreadRecord`; in `AgentThreadProgressLine.MarshalJSON`
  (:512-523) **emit `message_key` (+ `message_code`/`message_args`)** so REST + SSE
  carry it.
- `ai_agent_client_development.go`:
  - add `taskThreadMessages map[string][]AIAgentTaskThreadMessageRecord` to the
    store; init in the constructor.
  - `archiveThreadAssistantMessageLocked(thread, now)` helper (body extraction +
    idempotent append) per "Body extraction" above.
  - call it at completion in `appendThreadProgressLocked` when
    `event.AssignmentState` is terminal, and as the guard before the overwrite at
    :2226 / :2317 / :2341 / :2358.
  - append a **user** message record in the follow-up path
    (`CreateAIAgentTaskThreadMessage` / `upsertTaskThreadMessageFromActionLocked`)
    using the user prompt body + `SourceMessageID`.
  - hydrate `thread.Messages` from `taskThreadMessages` in `visibleTaskThreadsLocked`
    (last N, sanitized) — read path only.
- Persistence (separate collection must survive):
  - `AIAgentClientSnapshot`: add `TaskThreadMessages map[string][]AIAgentTaskThreadMessageRecord json:"task_thread_messages,omitempty"`.
  - `AIAgentClientThreadsSnapshot`
    (`dynamodb_ai_agent_client_snapshot_split.go:51-55`): add the same field.
  - `splitFromCombined` (:99) and `combinedFromSplit` (:116): carry the map.
  - `snapshot()`: copy `s.taskThreadMessages` into the snapshot, **capped** (last N
    per thread + per-body byte cap) to protect the 400 KB DynamoDB item.
  - `restoreSnapshotWithSubscriberMode`: restore `s.taskThreadMessages`.
  - `restoreCorePreservingRest`: extend the "preserved" comment (messages live with
    threads, untouched by core-only).
  - **Verify**: `RecordAIAgentAssignmentEvent` / `RecordAIAgentThreadProgress`
    persist via `saveSnapshot` (→ `saveAllSplit`, writes the threads item), NOT the
    core-only Sync path — otherwise archived messages never reach storage.

### Phase 3 — daemon

- `.../saasplane/saasplane.go` `EventTextDelta` branch (:889-898): set
  `req.Metadata` with `ProgressMessageMetadataKey = "assistant.partial"` so body
  deltas are tagged (mirror the `EventProgress` metadata pattern).
- Owning doc updated in Phase 0 (AGENTS rule).
- **Sub-decision (open):** assistant.partial body is per-chunk today (one delta per
  content block). Either the daemon emits the running accumulated body, or the
  client concatenates all `assistant.partial` lines for the run. Chosen default:
  **per-chunk deltas + client concatenation**, to match the existing S3 concat
  render and keep the daemon simple. Confirm against `AgentThinkingLog` (which
  currently takes only the last partial line).

### Phase 4 — client

- Low priority; server hydration of `thread.messages` does most of the work.
- Ensure `getTaskThreadMessages` (`useAiAgentTask.ts:126`) stops trusting
  `thread.message` as history (it is a status/summary fallback only).
- If Phase 3 keeps per-chunk deltas, make `AgentThinkingLog` concatenate
  `assistant.partial` lines for the run instead of taking only the last.
- Align `response.ts` `messages[]` shape to the now-canonical contract (it is a
  hand-written best-effort guess today).

### Phase 5 — tests

1. (CP) assistant answer survives a follow-up to the same thread — extends
   `ai_agent_client_http_test.go:2034`.
2. (CP) assistant answer survives an agent replacement (routes through
   `markTask*ThreadsStoppedLocked`).
3. (CP) duplicate terminal event does not double-append (idempotency).
4. (CP) `message_key` survives REST + SSE JSON (MarshalJSON regression).
5. (CP) split snapshot save/restore preserves `task_thread_messages` and rehydrates.
6. Update the two overwrite-locking tests (`:2034`, `:1213`) to also assert
   `messages` contains the prior assistant body.
7. (daemon) `EventTextDelta` carries the `assistant.partial` metadata key.
8. (client) renders multiple `thread.messages` in order; does not hide them as
   status.

### Phase 6 — ship

- Build/test each repo (`GOWORK=off` for Go, matching CI).
- Cut contracts tag if required; deploy control-plane to dev via the
  `deploy-ai-agent-testnet` workflow (`deployment_target=development`) and verify
  (see the dev deploy/verify recipe). Confirm history survives a real follow-up.

## Risks & open decisions

- **assistant.partial granularity** (per-chunk vs accumulated) — defaulted to
  per-chunk + client concat; confirm during Phase 3/4.
- **History cap** — first cut hydrates/persists last N per thread; cursor-based
  pagination is a follow-up. Log what is dropped (no silent truncation).
- **Core-only persist lag** — like `Lines` today, a message archived between full
  `saveAllSplit` cycles is in memory until the next full save; acceptable and
  pre-existing, but the archive paths must use `saveSnapshot` (full), which they do.
- **Backfill** — existing threads start with empty history; archiving populates
  going forward (no migration of old answers, which are already lost/capped).
