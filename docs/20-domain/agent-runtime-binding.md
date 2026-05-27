# Agent Runtime Binding

> Riido task: RIID-4665 `[Control Plane] agent runtime binding migration`

This document is the SSOT for the public control-plane agent/runtime binding
domain slice.

## Ownership

Agent runtime binding keeps the SaaS control plane from accepting work claims
or daemon polls from a runtime identity that is not bound to the requested
agent.

This slice owns:

- immutable static agent/runtime binding records
- binding normalization and duplicate-agent rejection
- assignment-time `runtime_provider` checks
- daemon poll/heartbeat identity checks for `daemon_id`, optional `device_id`,
  and `runtime_id`

This slice does not own:

- assignment queue state
- active lease refresh or stale lease recovery
- provider status persistence
- HTTP handler routing
- environment parsing in `cmd/riido_ai_server`
- AWS, DynamoDB, EventBridge, Terraform, or production secret wiring

## Binding Record

`AgentRuntimeBinding` contains:

- `agent_id`
- `daemon_id`
- optional `device_id`
- `runtime_id`
- `runtime_provider`

The DTO shape is imported from `github.com/teamswyg/riido-contracts/assignment`
as of RIID-4688. Registry storage, normalization, duplicate rejection, and
binding validation remain owned by this control-plane package.

`agent_id`, `daemon_id`, `runtime_id`, and `runtime_provider` are required.
Duplicate `agent_id` values are invalid after normalization.

## Assignment Binding Rule

When an assignment is created for an agent, the control plane may check the
agent registry before accepting the assignment.

The assignment is allowed only when:

- the agent exists in the registry
- requested `runtime_provider` matches the registry binding

If no registry is configured, this slice imposes no assignment binding gate.

## Daemon Binding Rule

When a daemon polls, heartbeats, or syncs agent-owned runtime state, the request
must match the registry binding.

The daemon request is allowed only when:

- the agent exists in the registry
- `daemon_id` is present and matches
- `runtime_id` is present and matches
- `device_id` matches when the binding has a non-empty `device_id`

If the binding has an empty `device_id`, the request `device_id` is not
constrained by this slice. If no registry is configured, this slice imposes no
daemon binding gate.

## Migration State

RIID-4665 moves `StaticAgentRegistry` and the binding validation helpers from
the former private `riido_daemon/internal/riidoaiserver` package into this
public repository.

RIID-4668 moves the broader assignment API DTO surface, including
`PollRequest`, into this repository. The full store actor, poll response
behavior, heartbeat behavior, event sync behavior, metrics route wiring, and
HTTP server API migrations remain separate units.

RIID-4688 changes `AgentRuntimeBinding` and assignment polling DTOs to consume
the tagged `riido-contracts v0.3.0` shared assignment contract while keeping
all binding validation behavior local to `riido-control-plane`.
