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

## Device Identity & Visibility (machine device, workspace-connection scoped)

A device is an entity of the **physical machine**, not of an account. One machine
maps to exactly **one** device row, shared across the accounts and workspaces it
connects to.

- The desktop sends `machine_id` (the daemon's stable machine UUID, shared via
  the daemon `daemon.id` file) on `EnrollDeviceRequest`.
- The `DeviceID` is derived from **`machine_id` alone** (`deviceIDForMachine`), so
  every enrollment / snapshot of a machine resolves to the same device row.
  The `DeviceID` is not a secret — the rotating device secret is the auth factor.
  Callers without `machine_id` keep the legacy per-workspace behavior.
- A device tracks the set of workspaces it is connected to
  (`DeviceRecord.ConnectedWorkspaceIDs`). It **connects** to a workspace when it
  enrolls or reports a runtime snapshot in that workspace (auto-connect).
- Re-enroll/re-snapshot must not wipe runtimes already reported; the connected
  workspace set is unioned, not replaced.
- Legacy globally-colliding runtime IDs minted under the hardcoded
  `agentd-local` daemon id are pruned on snapshot restore
  (`pruneLegacyRuntimeRecords`); they predate per-machine UUIDs and are stale.

### Visibility

Device visibility is **workspace-connection scoped**, not owner scoped:

- A device is visible to **every member of any workspace it is connected to**.
  `GET .../workspaces/{ws}/ai-agent/devices` returns **all devices connected to
  `{ws}`** (`deviceConnectedToWorkspace` in `visibleDevicesLocked`).
- Isolation holds at the **workspace boundary**: a device is not visible in
  workspaces it is not connected to, nor to non-members.
- "This machine's own runtimes" (e.g. onboarding) is a separate,
  account/workspace-independent path: the desktop queries the **local daemon**
  (`riido daemon status`) directly, not the server device list.

Assigning/executing agents on a connected device from another workspace follows
the same workspace-connection model. The daemon authenticates with a single
device credential (its enroll workspace), but `GET /v1/daemon/agent-bindings`
returns bindings for agents in **every workspace the device is connected to**
(`DeviceRecord.ConnectedWorkspaceIDs`), not only the credential's enroll
workspace. The `binding.device_id` guard still restricts the result to agents
bound to **this** device's runtimes. Without this, an agent assigned from another
connected workspace is never surfaced to the daemon, so the daemon never polls it
and its assignment stays `queued` forever.

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
