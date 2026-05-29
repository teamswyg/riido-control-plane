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
  `instruction`, including URL-only thumbnail policy and the 1000 character
  instruction limit.
- This repository owns PATCH validation, save/update behavior, response
  projection, generated TypeScript shape, and smoke-test coverage.
- `riido-daemon` owns runtime consumption of an assigned instruction value after
  the assignment contract carries it; this server doc does not define provider
  prompt placement.
- `riido-infra` owns deployment/storage changes only when the API requires new
  media storage, secrets, persistence topology, or release evidence.
- Figma menu placement (`node-id=156-19307`) is a client route affordance, not
  a server menu contract. This repository owns the protected data endpoints used
  after a client opens the AI Agent, runtime, or agent-management route.

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
- `POST /v1/client/ai-agent/tasks/{task_id}/comments`
- `POST /v1/client/ai-agent/tasks/{task_id}/stop`
- `GET /v1/client/ai-agent/agents/{agent_id}/editability`
- `PATCH /v1/client/ai-agent/agents/{agent_id}`
- `DELETE /v1/client/ai-agent/agents/{agent_id}`
- `GET /v1/client/ai-agent/events`

The SSE endpoint supports `?replay=1` for deterministic smoke checks. Without
`replay=1`, it keeps the connection open as a client event stream.

## Policy

- visible agents are the viewer's owned agents plus other users' public agents
- task participant dropdown responses are ordered owned-first, then by name
- non-owner, non-admin users cannot mutate other users' public agents
- editing is blocked while `assigned_task_count` is greater than zero
- `profile_thumbnail_url` is saved as an optional HTTPS image URL string on the
  agent record; binary image upload/storage is outside this mock API
- `instruction` is saved as optional client-authored agent guidance text and is
  rejected when it exceeds 1000 characters
- delete returns forced assignment effects for queued/running mock tasks
- task-thread comments can enqueue work when the selected agent is busy
- task-thread stop actions return `stopped_by_user_request`
- task-thread status updates use typed `AgentTaskCommentKind` values
- daemon progress ingest accepts parsed `<riido_log>...<end>` batches through
  `POST /v1/agents/{agent_id}/thread-progress`
- client task-thread progress is streamed as the typed
  `agent_thread_progress` event on `GET /v1/client/ai-agent/events`

## Figma Handoff Evidence

Confirmed in Chrome against `v.1.22 AI Agent` on 2026-05-28:

- `node-id=236-21379`: normal task thread with comment input and agent reply
- `node-id=153-8761`: queued task comment when the agent is already busy
- `node-id=227-19354`: task stop flow with stopped agent comment
- `node-id=156-19307`: AI Agent menu placement in the workspace sidebar,
  including `Menubar/default` and `Menubar/setting` dark/light variants

`node-id=156-19307` does not add a new endpoint. It confirms that the frontend
needs visible entry points into AI Agent/runtime/agent-management surfaces; the
current server responsibility remains `bootstrap`, `devices`, agent mutation,
task-thread actions, and `events`.

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
- `POST /v1/client/ai-agent/tasks/{task_id}/comments`
- `POST /v1/client/ai-agent/tasks/{task_id}/stop`
- `GET /v1/client/ai-agent/events?replay=1`
- `POST /v1/agents/{agent_id}/thread-progress`

The workflow reads the ALB base URL from a manual workflow input or the
`RIIDO_AI_SERVER_TESTNET_BASE_URL` repository variable, and the bearer token
from the `RIIDO_AI_SERVER_TESTNET_TOKEN` repository secret.
