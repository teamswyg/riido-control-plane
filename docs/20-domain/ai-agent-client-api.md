# AI Agent Client API

> Riido task: RIID-4721 `[Server] AI Agent client-facing endpoint handlers`

This file is the control-plane SSOT for the mockable AI Agent client API
implemented by `internal/riidoaiserver`.

## Contract Source

The contract projection is checked in under
`contracts/ai-agent-client/`:

- DSL: `control-plane-ai-agent-client.dsl.riido.json`
- IR: `control-plane-ai-agent-client.ir.riido.json`
- OpenAPI: `control-plane-ai-agent-client.openapi.json`
- generated React Query: `web/generated/aiAgentClient.ts`

Canonical vocabulary, shared enum semantics, and lifecycle/deprecation grammar
are owned by `riido-contracts`. This repository owns the AI Agent client API
sub-DSL that imports those canonical terms and projects them to OpenAPI for the
running mock API, smoke tests, and generated frontend client.

Client-usability API changes enter through `riido-control-plane`. If a change
modifies business meaning, policy scope, or canonical vocabulary, the decision
must be escalated to `riido-contracts` before this repository updates the
sub-DSL.

Cross-repository React Query delivery to `riido-client` is owned by
[`api-client-delivery.md`](../30-architecture/api-client-delivery.md).

## SSOT Dependency Direction

This file is downstream of the canonical AI Agent policy in `riido-contracts`.
It may repeat policy words only to explain local HTTP behavior, mock data,
generator output, and black-box harness coverage.

For agent settings:

- `riido-contracts` owns the meaning of `profile_thumbnail_url` and
  `description` and `instruction`, including URL-only thumbnail policy, the 160
  character description limit, and the 1000 character instruction limit.
- `riido-contracts` owns onboarding template catalog semantics. This repository
  projects the catalog in `ClientBootstrapResponse.agent_templates` and seeds
  deterministic mock templates for frontend development.
- This repository owns POST/PATCH validation, create/save/update behavior,
  response projection, generated TypeScript shape, and smoke-test coverage.
- `riido-daemon` owns runtime consumption of an assigned instruction value after
  the assignment contract carries it; this server doc does not define provider
  prompt placement.
- `riido-infra` owns deployment/storage changes only when the API requires new
  media storage, secrets, persistence topology, or release evidence.
- Figma menu placement (`node-id=156-19307`) is a client route affordance, not
  a server menu contract. This repository owns the protected data endpoints used
  after a client opens the AI Agent, runtime, or agent-management route.
- Figma task-thread annotations (`node-id=153-15931`) cite
  `riido.aiAgent.events.stream` and `riido.aiAgent.tasks.stop` as generated
  client consumption paths. This repository owns the matching HTTP/SSE behavior
  and generated shape. It does not own client scroll-to-thread behavior, hover
  states, modal layout, viewer-away notification rendering, or progress
  animation references.
- Figma participant dropdown annotations (`node-id=153-12742`) confirm the
  `assignable-agents` API consumption context. This repository owns only the
  visible AI Agent list, owned-first ordering, generated DTO shape, and black-box
  tests. It does not own member sorting, long-name truncation, max dropdown
  height, scrollbar width, checkbox layout, or the mixed member/agent visual
  composition.
- Figma runtime settings annotations (`node-id=162-23090`) confirm the
  `devices` API consumption context. This repository owns the protected
  device/runtime read model, `device_runtime_snapshot` event shape, generated
  DTOs, and black-box tests. It does not own the agent hover popover, daemon
  stop modal, restart animation, or desktop-local daemon lifecycle command
  composition.
- Figma onboarding annotations (`node-id=42-3014`) confirm the bootstrap
  consumption context. This repository owns the protected bootstrap projection
  of `agent_templates` and mock coverage. It does not own workspace selection,
  row selection, direct-setting expansion, scrolling, two-line ellipsis,
  preview-popover layout, or the client decision to skip template selection
  when no selectable runtime exists.
- Figma web onboarding annotations (`node-id=236-29749`) confirm that sign-up,
  terms consent, member invite, macOS app download, Windows waitlist, and
  animation references are outside this protected AI Agent API unless a separate
  owning SSOT promotes them. This repository must not add generated operations
  for auth/team/product/distribution presentation from that screen alone.

Bottom-up findings from this repository, such as validation cost, generator
shape, or frontend usability, can start here. If the finding changes domain
meaning, the next PR must update `riido-contracts` first and then refresh this
sub-DSL and generated output.

## Mock Runtime

The mock surface is enabled by `RIIDO_AI_SERVER_AI_AGENT_CLIENT_MOCK=true`.
When disabled, protected AI Agent client routes fail closed with `503` instead
of returning synthetic data.

The mock API implements:

- `GET /v1/client/ai-agent/bootstrap`
- `GET /v1/client/ai-agent/devices`
- `GET /v1/client/ai-agent/tasks/{task_id}/assignable-agents`
- `GET /v1/client/ai-agent/tasks/{task_id}/threads`
- `POST /v1/client/ai-agent/tasks/{task_id}/comments`
- `POST /v1/client/ai-agent/tasks/{task_id}/stop`
- `POST /v1/client/ai-agent/agents`
- `GET /v1/client/ai-agent/agents/{agent_id}/editability`
- `PATCH /v1/client/ai-agent/agents/{agent_id}`
- `DELETE /v1/client/ai-agent/agents/{agent_id}`
- `GET /v1/client/ai-agent/events`

The SSE endpoint supports `?replay=1` for deterministic smoke checks. Without
`replay=1`, it keeps the connection open as a client event stream.

## Policy

- visible agents are the viewer's owned agents plus other users' public agents
- task participant dropdown responses are ordered owned-first, then by name
- task participant dropdown UI presentation, member sorting, and overflow
  behavior are client-owned and do not rewrite the returned agent order
- bootstrap returns an ordered `agent_templates` catalog used by the Figma
  onboarding agent-template selection screen
- if no runtime is selectable, clients use existing device/runtime data to skip
  the template-selection step; SaaS does not add a separate onboarding skip
  command
- non-owner, non-admin users cannot mutate other users' public agents
- client-facing agent creation stamps `owner_principal_id` from authorization,
  requires `name`, `visibility`, and a viewer-owned `runtime_id`, and starts as
  editable with zero assigned tasks
- editing is blocked while `assigned_task_count` is greater than zero
- `profile_thumbnail_url` is saved as an optional HTTPS image URL string on the
  agent record; binary image upload/storage is outside this mock API
- `description` is saved as optional client-authored one-line agent summary
  text and is rejected when it exceeds 160 characters; UI truncation/wrapping is
  a client concern and does not rewrite the stored value
- `instruction` is saved as optional client-authored agent guidance text and is
  rejected when it exceeds 1000 characters
- `updated_at` is a server-authored RFC3339 date-time on every
  `AgentClientRecord`; it is refreshed when editable agent configuration is
  saved and lets clients render update dates or absolute-time tooltips
- delete returns forced assignment effects for queued/running mock tasks
- task-thread comments can enqueue work when the selected agent is busy
- task-thread stop actions return `stopped_by_user_request`
- task-thread status updates use typed `AgentTaskCommentKind` values
- task-thread screens first call `GET /v1/client/ai-agent/tasks/{task_id}/threads`
  to render historical AI Agent comments; `active_stream` is present only when
  the screen should also connect to the client event stream
- task-thread comments created while the viewer was on another screen are still
  returned by the later cold collection; scroll/focus presentation remains a
  client concern
- daemon progress ingest accepts parsed `<riido_log>...<end>` batches through
  `POST /v1/agents/{agent_id}/thread-progress`
- client task-thread progress is streamed as the typed
  `agent_thread_progress` event with `thread_id` on
  `GET /v1/client/ai-agent/events`
- runtime settings consume `GET /v1/client/ai-agent/devices` and
  `device_runtime_snapshot`; SaaS does not expose a client endpoint to stop or
  restart a user's local daemon

## Figma Handoff Evidence

Confirmed in Chrome against `v.1.22 AI Agent` on 2026-05-28 and 2026-05-29:

- `node-id=153-12742`: task participant dropdown section; annotations confirm
  member 가나다 sorting, AI Agent owned-first then name ordering, long-name UI,
  and max-height/scrollbar behavior
- `node-id=153-15931`: task-thread communication section; Dev Mode annotations
  include `riido.aiAgent.events.stream` on `node-id=153-8545` and
  `riido.aiAgent.tasks.stop` on `node-id=236-20768`
- `node-id=236-21379`: normal task thread with comment input and agent reply
- `node-id=153-8761`: queued task comment when the agent is already busy
- `node-id=227-19354`: task stop flow with stopped agent comment
- `node-id=156-19307`: AI Agent menu placement in the workspace sidebar,
  including `Menubar/default` and `Menubar/setting` dark/light variants
- `node-id=162-23090`: runtime settings page; Dev Mode annotations identify the
  agent hover popover, daemon stop modal, and restart-in-progress animation
- `node-id=275-22731`: runtime settings empty-state section; annotations mark
  provider install-card hover states, while the screen text shows no-daemon,
  no-current-runtime, Windows app waitlist, and marketing-consent variants
- `node-id=164-50215`: agent setting page with profile image, name,
  description, runtime, model, visibility, and instruction fields
- `node-id=134-6542`: agent add page with profile image, name, description,
  runtime, model, visibility, instruction, and save controls
- `node-id=42-3014`: onboarding planning page; annotations include runtime
  selection (`node-id=137-6746`), template/direct-setting selection
  (`node-id=138-7389`, `node-id=164-26969`), two-line template description
  ellipsis (`node-id=164-27719`), and no-installed-AI skip behavior
  (`node-id=164-30206`)
- `node-id=236-29749`: web onboarding section; annotations include chat
  animation reference, Google sign-up wording, Google sign-up requiring terms
  consent, email sign-up terms row click behavior, and button progress-bar
  references. The section also shows macOS app download, email sign-up, member
  invite/link-copy, Windows launch notification, waitlist completion, and
  marketing-consent variants.

`node-id=156-19307` does not add a new endpoint. It confirms that the frontend
needs visible entry points into AI Agent/runtime/agent-management surfaces; the
current server responsibility remains `bootstrap`, `devices`, agent mutation,
task-thread actions, and `events`.

`node-id=153-15931` confirms that frontend thread composition needs a cold read
before optional streaming. `tasks.threads` maps to
`GET /v1/client/ai-agent/tasks/{task_id}/threads`, `tasks.stop` maps to
`POST /v1/client/ai-agent/tasks/{task_id}/stop`, and `events.stream` maps to the
`GET /v1/client/ai-agent/events` SSE surface. The server returns
`active_stream` only for a currently active task thread. If a thread was created
while the viewer was elsewhere, the server responsibility is to return that
persisted visible thread on the later cold read; the client owns any scroll or
focus behavior that makes the newly relevant thread visible in the task view.

`node-id=153-12742` maps to the existing
`GET /v1/client/ai-agent/tasks/{task_id}/assignable-agents` endpoint and the
generated `listAIAgentTaskAssignableAgents` query. The server must preserve the
agent response policy, while the client composes that response with member rows
and renders truncation, max height, and scrollbars.

`node-id=162-23090` maps to the existing
`GET /v1/client/ai-agent/devices` endpoint and the generated
`listAIAgentDeviceRuntimes` query. A future nested wrapper may expose the same
operation as a `riido.aiAgent.devices.runtimes` chain, but the chain must be
generated from the OpenAPI operation rather than hard-coded from the Figma
annotation. The agent hover popover is rendered from existing agent profile
fields. The daemon stop modal and restart animation are client/desktop-local
behavior; this server does not add a protected SaaS `stop daemon` or
`restart daemon` endpoint for that screen.

`node-id=275-22731` also maps to `GET /v1/client/ai-agent/devices`. The server
returns ordinary device/runtime rows or empty arrays; the client decides whether
to render no-current-device runtime, no daemon, no installed runtime, other
device rows, or provider install cards. Provider install-card hover behavior,
external provider installation links, Windows app waitlist copy, and
marketing-consent button states are not generated AI Agent API operations in
this slice. `Q-CP-007` tracks whether the waitlist/marketing mutation belongs
in this control-plane API or an existing product/user-marketing system.

`node-id=164-50215` maps to existing agent bootstrap/update behavior plus the
read-model fields `AgentClientRecord.updated_at`, `model_id`, and
`model_label`. `node-id=134-6542` adds the client-facing create behavior:
`POST /v1/client/ai-agent/agents` returns the created `AgentClientRecord`,
derives `runtime_kind` from the selected runtime, and validates optional
`model_id` against the selected runtime's `RuntimeRecord.models` catalog. If
`model_id` is omitted, the server saves the selected runtime's default model.
The client can use `updated_at` for the list's update date and the
absolute-time tooltip shown in Figma. Row click, meatball edit entry,
save-button enablement, long-description truncation, dropdown layout, and
timestamp formatting remain client-owned.

`node-id=42-3014` maps to existing bootstrap/devices/create behavior plus one
explicit bootstrap field: `ClientBootstrapResponse.agent_templates`. The
templates give clients stable starter-agent names, descriptions, role labels,
thumbnail URLs, and instructions without hard-coding product copy in the
frontend. Selecting a template still creates a normal agent through
`POST /v1/client/ai-agent/agents`; direct setting uses the same create endpoint.
No-installed-AI branching is derived from `devices.runtimes` and does not add a
new SaaS command.

`node-id=236-29749` does not add a generated AI Agent client operation. Sign-up,
login, Google-auth terms consent, email/password validation, terms row default
state/click target, and member invitation are auth/team/client product surfaces.
The macOS app download CTA is a distribution route, not a provider install or
daemon lifecycle command. Windows launch notification and marketing-consent
variants remain tracked by `Q-CP-007`; this API must not expose a waitlist or
marketing mutation until that owning SSOT is chosen. Chat and progress-bar
animation references stay client presentation facts.

The `model` field from `node-id=164-50215` is implemented as a runtime-scoped
catalog projection. This repository must not hard-code model candidates as
generated enum values. It only mirrors the contracts decision
`runtime_model_catalog.v1`: clients render `RuntimeRecord.models`, send an
optional `model_id`, and receive the saved `model_id` plus `model_label` on
agent records.

## Boundary

This repository owns the mock HTTP behavior, API sub-DSL projection, generated
client drift gate, and future generated-client delivery workflow. It does not
own production persistence, daemon runtime probing, Terraform state, live AWS
evidence, final DNS naming, or direct edits to the target frontend repository.

## Testnet Smoke

The `ai-agent-client-testnet-smoke` GitHub Actions workflow is intentionally
separate from the local API/generator workflow. It calls the deployed testnet
ALB for:

- `GET /healthz`
- `GET /readyz`
- `GET /v1/client/ai-agent/bootstrap`
- `GET /v1/client/ai-agent/devices`
- `GET /v1/client/ai-agent/tasks/{task_id}/assignable-agents`
- `GET /v1/client/ai-agent/tasks/{task_id}/threads`
- `POST /v1/client/ai-agent/tasks/{task_id}/comments`
- `POST /v1/client/ai-agent/tasks/{task_id}/stop`
- `POST /v1/client/ai-agent/agents`
- `GET /v1/client/ai-agent/events?replay=1`
- `POST /v1/agents/{agent_id}/thread-progress`

The workflow reads the ALB base URL from a manual workflow input or the
`RIIDO_AI_SERVER_TESTNET_BASE_URL` repository variable, and the bearer token
from the `RIIDO_AI_SERVER_TESTNET_TOKEN` repository secret.
