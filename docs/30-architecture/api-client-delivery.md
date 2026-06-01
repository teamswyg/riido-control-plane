# API Client Delivery

> Riido task: RIID-4746 `[Control Plane] AI Agent client API delivery SSOT 정리`

This file owns the control-plane decision for delivering the AI Agent client API
as generated React Query code to `riido-client`.

## Purpose

`riido-control-plane` must make the AI Agent client API feel like a lightweight
frontend module while keeping API shape, lifecycle metadata, and generated code
traceable to one source of truth.

The target client repository is not edited by this slice. This document defines
the future delivery boundary and the checks that must exist before a workflow is
allowed to open a generated-client PR.

## Ownership

| Owner | Owns | Does not own |
| --- | --- | --- |
| `riido-contracts` | canonical Riido vocabulary, domain policy grammar, shared enum and sum-type semantics, lifecycle/deprecation grammar | control-plane handler behavior, client repository branches |
| `riido-control-plane` | AI Agent client API sub-DSL, HTTP/SSE implementation, OpenAPI projection handoff, generated-client delivery workflow, release manifest | canonical business authority that changes non-API domain meaning |
| `riido-client` | consuming generated React Query code and reporting API usability feedback | OpenAPI SSOT, Orval execution, generated-code edits |

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

Figma agent setting annotations (`node-id=164-50215`), add-screen evidence
(`node-id=134-6542`), and list-screen evidence (`node-id=432-35713`) are
generated-client consumption context for bootstrap/create/update/editability
APIs. The delivery artifact may document that `createAIAgent` feeds the
add-agent save flow, `created_at` feeds list creation dates, and `updated_at`
feeds list update dates and absolute-time tooltips. It must not turn row click,
meatball edit/delete entry, no-description row layout, status-label copy/color,
save-button enablement, long-description presentation, dropdown rendering, or
provider-specific model labels into generated API facts before the contracts
model-catalog question is resolved.

Figma onboarding annotations (`node-id=42-3014`) are generated-client
consumption context for bootstrap, devices, and create APIs. The delivery
artifact may document that `agent_templates` feeds the starter-agent selection
screen, but workspace selection/list scrolling and the `새 워크스페이스` row
shown in `node-id=164-30192`, row selection, direct-setting expansion, scroll,
direct-setting `이름` / `설명` / `지침` placeholders from `node-id=164-26969`,
two-line ellipsis, and no-installed-AI start state rendering from
`node-id=164-30206` remain client composition over generated data. The
all-disconnected Claude Code/Codex/OpenClaw/Cursor Agent rows and `시작하기` CTA
must not become generated provider-install/start helpers without a separate
owning SSOT. It must not make frontend hard-coded template copy a second SSOT.

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

Generated-client delivery is allowed only from a `riido-control-plane` Git tag
that represents an API release, for example `v1.20.2`.

Regular pushes to `main` may validate local drift, but they must not push a
branch to `riido-client`. This keeps frequent server development from creating
client churn and CI cost.

The contracts SSOT defines generated client delivery PRs as review handoffs.
This workflow may open or update a `riido-client` PR, but it must not auto-merge
that PR. `riido-client` owns the final generated-code review, application
integration, and merge decision.

## Target Branch

The future workflow creates a branch in `teamswyg/riido-client` from its current
`main`:

```text
react-query-{tag}-{shortsha}
```

Examples:

```text
react-query-v1.20.2-a1b2c3d
```

The branch must contain generated files only under this allowlist:

```text
src/generated/react-query/riido-control-plane/**
```

After generation, the workflow must fail if `git diff --name-only` contains any
path outside that allowlist. This prevents Orval config drift, package metadata
drift, or unrelated client edits from riding along with generated API delivery.

## Generator Boundary

Only the control-plane workflow output is trusted. `riido-client` consumes the
generated files but does not run Orval as part of normal development.

Generator requirements:

- pin the Orval package version in the control-plane generator workspace
- commit the lockfile for that generator workspace
- run with frozen lockfile semantics
- avoid `npx` or floating package installs
- disable or audit install scripts where the package manager supports it
- keep the generated output deterministic enough for reviewable diffs
- emit client-facing comments and notes in Korean, except for stable code
  identifiers, endpoint paths, enum literal values, package names, and repository
  names
- carry OpenAPI schema descriptions and operation summaries into generated
  JSDoc so frontend developers do not have to infer API behavior from type names
- generate a config-bound API facade so frontend developers can consume the
  surface as one module instead of stitching together isolated functions
- derive the facade module and namespace from DSL/IR client metadata projected
  into OpenAPI `x-riido-client`, not from generator-local operation-id switches

`tools/reactquerygen` remains a small deterministic public fixture generator for
the checked-in mock surface. It is useful for drift tests, but it is not the
cross-repository Orval delivery mechanism.

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
- query operation `client.cache_tag` describes the cache root key
- command operation `client.invalidates` describes deterministic cache roots
  that a client may invalidate after a mutation

The IR projection must preserve those fields without changing their meaning.
OpenAPI exposes the same data as `x-riido-client-modules` and
`x-riido-client`. TypeScript codegen consumes only those OpenAPI extension
fields for facade structure, comments, root query keys, and invalidation
helpers. If the metadata is missing or references an unknown cache tag,
generation fails.

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
useMutation(riido.aiAgent.tasks.assign.mutation());
useMutation(riido.aiAgent.tasks.stop.mutation());
queryClient.prefetchQuery(riido.aiAgent.devices.runtimes.query());
await riido.aiAgent.tasks.stop.invalidates.bootstrap(queryClient);
```

For client components that prefer a hook-shaped library surface, the generated
React wrapper is imported explicitly:

```ts
const riido = useRiidoControlPlaneClient(config);

const bootstrap = riido.aiAgent.bootstrap.useQuery({ enabled: !!config.aiAgentToken });
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

If a fine-grained token is used temporarily, it must be scoped to the target
repository and limited to contents and pull-request write permissions.

The workflow must not require npm publish tokens, cloud credentials, Terraform
state, customer data, or production request tokens.

## Acceptance Criteria

- `riido-control-plane` docs name the API sub-DSL owner and canonical contract
  escalation path.
- Release delivery is tag-triggered, not main-push-triggered.
- Generated-client delivery opens or updates a client PR only; it never
  auto-merges the client PR.
- Target `riido-client` branch naming and generated path allowlist are defined.
- Orval is a pinned control-plane generator dependency when the workflow is
  implemented.
- Generated history and manifest artifacts carry lifecycle/deprecation
  information from DSL/IR through the client handoff.
- Client-facing generated comments and notes are Korean-first and preserve
  OpenAPI descriptions in JSDoc.
- Generated clients include `queryOptions`, `mutationOptions`, and the
  config-bound `createRiidoControlPlaneClient` facade without taking ownership of
  app-level query policies.
- Generated facade paths are sourced from DSL/IR/OpenAPI `client` metadata and
  fail generation if that metadata is missing or inconsistent.
- This slice does not edit `teamswyg/riido-client`.
