# Agent task context delivery

RIID-5818 owns readable, bounded task context and conversation continuity.
The assignment remains one task; references do not authorize additional work.

The request-scoped reader keeps the caller's existing bearer and team scope.
Only project, milestone and parent-task IDs already returned by the authorized
task detail may be followed. Parent reads are deduplicated, at most three in
parallel, and share the original request deadline. A failed, mismatched or
inaccessible parent is marked unavailable; its content is never substituted
from another component. Provider document content takes precedence over stored
HTML when available. A fallback component snapshot is explicitly marked as having
unconfirmed freshness. Independent document reads are not an atomic cross-document
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
