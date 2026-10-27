# Agent task context delivery

RIID-5818 owns readable, bounded task context and conversation continuity.
The assignment remains one task; references do not authorize additional work.

RIID-5900 limits CLI reads to 100 authenticated active-lease attempts per durable
assignment ID. All views share the counter, including live agent discovery,
reused request IDs, denied targets and unavailable data. Prompt capture, viewer
read-history GET and local whoami do not consume it. DynamoDB increments the
counter atomically across servers; exhausted budgets return HTTP 429 and storage
failures return 503 before data access. Repeating the same execution cannot
restore the budget. Heartbeat/lease refresh, runtime run-label changes and
snapshot recapture keep it; a new follow-up assignment receives a new budget.
Terminal transactions fence the counter even before the first CLI call and set
its TTL to the same first-terminal +24h deadline as context payloads. Read audit
records retain their existing policy. Active-lease validation precedes quota
consumption, so physical TTL cleanup cannot revive a terminal capability.

Comment-triggered executions retain extended context from the existing verified
monolith delivery evidence. The producer reads recent same-task root comments
(20 plus an out-of-page or reply trigger, at most 21), same-team linked work and
meetings, and bounded file/link metadata under the stored author's current
permissions. CP preserves that evidence without requesting a user credential
or making anonymous reads. Only older evidence with no `extended` field receives
the trigger-only fallback. The original comment remains the execution trigger;
normal snapshot capture, read audit and current viewer filtering still apply.

The request-scoped reader keeps the caller's existing bearer and team scope.
Only project, milestone and parent-task IDs already returned by the authorized
task detail may be followed. Parent reads are deduplicated, at most three in
parallel, and share the original request deadline. A failed, mismatched or
inaccessible parent is marked unavailable; its content is never substituted
from another component. Supported provider JSON is read as a bounded structured projection. An explicit
empty projection remains empty. Unversioned provider HTML never replaces an
existing nonempty component snapshot. Every legacy representation is explicitly
marked with unconfirmed freshness; none proves the latest Yjs revision.
Conflicting nonempty projections retain a bounded alternate as reference evidence
with unknown chronology instead of silently asserting which one is newer. Independent document reads are not an atomic cross-document
snapshot. Link/attachment metadata uses the same authenticated component routes;
arbitrary link or attachment contents are never fetched by the server.

The prompt contains the current task document first, then parent documents as
reference context. HTML is rendered as readable text with headings, lists,
tables, code and safe HTTP(S) links retained; scripts, styles and hidden content
are omitted. Text budgets preserve both beginning and end with an explicit
truncation marker. Missing or truncated context is not proof that it was read.

Follow-ups receive the latest task snapshot and the recent persisted user/agent
messages from the selected task, agent and conversation. Progress rows and other
conversations are excluded. Messages retain chronological order and roles, are
bounded and deduplicated; the latest user instruction remains authoritative.
Native resume continues to use the existing session and repository envelope.
Context delivery does not prove native execution or semantic task completion.

Verification covers rich document rendering, budgets, denied/mismatched parents,
deadline cancellation, conversation isolation and latest-instruction precedence.
Fixtures are synthetic; customer task bodies and authentication values are not
published as evidence. No frontend deployment is required by this change.

Private metadata reads for Agent Records and first-task onboarding provenance
can use `RIIDO_AI_SERVER_TASK_CONTEXT_GRAPHQL_URL` for the monolith GraphQL
endpoint. The caller's viewer Bearer and workspace/task checks remain required.
When omitted, the existing REST base plus `/graphql` remains the fallback.
REST component and provider-document reads keep their original base URL;
timeouts and redirect policy remain shared. Endpoint userinfo, query strings
and fragments are rejected by the existing URL validator. Synthetic two-server
regressions cover both metadata paths and REST preservation; they do not prove
live customer onboarding or authorize arbitrary linked-document fetches.

## Execution context · RIID-5820

Shared `assignment.ExecutionContext` carries bounded workspace kind, verification status, diagnostic code and next action. The server sanitizes the context before storing or returning it, retains it across legacy progress without new context, and copies it to action responses, cold thread/history and work-status events. New assignments never inherit a previous run's context. Absolute directories and verification argv stay in the daemon's host-only IPC. Missing context is legacy/unknown, not host-verified completion or a retry authorization. Provider recovery/failover execution remains outside this change.

## Single-pass answers · RIID-5827

Answer-only requests (including summaries, explanations and translations) use the
normal provider turn without requiring a repository or extra classifier/reviewer.
The latest follow-up overrides earlier task intent gates and mutation requests.
Provided context is reused; additional document reads require a concrete need.
Empty retained history falls back to the previous final answer. Requested changes
must be reported with observed evidence; this prompt policy is not a semantic
completion guarantee. See [delivery and measurements](single-pass-response-delivery.md).

## Document delivery repair · RIID-5818

Initial explicit instructions survive missing, denied, timed-out or invalid task
context. Context loss must not replace an already authorized directive with a
generic request to repeat it. Follow-up instruction and task ownership rules stay
unchanged.

With an explicitly configured GraphQL metadata endpoint, document JSON/HTML and
file/link metadata use two independent viewer-authorized queries per checked
component (one document, one resources), so optional resource failure cannot
discard the document. Workspace and component identity checks still run first; parent IDs
still come only from the checked hierarchy. There is no permission widening,
cross-principal cache or fallback to another server after GraphQL denial. A
REST-only client retains bounded JSON-first and HTML-fallback compatibility.
Supported Tiptap projections are rendered without evaluating raw HTML or fetching
embedded resources. Null, missing, malformed and unsupported data are unavailable;
a valid empty document is distinct and cannot resurrect older content.

Legacy projections have no common revision witness. JSON availability is useful
context, not proof of current Yjs parity; HTML and component snapshots are also
marked unconfirmed. Finite read outcome/operation/source and duration attributes
expose degradation without document IDs, content, tokens or upstream error text.
The overall timeout remains configured by the operator; synthetic call-count and
deadline checks are not production latency or model-comprehension guarantees.

## Related meeting bodies · RIID-5900

At assign/reply time, the existing requester bearer may capture up to 20 directly linked meetings, using the task's verified team. Each meeting detail must match its requested ID and have no deletion timestamp before provider JSON/HTML is read. Legacy meetings without a document use their authorized description. Denial never falls back to stored text. Empty projections remain empty; stored representations do not establish current Yjs freshness. At most four meeting reads run concurrently under the original request deadline. Each body, including alternate text, is limited to 8192 UTF-8 bytes. Meeting documents also participate in the existing 256KiB snapshot bound and terminal-plus-24h retention; no user token is persisted.

The existing assignment-context endpoint adds `view=meetings` (optional task target) and `view=meeting` (required linked meeting ID). Base context, resources, list and prompt contain meeting metadata only. Only explicit meeting reads return the selected document and record `related_meeting/body` with its captured version. Empty is a successful empty body; denied is 403, deleted is 404, and unavailable is 503. Outside-snapshot targets remain 403. Per-body or whole-snapshot truncation is explicit. Current viewer permission to read that meeting body is required to expose its historical body-read metadata.

This slice covers meeting list and body selection only. Meeting boards, meeting comments and additional linked-work traversal remain unimplemented; synthetic contract tests are separate from actual staging reads.
