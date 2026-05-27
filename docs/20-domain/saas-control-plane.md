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

These types are API/adapter contracts only. They do not own the store actor,
assignment queue, active lease, HTTP routing, SSE fan-out, outbox, snapshots, or
AWS adapters.

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

The store actor runtime, assignment claim logic, active lease refresh, HTTP
assignment/poll/heartbeat/event routes, SSE, metrics route wiring, review
account seed, `cmd/riido_ai_server`, Docker, Terraform, AWS adapters, and
deployment evidence remain separate migration units.
