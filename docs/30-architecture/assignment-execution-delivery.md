# Assignment execution and live response delivery

The control-plane retains ownership of authenticated assignment state and SSE.
Provider execution remains in the daemon. The client must receive partial text
as an evolving response, activity as progress, and provider failures as visible
diagnostics. Existing clients continue to receive a readable message.

`AgentThreadProgressLine.message_key` is an optional presentation discriminator:
`assistant.partial` is a complete response-so-far snapshot scoped to the event's
assignment/thread/run. Clients replace the previous snapshot only when the
sequence increases; they do not append snapshots as text deltas. `runtime.activity`
is a tool or approval activity line. Other template metadata stays server-side.
Snapshot retention keeps the latest assistant body per thread. Terminal messages
replace the live response. Error/warning events stay visible during execution.

Assignment instructions require execution when the request and available tools
are sufficient, and an explanation of observed blockers otherwise. Instructions
cannot provision a missing repository or establish evidence that changes passed
tests; those are separate daemon execution responsibilities.

Verification: provider/reporting fixtures in the daemon; progress serialization,
SSE and assignment tests here; local desktop manual streaming/failure/cancel and
disposable-workspace execution checks. Roll out this additive projection before
the matching client and daemon. Production deployment is outside this work unit.

Local profile: `RIIDO_AI_SERVER_LOCAL_STATE_DIR` selects single-process file
snapshots for assignments and Agent UI state. It requires an explicit numeric
loopback listen address and configured authentication, and is incompatible with
DynamoDB configuration. Device enrollment and API authorization still use the
normal handlers. State files contain credential hashes and must remain private.
This mode is for local validation, not multi-replica production deployment.

An optional local-state `task-contexts.json` maps workspace IDs to task IDs to
normal `AIAgentTaskContext` objects. It enables disposable task validation without
changing the business backend. Unknown workspace/task pairs fail; an explicitly
configured business task-context reader takes precedence.

Reconciliation must not hold the persistent Agent store's operation mutex while
calling the assignment actor. Heartbeats perform the reverse lookup (assignment
actor → Agent registry), so that lock order deadlocks concurrent task queries.
Read candidate threads under the local lock, fetch projections without it, then
reload and apply only if the candidate thread is unchanged. Terminal and newer
progress remain authoritative. Regression: reentrant registry lookup during a
projection read must finish within one second.


Local validation on 2026-10-12 used the normal authenticated enrollment and
assignment endpoints with Codex CLI 0.161.0: actual repository change + passing
checks, SSE partial response before completion, explicit failed-check cause, and
active-provider cancellation to `stopped` with zero sessions were observed.
Regression tests cover the reconciliation lock cycle and monotonic sequence IDs
when replacing partial snapshots. After separating concurrent dashboard, request
and notification changes into their own work, the dedicated PR branch passes
`GOWORK=off go test ./...` with the pinned contracts commit. A fresh source build
also passed the desktop verify-local-agent-stack.py harness (15 live snapshots).
Prior shared-checkout failure/cancellation evidence remains distinct from this
success-path rerun. The local services are stopped after validation.

This Draft preserves the validated source baseline and requires reconciliation
with newer main before merge. The related contracts PR adds only optional
workspace_id; client assignment routes derive it from the authenticated principal
and the assignment store preserves it for exact daemon repository bindings.
No new notification, dashboard or additive-request API is included.
