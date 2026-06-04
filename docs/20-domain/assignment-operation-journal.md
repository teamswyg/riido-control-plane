# Assignment Operation Journal

> Riido task: RIID-4669 `[Control Plane] assignment operation journal port migration`

This document is the public SSOT for the assignment operation journal and
claim-port contract that can be verified without AWS credentials.

## Ownership

The assignment operation journal records control-plane assignment mutations as
append-only operation records. The public control-plane repository owns the
schema version strings, operation type values, operation record shape, claim
result shape, active-assignment lease shape, assignment projection shape, and
ports that durable adapters implement.

This document does not own the in-memory store actor, snapshot replay, HTTP
routes, SSE fan-out, DynamoDB request payloads, Terraform, AWS credentials,
secret values, or production deployment evidence.

RIID-4681 wires these ports into the public store actor runtime. That runtime
consumption does not move adapter ownership into this document: this document
continues to own the journal/claim/lease contract, while
[`saas-control-plane.md`](saas-control-plane.md) owns the store actor behavior.

## Port Surface

The public operation journal and claim ports are:

- `AssignmentOperationStore`
- `AssignmentOperationLoader`
- `AssignmentQueueReader`
- `AssignmentClaimer`
- `AssignmentActiveLeaseStore`
- `AssignmentProjectionReader`

These ports are adapter contracts. They do not choose DynamoDB, files, or any
other persistence substrate.

## Record Surface

The public operation records are:

- `AssignmentOperationRecord`
- `AssignmentProjection`
- `AssignmentActiveLease`
- `AssignmentClaimResult`

`AssignmentOperationRecord` carries the assignment snapshot plus the task events
created by the mutation. `validateAssignmentOperationRecord` is the executable
gate for required fields and assignment/event consistency.

`AssignmentActiveLease.Expired` treats `lease_expires_unix_ms` as the preferred
expiry source when it is present, falls back to `lease_expires_at`, and treats an
empty lease as expired.

## Replay Reducer Boundary

`stateFromAssignmentOperations` is the executable reducer for rebuilding the
in-memory assignment projection from durable operation records. It is public
repository code because the replay behavior must be testable without AWS.

The reducer must:

- validate each `AssignmentOperationRecord` before applying it
- replay records ordered by last event sequence, then `recorded_at`, then
  `operation_id`
- de-duplicate events by `(task_id, seq)` so repeated operation reads do not
  duplicate a task event
- keep events with the same sequence when they belong to different tasks
- track `next_event_seq` and `next_assignment_seq`
- sort replayed task events by sequence
- rebuild `task_id -> current_assignment_id` using newest assignment ordering
- rebuild `agent_id -> assignment_id[]` in assignment creation order

The reducer returns `storeState`, which is still an internal projection shape.
It is not an adapter, database schema, HTTP DTO, or AWS persistence contract.

## Version Strings

The public schema constants are:

- `AssignmentOperationSchemaVersion`
- `AssignmentProjectionSchemaVersion`
- `AssignmentAgentActiveSchemaVersion`
- `DefaultAssignmentActiveLeaseSeconds`

The string values live in code and must match this document's ownership. The
default active-assignment lease is 20 seconds. This matches the shared
assignment polling contract: daemon heartbeats should arrive every 5 seconds,
and a 20 second gap means the active daemon/runtime lease is stale. A stale
lease is failed before later queued work for the same agent can be claimed, and
heartbeat must not refresh an already-stale lease.

A future adapter migration may add wire-level fixtures, but this slice keeps the
contract in the package where only the control plane consumes it.

## Migration State

RIID-4669 moves the operation journal port and record surface from the former
private `riido_daemon/internal/riidoaiserver` package into this public
repository.

RIID-4673 moves the `stateFromAssignmentOperations` replay reducer and internal
projection rebuild helpers into this public repository.

RIID-4681 wires the operation journal save/replay/claim ports and
active-assignment lease/projection ports into the public store actor runtime
without adding AWS or other external dependencies.

Assignment HTTP/SSE routes, DynamoDB assignment operation adapter, stream relay,
EventBridge publisher, Terraform, and production evidence remain separate
migration units.
