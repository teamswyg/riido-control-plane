# API Client Delivery

> Riido task: RIID-4746 `[Control Plane] AI Agent client API delivery SSOT 정리`

This file owns the control-plane decision for delivering the AI Agent client API
as generated React Query code to `riido-client`.

## Purpose

`riido-control-plane` must make the AI Agent client API feel like a lightweight
frontend module while keeping API shape, lifecycle metadata, and generated code
traceable to one source of truth.

The target client application code is not edited by this slice. This document
defines the current generated-client delivery boundary and the checks that must
pass before a workflow is allowed to open or update a generated-client PR in
`riido-client`.

## Ownership

| Owner | Owns | Does not own |
| --- | --- | --- |
| `riido-contracts` | canonical Riido vocabulary, domain policy grammar, shared enum and sum-type semantics, lifecycle/deprecation grammar | control-plane handler behavior, client repository branches |
| `riido-control-plane` | AI Agent client API sub-DSL, HTTP/SSE implementation, OpenAPI projection handoff, generated-client delivery workflow, release manifest | canonical business authority that changes non-API domain meaning |
| `riido-client` | consuming generated React Query code and reporting API usability feedback | OpenAPI SSOT, control-plane codegen execution, generated-code hand edits |

The AI Agent client API sub-DSL is not a random generated artifact. It is the
control-plane-owned API surface language that imports canonical terms from
`riido-contracts`. Client-facing API change requests enter through
`riido-control-plane`. If a request changes business meaning, narrows or expands
domain policy, or exposes a contradiction in the canonical language, it must be
escalated to `riido-contracts`.

The same escalation path applies in the other direction: backend implementation,
cost, or operations findings may ask `riido-contracts` to change or remove a
canonical concept. That does not immediately delete the lower API sub-DSL. The
API surface moves through lifecycle states so existing clients can migrate.

Agent setting fields follow that dependency direction. `profile_thumbnail_url`,
`description`, `instruction`, server-authored `created_at`, and
server-authored `updated_at` are canonical terms from `riido-contracts`; this
repository delivers them as generated client API shape and history/manifest
metadata. If client usage suggests a different thumbnail storage model,
description limit, instruction semantics, timestamp meaning, or model catalog,
the delivery workflow records the local finding but does not alter the
canonical meaning without a contracts change.

Figma menu placement (`node-id=156-19307`) is a generated-client consumption
context, not a generated-client endpoint. The delivery workflow may document
which generated calls a route uses after navigation, but client menu labels,
ordering, selected state, and route wiring stay outside the control-plane API
projection.

The exhaustive Figma top-level coverage SSOT is owned by `riido-contracts` in
`docs/30-architecture/figma-ai-agent-coverage.riido.json`. This repository keeps
only a downstream generated-client projection in
[`figma-ai-agent-control-plane-projection.md`](figma-ai-agent-control-plane-projection.md)
and
[`figma-ai-agent-control-plane-projection.riido.json`](figma-ai-agent-control-plane-projection.riido.json).
That local projection may repeat node ids and generated path hints, but only to
check OpenAPI/generated TypeScript drift. It must not redefine Figma coverage,
business policy, daemon lifecycle, client layout, or infra topology.

Figma task-thread annotations (`node-id=153-15931`) are also consumption context.
They may name generated call chains such as `riido.aiAgent.events.stream` and
`riido.aiAgent.tasks.assign`, `riido.aiAgent.tasks.unassign`, and
`riido.aiAgent.tasks.stop`, but the delivery artifact must derive those chains
from the control-plane OpenAPI projection and generated-client manifest. It must
not hand-code annotation strings as a second source of truth. Task-thread
screens first consume the generated cold collection call for
`GET /v1/client/ai-agent/tasks/{task_id}/threads`; the returned
`active_stream` link, when present, is the generated handoff into the SSE client
event stream.

Figma normal task-thread screen (`node-id=236-21379`) is generated-client
composition context for `riido.aiAgent.tasks.assign`,
`riido.aiAgent.tasks.unassign`, `riido.aiAgent.tasks.threads`,
`riido.aiAgent.tasks.threadMessages.create`, and
`riido.aiAgent.tasks.stop`. The compatibility
`riido.aiAgent.tasks.submitComment` route may remain in the artifact while
client screens still submit a comment-like action without a `thread_id`. The
generated artifact may document that those calls are commonly used together on
the task page, but it must not turn the generic task comment box, right details
panel, reply input layout, send-button visual state, or agent row presentation
into API fields. Those remain client/task UI facts around the generated request
and response shapes.

Task participant assignment is documented as generated API behavior:
`tasks.assign` creates the initial typed thread response with
`comment_kind=assignment_started`, and `tasks.unassign` maps participant removal
to `comment_kind=stopped_by_user_request`. Whether stopped rows are hidden after
unassign is client presentation and must not become generated API copy.

Figma busy-agent queued screen (`node-id=153-8761`) is generated-client
composition context for the same task-thread calls. The delivery artifact may
document that `tasks.threadMessages.create` targets a known `thread_id` for a
next instruction, that the compatibility `tasks.submitComment` can still return
`comment_kind=queued_by_busy_agent`, that `tasks.threads` returns the queued row
on cold read, that `events.stream` can carry the typed queued status, and that
`tasks.stop` is the visible stop/cancel affordance. It must not hard-code the
Korean display copy, timestamp wording, avatar, or comment-row layout as a
generated API fact.
This busy-agent row is not the daemon handoff state. When the daemon reports
assignment `ready`, the client read model must expose the thread as
started/running with `comment_kind=assignment_started`, not
`queued_by_busy_agent`.

Figma stopped-by-deleted-agent screen (`node-id=227-19354`) is generated-client
composition context for `riido.aiAgent.agents.delete`,
`riido.aiAgent.tasks.threads`, and `riido.aiAgent.events.stream`. The delivery
artifact may document that deleting an agent can return
`running_tasks_force_stopped` and later expose
`comment_kind=stopped_by_agent_deleted` in the task-thread cold collection or
stream. It must not hard-code the Korean stopped-row copy, Riido actor label,
timestamp wording, hidden action state, avatar, or row layout as generated API
facts.

Figma participant dropdown annotations (`node-id=153-12742`) are generated-client
consumption context for `listAIAgentTaskAssignableAgents`. The delivery artifact
may document that this query feeds the AI Agent rows in the participant dropdown,
but member sorting, long-name truncation, max-height, scrollbar width, and
checkbox layout remain client implementation details.

Figma additional planning section (`node-id=153-15935`) is generated-client
boundary context for assignment targets. The delivery artifact may document that
AI Agent task calls are valid only on task and subtask surfaces. It must not
ship helper chains for projects, milestones, intakes, existing AI property
filling, or agent mentions unless a separate owning SSOT adds those operations.
Client code may choose to hide or disable agent UI on those non-target surfaces,
but generated code must not provide a misleading fallback path.

Figma runtime settings annotations (`node-id=162-23090`) are generated-client
consumption context for `listAIAgentDeviceRuntimes`,
`getAIAgentDaemon`, `startAIAgentDaemon`,
`restartAIAgentDaemon`, `stopAIAgentDaemon`,
`device_runtime_snapshot`, and `device_daemon_status_changed`. The delivery
artifact may document that these generated pieces feed the runtime settings
route, `내 기기`/`다른 기기` grouping, runtime name/version/status rows,
attached-agent avatar rows, agent-bound daemon detail labels, and
agent-bound daemon start/restart/stop buttons. It must not turn the agent
hover popover, daemon stop modal layout, or restart animation rendering into
server-owned presentation logic.

Figma agent setting section evidence (`node-id=432-37336`), add-screen evidence
(`node-id=134-6542`), and list-screen evidence (`node-id=432-35713`) are
generated-client consumption context for bootstrap/create/update/editability
APIs. The delivery artifact may document that `createAIAgent` feeds the
add-agent save flow, `created_at` feeds list creation dates, and `updated_at`
feeds list update dates and absolute-time tooltips. It must not turn row click,
meatball edit/delete entry, no-description row layout, status-label copy/color,
save-button enablement, long-description presentation, dropdown rendering, or
provider-specific model labels into generated enum/API facts. The resolved
`runtime_model_catalog.v1` rule is that clients render `RuntimeRecord.models`;
model labels stay runtime-scoped display data rather than enum members.

Figma onboarding annotations (`node-id=42-3014`) are generated-client
consumption context for bootstrap, devices, fixture query, and create APIs. The
delivery artifact may document that `riido.aiAgent.onboarding.fixtures` feeds
the fixture-selection screen, including copyable profile fields,
`default_visibility`, and `recommended_runtime_kind`, and that
`riido.aiAgent.onboarding.fixtures.createAgent` creates a normal agent with a
complete `CreateAgentConfigurationRequest` body. Fixture records must not
include model defaults or `model_id`; `runtime_model_catalog.v1` derives the
selected model from the chosen runtime's `RuntimeRecord.models` and the
omitted-`model_id` defaulting rule.
Workspace selection/list
scrolling and the `새 워크스페이스` row
shown in `node-id=164-30192`, row selection, direct-setting expansion, scroll,
direct-setting `이름` / `설명` / `지침` placeholders from `node-id=164-26969`,
two-line ellipsis, and no-installed-AI start state rendering from
`node-id=164-30206` remain client composition over generated data. The
all-disconnected Claude Code/Codex/OpenClaw/Cursor Agent rows and `시작하기` CTA
must not become generated provider-install/start helpers without a separate
owning SSOT. It must not make frontend hard-coded fixture copy a second SSOT.

Figma web onboarding annotations (`node-id=236-29749`) are not generated-client
endpoint evidence for this API. The delivery artifact may mention that AI Agent
screens are reached after auth/onboarding, but sign-up/login, terms consent,
email validation, member invite, app download, Windows launch notification,
marketing consent, and animation references stay in auth/team/product/client
documentation until a separate owning SSOT adds a generated operation. The AI
Agent generated client must not ship placeholder waitlist or marketing helpers
from this screen alone.

## DSL Lifecycle

The sub-DSL must preserve lifecycle metadata before OpenAPI generation:

| State | Meaning |
| --- | --- |
| `active` | current default API surface |
| `deprecated` | still generated, but JSDoc and OpenAPI mark replacement guidance |
| `superseded` | replaced by a newer operation/schema, still emitted for migration |
| `frozen` | no new features; bug/security fixes only |
| `removed_from_active_projection` | omitted from default generated client after the removal horizon |
| `archived` | retained only in historical manifest and changelog artifacts |

Lifecycle metadata must be lossless from DSL to IR. OpenAPI may expose the
standard `deprecated` flag, but Riido-specific metadata must remain in extension
fields such as `x-riido-lifecycle`, `x-riido-replacement`, and
`x-riido-removal-horizon`.

## Release Trigger

Generated-client delivery is handled by
`.github/workflows/generated-client-delivery.yml`. The current executable
trigger is `workflow_dispatch` so an operator can choose between package-only
review evidence and an actual cross-repository PR handoff. The package job can
be run manually with `create_pr=false` to produce a reviewable artifact without
requiring `riido-client` write credentials.

Delivery to `riido-client` is allowed from reviewed sources only when the same
Riido `branchName` and secret gates are preserved:

- a `riido-control-plane` Git tag that represents an API release
- an explicit manual workflow dispatch with `create_pr=true` (the current
  enabled delivery path)
- a `main` push that changes the AI Agent client OpenAPI/DSL/IR projection or
  the generated-client delivery generators/workflow, after a future automation
  slice wires it through the same branchName and credential checks

The contracts SSOT defines generated client delivery PRs as review handoffs.
This workflow may open or update a `riido-client` PR, but it must not auto-merge
that PR. `riido-client` owns the final generated-code review, application
integration, and merge decision.

Generated-client delivery is a Riido work unit. Before a delivery PR is opened
or refreshed in `riido-client`, the operator or automation must create the Riido
task and pass the returned `branchName` as the workflow `target_branch`. The
workflow must not synthesize helper branch names such as `react-query-*`.

## Target Branch

The workflow creates or updates a branch in `teamswyg/riido-client` from its
current `main`. The branch name is the Riido task response `branchName` and is
passed explicitly through workflow dispatch:

```text
{RIIDO_TASK_KEY}-{Riido task title slug}
```

Examples:

```text
A-60-AI-Agent-generated-client-handoff-최신화-및-자동-PR-검토
RIID-4865-클라이언트-온보딩-워크플로우-개선
```

`target_branch` must match the Riido branchName format
`^[A-Z][A-Z0-9]*-[0-9]+-.+` and must not contain `/`. The delivery generator
and workflow both reject non-Riido names so a generated client handoff cannot
silently drift away from the work-unit SSOT.

The branch must contain generated files only under this allowlist:

```text
src/generated/react-query/riido-control-plane/**
```

After generation, the workflow must fail if `git diff --name-only` contains any
path outside that allowlist. This prevents generator config drift, package
metadata drift, or unrelated client edits from riding along with generated API
delivery. If the generated files are identical to `riido-client` `main`, the
workflow must stop without creating or refreshing a client PR.

## Generator Boundary

Only the control-plane workflow output is trusted. `riido-client` consumes the
generated files but does not run control-plane codegen as part of normal
development.

Generator requirements:

- run `tools/reactquerygen` and `tools/generatedclienthandoff` from the
  control-plane Go module
- keep the generator dependency-free except for the repository's approved Go
  module boundary
- generate the client PR body from the same OpenAPI/DSL/IR inputs as the
  delivered TypeScript artifacts
- normalize the target branch with `riido-client`'s pinned Prettier setup before
  the final manifest and PR body are written
- run target-repository generated-path Prettier check and `pnpm run type-check`
  before the PR is opened or updated
- keep the generated output deterministic enough for reviewable diffs
- emit client-facing comments and notes in Korean, except for stable code
  identifiers, endpoint paths, enum literal values, package names, and repository
  names
- carry OpenAPI schema descriptions and operation summaries into generated
  JSDoc so frontend developers do not have to infer API behavior from type names
- carry `x-riido-client.generated_path` into generated JSDoc, and additionally
  emit the module-local path such as `tasks.threadMessages.create` so generated
  files can be searched by the route developers discuss in handoff notes
- generate a config-bound API facade so frontend developers can consume the
  surface as one module instead of stitching together isolated functions
- derive the facade module and namespace from DSL/IR client metadata projected
  into OpenAPI `x-riido-client`, not from generator-local operation-id switches

`tools/reactquerygen` is the deterministic OpenAPI-to-TypeScript generator for
both the checked-in development surface and the cross-repository handoff. The
handoff wrapper `tools/generatedclienthandoff` adds the generated README,
history, manifest, barrels, and PR body from the same OpenAPI/DSL/IR inputs.

## Projection Placement

The current projection chain is kept:

```text
contracts/ai-agent-client/*.dsl.riido.json
  -> contracts/ai-agent-client/*.ir.riido.json
  -> contracts/ai-agent-client/*.openapi.json
  -> generated TypeScript client files
```

Do not introduce an independent `dsl2` or `ir2` as a second SSOT for frontend
ergonomics. The client-facing library shape is a **client facet** inside the
existing AI Agent API sub-DSL:

- top-level `client_modules` describes generated module and namespace comments
- operation-level `client.module` and `client.facade_path` describe the nested
  library path
- operation-level `client.generated_path` is derived by `riido-contracts` from
  `client.module + "." + client.facade_path` and is used as generated comment
  metadata, not as a second route owner
- query operation `client.cache_tag` describes the cache root key
- command operation `client.invalidates` describes deterministic cache roots
  that a client may invalidate after a mutation

The IR projection must preserve those fields without changing their meaning.
OpenAPI exposes the same data as `x-riido-client-modules` and
`x-riido-client`. TypeScript codegen consumes only those OpenAPI extension
fields for facade structure, searchable generated comments, root query keys, and
invalidation helpers. If the metadata is missing, has a mismatched
`generated_path`, or references an unknown cache tag, generation fails.

The local generated endpoint smoke gate is documented in
[`ai-agent-generated-endpoint-smoke-matrix.md`](ai-agent-generated-endpoint-smoke-matrix.md).
The matrix file
`contracts/ai-agent-client/control-plane-ai-agent-client.smoke-matrix.riido.json`
must list every OpenAPI operation with `x-riido-client.generated_path` exactly
once and name the HTTP smoke test that exercises it. This matrix does not own
endpoint shape; it is executable evidence that generated facade paths such as
`riido.v2.aiAgent.tasks.threadMessages.create` and
`riido.v2.aiAgent.agents.daemon.stop` still map to live control-plane handlers
before client delivery.

This keeps SSOT ownership layered instead of duplicated:

- `riido-contracts` owns canonical vocabulary, policy grammar, enum/sum-type
  meaning, and lifecycle/deprecation grammar.
- `riido-control-plane` owns the AI Agent client API sub-DSL and its deterministic
  API/client projection.
- `riido-client` owns screen composition, hook call timing, optimistic updates,
  global error UX, invalidation timing, retry policy, and token refresh policy.

Client usability feedback can enter through `riido-control-plane` when it is
about endpoint shape, generated comments, cache relationship metadata, or
library ergonomics. Feedback that changes business meaning or policy authority
must still escalate to `riido-contracts`.

## Generated Artifacts

The client branch should receive:

- React Query hooks and DTO types generated from the OpenAPI projection
- `apiHistory.generated.ts`
- `contractManifest.generated.ts`
- `README.generated.md`
- `index.ts`
- `react.ts`

The same generated handoff step also writes `PR_BODY.generated.md` for the
client PR body. The PR body must include:

- changed source commit/ref and OpenAPI digest
- generated operation count and generated path list
- previous `contractManifest.generated.ts` versus current generated operation
  diff: added paths, removed paths, changed HTTP/operation/lifecycle entries, or
  an explicit no-API-surface-diff note
- SSOT decisions that affect frontend usage
- verification commands used by the workflow

The PR body is generated from the same OpenAPI/DSL/IR inputs as the delivered
files and, during cross-repository delivery, from the target branch's previous
generated contract manifest. This keeps the review text from drifting from the
generated artifact while giving frontend reviewers a real change summary instead
of only a full endpoint inventory.

When the workflow writes into `riido-client`, it must apply the target
repository's pinned Prettier configuration before finalizing the handoff. The
core and React generated files are formatted first, then
`tools/generatedclienthandoff` is run again against those formatted files so the
manifest hashes and PR body describe the exact files that land in the client
branch. Before replacing the generated directory, the workflow preserves the
previous `contractManifest.generated.ts` and passes it back to the handoff tool;
the resulting PR body must therefore tell reviewers whether this delivery adds,
removes, changes, or merely re-stamps generated API metadata. The remaining
generated README/history/manifest/barrel files are then formatted under the same
target-repository Prettier policy. The delivery job then runs a generated-path
Prettier check and the target repository's `pnpm run type-check` before it
commits and opens or updates the PR.

## Generated Client Facade

The generated client must keep primitive exports and a documented facade
together:

- primitive request functions, query keys, query options, mutation options, and
  generated types remain individually exported for tests, tree-shaking, and
  direct use
- `createRiidoControlPlaneClient(config)` groups the same primitives by DSL
  client metadata, for example `aiAgent.agents.editability`,
  `aiAgent.tasks.assign`, `aiAgent.tasks.unassign`, `aiAgent.tasks.stop`, and
  `aiAgent.devices.runtimes`
- workspace-scoped v2 operations are grouped under the versioned module root,
  for example `v2.aiAgent.bootstrap`, `v2.aiAgent.agents.create`, and
  `v2.aiAgent.tasks.threadMessages.create`
- the facade is config-bound, so consumers do not repeatedly pass `baseUrl`,
  `aiAgentToken`, and optional `fetcher`
- each operation exposes concise library-style aliases: `query(...)` mirrors
  `queryOptions(...)`, and `mutation(...)` mirrors `mutationOptions(...)`
- query endpoints expose `queryKeyRoot`, `queryKey`, `invalidate`,
  `invalidateAll`, and `prefetch`
- mutation endpoints expose deterministic `invalidates` helpers, but do not call
  them automatically
- the core facade does not import React hooks and remains server-safe
- the React wrapper is generated separately and imports hooks from
  `@/lib/react-query`, not directly from `@tanstack/react-query`
- the facade does not create or replace TanStack `QueryClient`
- the facade does not own token refresh, global error toasts, retry policy,
  automatic cache invalidation, or app-specific optimistic updates

The intended frontend usage is:

```ts
const riido = createRiidoControlPlaneClient(config);

useQuery(riido.aiAgent.bootstrap.query());
useQuery(riido.v2.aiAgent.bootstrap.query({ params: { workspace_id } }));
useMutation(riido.aiAgent.tasks.assign.mutation());
useMutation(riido.v2.aiAgent.agents.create.mutation());
useMutation(riido.aiAgent.tasks.stop.mutation());
queryClient.prefetchQuery(riido.aiAgent.devices.runtimes.query());
await riido.aiAgent.tasks.stop.invalidates.bootstrap(queryClient);
```

For client components that prefer a hook-shaped library surface, the generated
React wrapper is imported explicitly:

```ts
const riido = useRiidoControlPlaneClient(config);

const bootstrap = riido.aiAgent.bootstrap.useQuery({ enabled: !!config.aiAgentToken });
const workspaceBootstrap = riido.v2.aiAgent.bootstrap.useQuery({
  params: { workspace_id },
  enabled: !!config.aiAgentToken && !!workspace_id,
});
const stopTask = riido.aiAgent.tasks.stop.useMutation({
  onSuccess: async () => {
    await riido.aiAgent.tasks.stop.invalidates.all(queryClient);
  },
});
```

This keeps the generated output feeling like a lightweight module while leaving
application integration policy inside `riido-client`.

The history artifact is intentionally lightweight. It should expose release
entries, operation-level lifecycle changes, replacements, removals, and notable
breaking or deprecation notes without making frontend developers read the full
OpenAPI diff.

The release manifest is the authoritative machine-readable summary for a
control-plane API release. It links the control-plane tag, source DSL/IR digest,
OpenAPI digest, generator version, generated output path, and lifecycle summary.

Client-facing README, history notes, manifest notes, and generated JSDoc are part
of the delivered developer experience. They must be written in Korean because
`riido-client` UI and team-facing development notes are Korean-first. English is
kept only where changing it would change code identity or wire contract shape.

## Secret And Permission Boundary

Prefer a GitHub App installed only where needed. Required secrets live in
`riido-control-plane`, not in `riido-client`:

- `RIIDO_CLIENT_DELIVERY_APP_ID`
- `RIIDO_CLIENT_DELIVERY_PRIVATE_KEY`

The delivery workflow creates a short-lived installation token from those two
secrets and uses it only for `teamswyg/riido-client` checkout, branch push, and
PR create/update. The GitHub App must be installed on `teamswyg/riido-client`
and needs repository metadata read, contents read/write, and pull-request
read/write permissions. The generated-client package artifact can still be built
without these secrets; only the cross-repository PR delivery job requires them.

If a fine-grained token is used temporarily, it must be scoped to the target
repository and limited to contents and pull-request write permissions. The
current workflow accepts that temporary token as
`RIIDO_CLIENT_DELIVERY_TOKEN`. If neither the GitHub App secrets nor this
fallback token are configured, the delivery job must fail before checking out
`riido-client` and explain that cross-repository write permission is missing.
This failure is intentional only for `create_pr=true`; `create_pr=false` must
still build and upload the package artifact without cross-repository secrets.

The workflow must not require npm publish tokens, cloud credentials, Terraform
state, customer data, or production request tokens.

## Acceptance Criteria

- `riido-control-plane` docs name the API sub-DSL owner and canonical contract
  escalation path.
- Generated-client delivery is workflow-dispatched with the Riido work
  `branchName`; the workflow must not invent a synthetic delivery branch.
- Generated-client delivery opens or updates a client PR only; it never
  auto-merges the client PR.
- Target `riido-client` Riido branchName validation and generated path allowlist
  are defined.
- `tools/reactquerygen` and `tools/generatedclienthandoff` produce the delivered
  TypeScript, manifest, history, README, barrels, and PR body from the same
  OpenAPI/DSL/IR source.
- Generated history and manifest artifacts carry lifecycle/deprecation
  information from DSL/IR/OpenAPI through the client handoff when that metadata
  is present.
- Client-facing generated comments and notes are Korean-first and preserve
  OpenAPI descriptions in JSDoc.
- Generated JSDoc includes searchable generated paths, including the
  module-local form such as `tasks.threadMessages.create` and the example access
  form such as `riido.aiAgent.tasks.threadMessages.create`.
- Generated clients include `queryOptions`, `mutationOptions`, and the
  config-bound `createRiidoControlPlaneClient` facade without taking ownership of
  app-level query policies.
- Generated facade paths are sourced from DSL/IR/OpenAPI `client` metadata and
  fail generation if that metadata is missing or inconsistent.
- This slice does not edit `teamswyg/riido-client`.
