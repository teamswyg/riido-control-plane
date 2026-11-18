# Task Agent requests — RIID-5828

2026-10-12

The additive assignment route and structured comment route share a durable
request identity. Additional instructions append to task context, never to the
system instruction. New requests to a busy Agent remain queued; retrying a request
returns its original assignment, including after terminal completion and replay.
The API reports queued until the daemon actually starts. Task assignment membership
is independent of execution state. Human member mutations remain in monolith.

Verification: request retry/replay, persistence failure, busy Agent follow-up,
multiple Agents, prompt composition, HTTP authorization and generated API drift.
Keep v1 and callers without request_id compatible. Server contract precedes client
rollout; deployment evidence is recorded separately from these executable contracts.

## Delivery and rollout

The monolith commits the comment and a v2-only outbox record together. The
internal comment delivery endpoint accepts only an event ID with a dedicated
server credential. It fetches persisted evidence from the monolith, where the
comment, task, team, workspace and current author membership are checked. The
Agent catalog applies the author's public/private Agent access before enqueue.
No browser JWT is persisted for background retries. User text is never an
instruction to the server and remains in the provider user prompt.

Identified assignment creation uses a DynamoDB transaction with create-only
conditions for both the operation and assignment projection. Losing replicas
read the winner, including a winner that has already completed. Client snapshots
use version compare-and-swap; stale writes fail rather than erasing another
Agent's membership. Callers retry the same identity after persistence errors.
Before activation, drain old replicas that still write snapshots without CAS.
Existing persisted active threads seed membership when the new map is absent.
Cancellation and daemon event writes read the latest durable projection and commit
the operation and projection with matching state, event sequence and lease token.
A raced claim retries cancellation as cancelling; the active lease stays owned
until acknowledgement or expiry. Late running events cannot undo cancellation. A durable queue miss never falls
back to a stale local queued copy.
Replaying an accepted request after explicit unassignment does not restore membership.

Runtime configuration: `RIIDO_AI_SERVER_COMMENT_DELIVERY_EVIDENCE_URL` points to
the monolith evidence service; `RIIDO_AI_SERVER_COMMENT_DELIVERY_KEY` is a
separately provisioned server secret of at least 32 bytes. Both empty disables
the receiver. HTTPS is required outside loopback development. Never place values
in this repository. Configuration ownership is recorded in
`runtime-cd-ownership.riido.json`.

Apply the monolith's `0013-agent-comment-delivery.sql` migration before enabling
its worker. Deploy compatible contracts/control-plane/daemon before the client.
Both Go consumers pin the published contracts commit `16034fc69f42` through
`v0.3.7-0.20261015041017-16034fc69f42`; builds do not depend on a local workspace.
Request receipts are persisted separately from display history. Intent-gated
retries reuse their first outcome even after task text changes or unassignment.
Active threads remain available for cancellation beyond the display history limit.

Rollback disables new comment submissions and the worker, preserving pending
rows and identified assignments for a later compatible restart. Do not drop the
outbox as part of rollback. The old request-less routes remain compatible; only
identified requests opt into durable new execution per submission.

현재 참여자 generation 및 제거 tombstone을 유지한다. 식별 요청 생성도 journal/projection/generation/receipt의 같은 transaction을 사용하며, 일반 writer는 최신 참여자 relation을 보존하고 변경 part와 CURRENT를 한 번에 저장한다.

## Automatic PR checkout selection

For a new execution, a generated task branch is a suggestion, not evidence
that the branch exists. When the authenticated task read supplies complete
GitHub and repository metadata, use the most recently updated linked PR in
the selected repository. PR state does not affect selection. Compare parsed
`updatedAt` instants; equal or unavailable dates use the greater PR number as
a deterministic tie-breaker. Duplicate records retain their newest date.

Checkout uses `refs/pull/<number>/head`, the selected PR's actual head content,
without requiring its display branch name or substituting the default branch.
This also works after a head branch is deleted. The existing daemon fetches
that exact ref using its normal Git credentials. No new credentials or GitHub
API permission are introduced.

An explicit configured branch is preserved. A user can override automatic
selection on the first instruction line with `PR #123`, `branch: name`, or
`브랜치: name`; a PR must be linked to this task in the selected repository.
Invalid or unauthorized choices fail before queueing and never fall back to a
contextless assignment. Ordinary prose does not count as a checkout command.

The same rules apply when retrying a workspace preparation failure through a
conversation reply. Existing request deduplication and native-session binding
remain in force; this selection does not restart existing executions or move a
resumed session into another worktree. Incomplete metadata and PRs from other
repositories cannot supply an automatic checkout.
