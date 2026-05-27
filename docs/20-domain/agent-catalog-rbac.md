# Agent Catalog RBAC

> Riido task: RIID-4663 `[Control Plane] agent catalog RBAC domain migration`

This document is the SSOT for the public control-plane agent catalog RBAC
domain slice.

## Ownership

`internal/riidoaiserver` owns the agent catalog authorization rules that can be
verified without a store, HTTP server, AWS account, or daemon process.

This slice owns:

- agent catalog record visibility
- agent catalog read/update/delete decisions
- static bearer-token authorization for testable request scopes
- conversion from authorization identity into an agent catalog principal

This slice does not own:

- persistent agent catalog storage
- DynamoDB, EventBridge, IAM, WAF, or Terraform wiring
- external identity-provider HTTP adapters
- daemon-side provider process execution
- production token values or credentials

## Roles

The only role recognized by this slice is `admin`.

An `admin` principal can read, update, and delete public and private agent
catalog records.

The owner of an agent catalog record is treated as an admin for that record
only. Ownership is defined by `principal_id == owner_principal_id`.

## Visibility

Agent records are either `public` or `private`.

The visible-record rule is:

- admin principals see every public and private record
- owners see their own public and private records
- non-admin non-owners see only other users' public records

The mutation rule is:

- admins can update and delete every record
- owners can update and delete only their own records
- public visibility never grants mutation permission

## Authorization Scopes

Static token authorization is a public-testable adapter for request-scope
checks. Token credentials are compiled from either a plaintext token or a
SHA-256 token hash. Production token values are not part of this repository.

The broader request authorization port and external-authorizer adapter are
owned by [`request-authorization.md`](request-authorization.md). Request scopes
gate endpoint access; this RBAC policy still re-evaluates owner, visibility,
and admin role before exposing or mutating an agent catalog record.

Agent catalog scope candidates are:

- `riido:*`
- `agent-catalog:*`
- `agent-catalog:read`
- `agent-catalog:write`
- `agent-catalog:create`
- `agent-catalog:*:<action>`
- `agent-catalog:<agent-id>:*`
- `agent-catalog:<agent-id>:<action>`

## BDD Scenarios

The public CI gate must cover these behavior scenarios:

- Given an admin principal, when records include public and private agents,
  then all records are visible and read/update/delete are allowed.
- Given a record owner, when the owner reads, updates, or deletes the owned
  record, then the action is allowed.
- Given a non-admin principal, when records include owned agents and another
  user's public/private agents, then the principal sees owned agents plus the
  other user's public agents only.
- Given a non-admin non-owner, when the other user's public agent is updated or
  deleted, then mutation is denied.
- Given a static token with an agent catalog or broader Riido scope, when the
  request action matches the scope candidates, then authorization succeeds.

## Migration State

RIID-4663 moves the stdlib-only RBAC and static-token authorization code from
the former private `riido_daemon/internal/riidoaiserver` package into this
public repository.

The HTTP handler, durable store, external authorizer, and AWS adapters remain
separate migration units.
