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

## Migration State

RIID-4668 moves the executable assignment contract and DTO surface from the
former private `riido_daemon/internal/riidoaiserver` package into this public
repository.

The store actor, assignment claim logic, active lease refresh, HTTP
assignment/poll/heartbeat/event routes, SSE, metrics route wiring, provider
status routes, review account seed, `cmd/riido_ai_server`, Docker, Terraform,
AWS adapters, and deployment evidence remain separate migration units.
