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
- configurable snapshot and task-event outbox port calls after assignment
  mutations
- configurable assignment operation journal save/replay/claim calls after
  assignment mutations
- durable active-assignment lease reads and heartbeat refreshes when the
  configured operation store implements the lease/projection ports

This actor does not own HTTP assignment routes, SSE fan-out,
DynamoDB/EventBridge adapter payload construction, Terraform, AWS credentials,
or deployment evidence.

## Store Snapshot And File Outbox Boundary

The public store snapshot and file outbox boundary owns the stdlib-only
persistence adapters that can be verified without AWS credentials:

- `SnapshotStore`
- `StoreSnapshot`
- `StoreSnapshotTask`
- `FileStoreSnapshot`
- `OpenStoreWithConfig`
- `EventSink`
- `OutboxRecord`
- `FileOutbox`

`StoreSnapshot` is a point-in-time assignment store snapshot. It preserves
tasks, assignments, agent-assignment indexes, task event history, and the next
assignment/event sequence counters. `FileStoreSnapshot` writes that snapshot as
strict JSON using atomic replace. Loading rejects unknown fields, unsupported
schema versions, trailing JSON, blank task ids, blank assignment ids, and agent
assignment references that do not exist in the snapshot assignment set.

`FileOutbox` appends task events as JSON Lines `OutboxRecord` values. The store
actor calls the outbox after task events are appended for assignment queue,
lease, and agent-event mutations. Outbox append errors do not fail the
assignment mutation; they increment the public `outbox_errors_total` metrics
counter and still record event-append latency counters.

This boundary does not own `DynamoDBStoreSnapshot`, `DynamoDBOutbox`, DynamoDB
Streams relays, EventBridge publishers, assignment operation durable save/claim
adapter implementation, Terraform, AWS credentials, Docker image contracts,
review account seed data, or deployment evidence.

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

## Health/Ready And Runtime Command Boundary

The public health/ready adapter exposes liveness and readiness responses that do
not require request authorization:

- `GET /healthz`
- `GET /readyz`

Both routes return the `Health` DTO with the current control-plane schema
version. Non-`GET` methods must fail with `405`.

`cmd/riido_ai_server` is the minimal stdlib-only runtime entrypoint for this
public repository. It owns only these environment variables:

- `RIIDO_AI_SERVER_ADDR`
- `RIIDO_AI_SERVER_SHUTDOWN_TIMEOUT_SECONDS`
- `RIIDO_AI_SERVER_AGENT_BINDINGS_JSON`
- `RIIDO_AI_SERVER_AUTHZ_TOKENS_JSON`
- `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_URL`
- `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_AUDIENCE`
- `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_TIMEOUT_SECONDS`

The agent binding and static-token JSON values use strict decoding, so unknown
fields and trailing JSON are rejected. Static-token authorization may be
combined with the external HTTP authorizer through the existing fallback
authorizer rule: only unauthenticated results fall through to the next
authorizer, while forbidden results stop evaluation.

This boundary does not own legacy broad bearer-token compatibility,
snapshot/outbox stores, durable operation save/claim wiring, DynamoDB,
EventBridge, Terraform, AWS credentials, CloudWatch/Prometheus adapters, Docker
image contracts, review account seed data, production secrets, or deployment
evidence.

## Container Image Contract Boundary

The public container image contract owns the buildable `riido_ai_server`
artifact shape that can be verified without AWS credentials:

- `packaging/containers/riido_ai_server.Dockerfile`
- `packaging/containers/riido_ai_server_container.riido.json`
- `tools/containercontract`

The executable contract requires a two-stage Go build, `CGO_ENABLED=0`, the
`./cmd/riido_ai_server` package, a `scratch` final image, copied CA
certificates, `EXPOSE 8080`, `RIIDO_AI_SERVER_ADDR=:8080`, non-root
`65532:65532`, and `ENTRYPOINT ["/riido_ai_server"]`.

`tools/containercontract` is the stdlib-only verifier for
`riido-container-image-contract.v1`. It emits
`riido-container-image-contract-check.v1` evidence and may optionally validate a
private Fargate task-definition IR when another repository supplies that path.

This boundary does not own ECR repositories, image push permissions, immutable
image digest evidence, Terraform/Fargate task definitions, AWS credentials,
runtime secret values, production environment variables, or deployment
evidence. Those remain `riido-infra` responsibilities.

## Durable Operation Boundary

The assignment operation journal and claim-port contract is owned by
[`assignment-operation-journal.md`](assignment-operation-journal.md).

That boundary owns operation records, assignment projection records,
active-assignment lease records, and durable claim/read ports. It does not own
the store actor, HTTP routes, SSE, DynamoDB payload construction, Terraform, or
deployment evidence.

It also owns the pure operation replay reducer that reconstructs internal
assignment projection state from operation records before a later store actor
slice consumes it. RIID-4681 is that store actor runtime-consumption slice:
the store actor can now save operation records, replay them when no snapshot is
available, claim the next assignment through an `AssignmentClaimer`, and consult
durable active-assignment lease/projection ports during poll and heartbeat.

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

RIID-4679 moves health/ready routes and the minimal `cmd/riido_ai_server`
environment/runtime entrypoint into this public repository.

RIID-4680 moves stdlib-only store snapshot and file outbox adapters into this
public repository.

RIID-4681 wires durable assignment operation journal ports into the public
store actor runtime without moving DynamoDB adapters or Terraform.

RIID-4682 moves the public Docker image contract, Dockerfile, container contract
verifier, and focused CI into this repository. ECR push, Terraform/Fargate task
definition IR, image digest deployment evidence, AWS credentials, and runtime
secret values remain private infra responsibilities.

Review account seed, ECR push, Terraform, AWS adapters, image digest evidence,
and production deployment evidence remain separate migration units.
