# SaaS Control Plane SSOT

> Riido task: RIID-4668 `[Control Plane] assignment contract/type migration`

This document is the public SSOT for the SaaS control-plane assignment contract
surface that can be verified without AWS credentials.

## Responsibility

`riido-control-plane` owns the server-side contract for assigning component
tasks to SaaS agent identities, daemon polling actions, daemon heartbeat
payloads, agent event sync payloads, task events, health responses, and metrics
snapshots.

This document does not own customer-PC provider process execution, local daemon
configuration, Terraform, AWS, EventBridge, DynamoDB, private deployment
evidence, or production secret values.

## Executable Contract

The executable assignment contract is
`internal/riidoaiserver/assignment_contract.riido.json`.

That contract owns:

- `service_schema_version`
- assignment state values
- assignment terminal classification
- assignment agent-active classification
- legal assignment transitions
- daemon poll action values
- task event type values

The generated Go surface in
`internal/riidoaiserver/assignment_contract_gen.go` must match the JSON
contract. Markdown must link to the executable contract instead of redefining
the transition matrix.

## Public DTO Surface

The public assignment DTO surface is:

- `AssignRequest`
- `Assignment`
- `PollRequest`
- `PollResponse`
- `AgentHeartbeatRequest`
- `AgentHeartbeatResponse`
- `AgentEventRequest`
- `AgentEventResponse`
- `TaskEvent`
- `Health`
- `MetricsSnapshot`

These types are API/adapter contracts only. They do not own HTTP routing, SSE
fan-out, outbox, snapshots, or AWS adapters.

## Store Actor Boundary

`internal/riidoaiserver.Store` is the public in-memory assignment actor. It owns
the stdlib-only runtime behavior that can be verified without AWS credentials:

- `AssignmentStore` command serialization through one actor goroutine
- assignment creation and reassignment cancellation handoff
- daemon poll actions (`none`, `start`, `active`, `cancel`)
- heartbeat-based active assignment timestamp refresh
- agent event transition validation and task event append
- metrics read-model counters for tasks, assignments, poll actions, and events
- in-memory provider status sync/read state used by store-safe routing

This actor does not own HTTP assignment routes, SSE fan-out, snapshot stores,
file outbox adapters, assignment operation durable save wiring, DynamoDB,
EventBridge, Terraform, AWS credentials, or deployment evidence.

## Assignment HTTP Adapter Boundary

The public HTTP assignment adapter routes request/response JSON into the
`AssignmentStore` port. It owns these stdlib-only routes:

- `POST /v1/component-tasks/{task_id}/assignment`
- `POST /v1/agents/{agent_id}/poll`
- `POST /v1/agents/{agent_id}/heartbeat`
- `POST /v1/agents/{agent_id}/events`

Every route must use `RequestAuthorizer` before it reaches the store. Request
bodies use the strict JSON decoder, so unknown fields are rejected and private
provider-path/token material cannot be accepted by accident.

This adapter does not own the task event SSE route, `/metrics`, health/ready
routes, `cmd/riido_ai_server` environment parsing, snapshot stores, file outbox
adapters, durable operation save/claim wiring, DynamoDB, EventBridge,
Terraform, AWS credentials, or deployment evidence.

## Task Event SSE Adapter Boundary

The public task event SSE adapter streams `TaskEvent` records from the
`AssignmentStore` subscription port. It owns this stdlib-only route:

- `GET /v1/component-tasks/{task_id}/events`

The route must use `RequestAuthorizer` with `component_task_events` /
`events:read` scope before subscribing. On connection it replays existing task
event history as SSE messages. With `replay=1`, the adapter flushes history and
returns without holding the stream open. Without `replay=1`, it keeps the stream
open and forwards later task events until the request context is cancelled.

The SSE adapter does not own `/metrics`, health/ready routes,
`cmd/riido_ai_server` environment parsing, snapshot stores, file outbox
adapters, durable operation save/claim wiring, DynamoDB, EventBridge,
Terraform, AWS credentials, daemon/GUI SSE consumers, or deployment evidence.

## Metrics HTTP Adapter Boundary

The public metrics HTTP adapter exposes the `MetricsSnapshot` read model from
the `AssignmentStore` port. It owns this stdlib-only route:

- `GET /metrics`

The route must use `RequestAuthorizer` with `metrics` / `read` scope before
reading the store. The response is the `MetricsSnapshot` DTO, including
`riido-ai-server-metrics.v1` schema version and the in-memory assignment,
poll, event, subscriber, outbox-error, and event-latency counters that are
available without AWS credentials.

The metrics adapter does not own health/ready routes, CloudWatch EMF,
Prometheus conversion, production tuning calibration, `cmd/riido_ai_server`
environment parsing, snapshot stores, file outbox adapters, durable operation
save/claim wiring, DynamoDB, EventBridge, Terraform, AWS credentials,
dashboards, daemon consumers, or deployment evidence.

## Durable Operation Boundary

The assignment operation journal and claim-port contract is owned by
[`assignment-operation-journal.md`](assignment-operation-journal.md).

That boundary owns operation records, assignment projection records,
active-assignment lease records, and durable claim/read ports. It does not own
the store actor, HTTP routes, SSE, DynamoDB payload construction, Terraform, or
deployment evidence.

It also owns the pure operation replay reducer that reconstructs internal
assignment projection state from operation records before a later store actor
slice consumes it.

## Provider Status Boundary

The provider status sync/read contract is owned by
[`provider-status.md`](provider-status.md).

That boundary owns provider status DTOs, normalization, read/write ports, and
the `GET`/`POST /v1/agents/{agent_id}/provider-status` HTTP adapter. It also
owns the pure store-safe routing guard that evaluates synced provider routing
status before a later assignment integration calls it. It does not own provider
executable detection, customer-PC provider process execution, durable store
actors, DynamoDB payloads, Terraform, or deployment evidence.

## Migration State

RIID-4668 moves the executable assignment contract and DTO surface from the
former private `riido_daemon/internal/riidoaiserver` package into this public
repository.

RIID-4669 moves the operation journal port and record surface into this public
repository.

RIID-4673 moves the assignment operation replay reducer into this public
repository.

RIID-4671 moves the provider status DTO/port/HTTP contract into this public
repository, using `riido-contracts v0.2.0` for shared provider/distribution
vocabulary.

RIID-4672 moves the pure store-safe routing guard into this public repository.

RIID-4674 moves the stdlib-only in-memory assignment store actor into this
public repository.

RIID-4675 moves the assignment HTTP adapter into this public repository.

RIID-4677 moves the task event SSE adapter into this public repository.

RIID-4678 moves the metrics HTTP adapter into this public repository.

Health/ready routes, snapshot stores, file outbox adapters, durable operation
save/claim wiring, review account seed, `cmd/riido_ai_server`, Docker,
Terraform, AWS adapters, and deployment evidence remain separate migration
units.
