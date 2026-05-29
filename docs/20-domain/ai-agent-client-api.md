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
- `riido-contracts` owns onboarding runtime-selectability semantics. This
  repository projects protected `DeviceRecord.runtimes` values through
  `GET /v1/client/ai-agent/devices` and validates selected `runtime_id` values
  when agents are created or updated.
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
- Figma stopped-by-deleted-agent screen (`node-id=227-19354`) confirms that
  deleting an agent with queued or running assignments projects stopped
  task-thread rows. This repository owns the executable delete behavior,
  `running_tasks_force_stopped` response count, task-thread cold collection
  projection, and `stopped_by_agent_deleted` generated enum value. It does not
  own the Korean display copy, Riido actor label, timestamp wording, hidden
  action state, avatar, or row layout.
- Figma participant dropdown annotations (`node-id=153-12742`) confirm the
  `assignable-agents` API consumption context. This repository owns only the
  visible AI Agent list, owned-first ordering, generated DTO shape, and black-box
  tests. It does not own member sorting, long-name truncation, max dropdown
  height, scrollbar width, checkbox layout, or the mixed member/agent visual
  composition.
- Figma additional planning section (`node-id=153-15935`) confirms the
  assignment target scope. This repository owns task/subtask-scoped generated
  behavior under `/v1/client/ai-agent/tasks/{task_id}/...`. It does not expose
  project, milestone, intake, existing AI property filler, or agent mention
  operations from this screen. Those surfaces need a separate owning SSOT and a
  new generated operation before the server accepts them.
- Figma runtime settings annotations (`node-id=162-23090`) confirm the
  `devices` API consumption context. This repository owns the protected
  device/runtime read model, `device_runtime_snapshot` event shape, generated
  DTOs, and black-box tests. It does not own the agent hover popover, daemon
  stop modal, restart animation, or desktop-local daemon lifecycle command
  composition.
- Figma onboarding annotations (`node-id=42-3014`) confirm the bootstrap and
  device/runtime consumption context. `node-id=137-6746` maps runtime selection
  to `GET /v1/client/ai-agent/devices`: Claude Code/Codex can be rendered as
  detected/selectable rows when their runtime records are online and detected,
  while OpenClaw/Cursor Agent can be rendered as non-detected disabled rows.
  `node-id=138-7389` maps starter-agent selection to
  `ClientBootstrapResponse.agent_templates`: the mock/bootstrap catalog exposes
  the `리도`, `영실`, `홍도`, and `지원` starter templates in order, while
  `직접 설정`, disabled-next state before selection, row selection, and preview
  skeleton/popover rendering are client presentation.
  This repository owns the protected bootstrap projection of `agent_templates`,
  protected runtime read-model projection, selected `runtime_id` validation, and
  mock coverage. It does not own workspace selection, workspace list scrolling
  or the `새 워크스페이스` row shown in `node-id=164-30192`, runtime radio
  rendering, detected/non-detected Korean labels, row dimming, direct-setting
  row/rendering, disabled-next presentation, scrolling, two-line ellipsis,
  preview skeleton/popover layout,
  all-disconnected provider-list rendering and the `시작하기` CTA shown in
  `node-id=164-30206`, or the client decision to skip template selection when
  no selectable runtime exists.
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
- AI Agent assignment targets are task and subtask surfaces only; `task_id`
  path parameters in generated AI Agent APIs represent a task or subtask id
- project, milestone, intake, AI property filler, and agent mention surfaces
  must not call the task-scoped assignable-agent, comment, stop, or thread
  endpoints; future target surfaces require a separate SSOT and generated API
- bootstrap returns an ordered `agent_templates` catalog used by the Figma
  onboarding agent-template selection screen
- the current deterministic mock catalog returns the Figma `node-id=138-7389`
  starter templates in order: `리도`, `영실`, `홍도`, `지원`
- `직접 설정` is a client route into explicit agent creation, not a fifth
  `AgentOnboardingTemplate`
- runtime selection uses ordinary device/runtime records; SaaS validates the
  selected `runtime_id` through create/update and does not expose a separate
  runtime-selection mutation
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
- agent deletion uses the existing `DELETE /v1/client/ai-agent/agents/{agent_id}`
  command; if queued/running task threads are affected, the response increments
  `running_tasks_force_stopped` and the read model exposes
  `comment_kind=stopped_by_agent_deleted` with `assignment_state=stopped`
- task-thread comments can enqueue work when the selected agent is busy
- busy-agent enqueue responses use `comment_kind=queued_by_busy_agent`,
  `assignment_state=queued`, and `work_status=queued`; user-facing Korean copy
  is client presentation over those typed fields
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
- `node-id=153-15935`: additional planning section; visible planning text says
  only tasks and subtasks can receive Agent assignment, existing Riido AI
  property filling does not recommend agents, agent mentions are unsupported,
  and device/runtime actions are owner-only. No annotation nodes were found on
  this section, so the visible planning text is the evidence.
- `node-id=153-15931`: task-thread communication section; Dev Mode annotations
  include `riido.aiAgent.events.stream` on `node-id=153-8545` and
  `riido.aiAgent.tasks.stop` on `node-id=236-20768`
- `node-id=236-21379`: normal task thread with the generic task comment input,
  an AI Agent reply row, an AI Agent reply input, and a `중지` action in the same
  task view
- `node-id=153-8761`: queued task comment when the agent is already busy; the
  annotation on `node-id=153-8835` says "다른 작업 진행 중인 에이전트한테
  참여자 할당했을 때 나오는 댓글 문구"
- `node-id=227-19354`: stopped-by-deleted-agent task thread row; the screen
  shows Riido-authored copy that the agent was deleted and the running task was
  stopped
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
  (`node-id=138-7389`, `node-id=164-26969`), workspace selection/list scrolling
  and the `새 워크스페이스` row (`node-id=164-30192`), two-line template
  description ellipsis (`node-id=164-27719`), and no-installed-AI start behavior
  with all provider rows marked `연결 안 됨` (`node-id=164-30206`). The inspected
  `node-id=137-6746` screen shows Claude Code/Codex as `감지됨` selectable rows
  and OpenClaw/Cursor Agent as `감지 안 됨` non-selectable rows. The inspected
  `node-id=138-7389` screen shows `리도`, `영실`, `홍도`, and `지원` starter
  template rows, followed by `직접 설정`, with a disabled-looking `다음` button
  before selection and a right-side preview skeleton. The inspected
  `node-id=164-26969` expansion is annotated `직접 설정 선택 시 스크롤`; it dims
  starter-template rows and opens `이름`, `설명`, and `지침` inputs with
  placeholder copy.
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

`node-id=236-21379` confirms that the normal task screen consumes three
generated operations as one route composition: `tasks.threads` for historical
thread rows, `tasks.submitComment` for AI-Agent-directed thread replies, and
`tasks.stop` for the visible active-thread stop affordance. The right details
panel, generic task comment box layout, reply input presentation, and send
button state stay client/task surface behavior. The server only distinguishes an
AI-Agent-directed message once the client calls
`POST /v1/client/ai-agent/tasks/{task_id}/comments` with `agent_id` and optional
`source_comment_id`.

`node-id=153-8761` confirms the busy-agent queued branch of the same task-thread
composition. `tasks.submitComment` remains the creation command. When the
selected agent is already working, the server returns a queued task-thread row
with `comment_kind=queued_by_busy_agent`, `assignment_state=queued`, and
`work_status=queued`; `tasks.threads` later returns that row as part of the cold
collection, and `events.stream` may replay or stream the typed status change.
The visible Korean copy, "방금 전" timestamp, avatar, row layout, and other
comment presentation remain client-owned. The visible `중지` affordance still
maps to `tasks.stop`; the server must not expose a second queued-cancel endpoint
for this screen.

`node-id=227-19354` confirms the forced-stop projection after agent deletion.
The generated client command is still `riido.aiAgent.agents.delete`, backed by
`DELETE /v1/client/ai-agent/agents/{agent_id}`. When the deleted agent had
queued or running assignments, `DeleteAgentResponse.running_tasks_force_stopped`
reports the affected count. The same server-side effect updates task-thread
read models so `riido.aiAgent.tasks.threads` can return a stopped row with
`comment_kind=stopped_by_agent_deleted` and `assignment_state=stopped`; if a
viewer is connected, `riido.aiAgent.events.stream` may deliver the typed status
update. The server does not add a separate `deleted-agent stop` endpoint and
does not own the Korean copy, Riido actor label, "방금 전" timestamp, hidden
stop affordance, avatar, or row layout shown in Figma.

`node-id=153-15935` confirms that AI Agent target selection is not a general
workspace object feature. `riido.aiAgent.tasks.assignableAgents`,
`riido.aiAgent.tasks.submitComment`, `riido.aiAgent.tasks.threads`, and
`riido.aiAgent.tasks.stop` are task/subtask route composition only. A client
opening a project, milestone, intake, existing AI property filler, or mention
surface should not use these calls as a fallback. The control plane also must
not add placeholder endpoint families for those surfaces until a separate
owning SSOT names the target and policy.

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
frontend. `node-id=138-7389` is the template-selection initial state for that
field. Selecting a template still creates a normal agent through
`POST /v1/client/ai-agent/agents`; direct setting uses the same create endpoint
and is not represented as an extra template record.
The direct-setting expansion from `node-id=164-26969` maps those expanded
`이름`, `설명`, and `지침` inputs to
`CreateAgentConfigurationRequest.name`, `description`, and `instruction`.
Dimmed starter rows, placeholder copy, and scroll behavior remain client
presentation. The server still validates `runtime_id`, `visibility`, optional
profile image URL, and optional `model_id` through the normal create request.
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
