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

## Version Strings

The public schema constants are:

- `AssignmentOperationSchemaVersion`
- `AssignmentProjectionSchemaVersion`
- `AssignmentAgentActiveSchemaVersion`
- `DefaultAssignmentActiveLeaseSeconds`

The string values live in code and must match this document's ownership. A
future adapter migration may add wire-level fixtures, but this slice keeps the
contract in the package where only the control plane consumes it.

## Migration State

RIID-4669 moves the operation journal port and record surface from the former
private `riido_daemon/internal/riidoaiserver` package into this public
repository.

The `stateFromAssignmentOperations` replay reducer, store actor state rebuild,
assignment HTTP/SSE routes, DynamoDB assignment operation adapter, snapshot
store, outbox, stream relay, EventBridge publisher, Terraform, and production
evidence remain separate migration units.
