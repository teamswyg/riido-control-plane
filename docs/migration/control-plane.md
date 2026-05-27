# Riido Control Plane Migration Plan

> Riido task: RIID-4638 `[Control Plane] 기존 riido_ai_server 마이그레이션 계획/문서화`

This document defines how the SaaS control-plane slice moves from the former
private `riido_daemon` repository into the public `riido-control-plane`
repository.

## Goal

`riido-control-plane` owns the SaaS backend behavior. It should be public and
testable without AWS credentials. Private deployment wiring stays in
`riido-infra`.

## Source In The Private Repository

The initial source paths are:

- `cmd/riido_ai_server`
- `internal/riidoaiserver`
- `docs/20-domain/saas-control-plane.md`
- `docs/30-architecture/riido-ai-server.md`
- `docs/40-decisions/0002-riido-ai-server-iac-is-terraform.md`
- `docs/40-decisions/0003-riido-ai-server-durability-is-dynamodb-outbox.md`
- server-related roadmap/audit files under `docs/50-roadmap/`
- server-specific workflows that do not require private AWS state

## Target Boundary

Move into `riido-control-plane`:

- assignment polling and claim logic
- event ingest and SSE/read-model behavior
- RBAC and request authorization ports
- durable store ports and stdlib-only adapters
- black-box HTTP/domain tests
- public Docker build contract, if it does not publish or deploy secrets

Do not move into `riido-control-plane`:

- Terraform modules
- AWS account, VPC, IAM, ACM, Route53, WAF, ECS, or ECR environment wiring
- Terraform state/backend config
- secret values, token material, private release evidence, or `.riido-local`
- daemon provider runtime code
- shared contracts that should be tagged from `riido-contracts`

## Migration Order

1. Move SaaS SSOT docs and ADR references.
   Keep the behavior docs public before moving the code that implements them.

2. Move pure domain/server core.
   Start with assignment/event/RBAC types and tests that do not need AWS.

3. Move HTTP/SSE adapter.
   Keep the server black-box testable with `net/http/httptest`.

4. Move durable store ports and stdlib-only adapters.
   AWS SDK integration remains outside the initial public migration unless a
   new ADR accepts that dependency.

5. Move generated assignment contract fixtures.
   Shared request/response fixtures should move to `riido-contracts` when both
   daemon and control-plane need them.

6. Add public CI.
   Public CI should run `go test ./...`, contract checks, RBAC black-box tests,
   and generated drift checks.

## Current Migration Slices

### RIID-4644 — contracts import gate

The first executable slice after the planning document is a compatibility gate
against `github.com/teamswyg/riido-contracts v0.1.0`.

This slice does:

- add the public contracts module to `go.mod`
- add a control-plane compatibility test that imports the IR, task lifecycle,
  and provider capability contracts
- replace the stdlib-only CI assertion with a "no non-Riido dependency" gate

This slice does not move `cmd/riido_ai_server`, `internal/riidoaiserver`,
AWS SDK integration, Terraform, production secret wiring, or deployment
evidence. Those remain separate migration units.

### RIID-4663 — agent catalog RBAC domain migration

This slice moves the first pure `internal/riidoaiserver` domain behavior into
the public control-plane repository.

This slice does:

- add `internal/riidoaiserver` agent catalog RBAC decisions
- add a package boundary comment for the first public `riidoaiserver` slice
- add static token request authorization and scope checks
- add BDD-style tests for admin, owner, public visibility, and mutation denial
- add `docs/20-domain/agent-catalog-rbac.md` as the SSOT for this domain slice
- add a focused public CI workflow for the migrated RBAC package

This slice does not move the HTTP server, SSE adapter, persistent store,
DynamoDB/EventBridge adapters, external identity authorizer HTTP adapter,
review account seed, Terraform, production credentials, or deployment evidence.

Later control-plane migration units build on this by moving store ports and the
HTTP agent catalog adapter while keeping AWS adapters out of public pull-request
verification.

### RIID-4664 — external HTTP authorizer migration

This slice moves the stdlib-only external request authorizer adapter behind the
public `RequestAuthorizer` port.

This slice does:

- add `ExternalHTTPAuthorizer` and its request/response schema version
  constants
- add fail-closed tests for 401, 403, `allowed:false`, malformed JSON,
  unsupported schema, invalid role, unsafe endpoints, and provider errors
- cover `FallbackAuthorizer` behavior where only unauthenticated results fall
  through to the next authorizer
- add `docs/20-domain/request-authorization.md` as the request authZ SSOT
- add a focused public CI workflow for the external authorizer adapter

This slice does not move production IdP rollout, tenant claim mapping,
JWKS/OIDC validation, environment parsing in `cmd/riido_ai_server`, HTTP route
wiring, Terraform, secret values, or deployment evidence.

### RIID-4665 — agent runtime binding migration

This slice moves the stdlib-only agent/runtime binding guard into the public
control-plane repository.

This slice does:

- add `AgentRuntimeBinding`, `AgentRegistry`, and `StaticAgentRegistry`
- add the minimal `PollRequest` fields required by daemon binding validation
- add focused tests for binding normalization, duplicate rejection, assignment
  provider mismatch, daemon ID mismatch, optional device ID behavior, and
  runtime ID mismatch
- add `docs/20-domain/agent-runtime-binding.md` as the binding SSOT
- add a focused public CI workflow for the agent runtime binding guard

This slice does not move the store actor, assignment queue, active lease
recovery, provider status store, HTTP handler, environment parsing, AWS
adapters, Terraform, or deployment evidence.

### RIID-4666 — agent catalog API port migration

This slice moves the public agent catalog DTOs and persistence port into the
public control-plane repository.

This slice does:

- add `CreateAgentCatalogRequest` and `UpdateAgentCatalogRequest`
- add `AgentCatalogListResponse` and `AgentCatalogRecordResponse`
- add `AgentCatalogStore` as a narrow list/get/save/delete port
- add tests that request DTOs do not accept owner or role input
- add tests that response DTOs include `schema_version` and catalog records
- extend `docs/20-domain/agent-catalog-rbac.md` with the API/port boundary
- add a focused public CI workflow for the agent catalog API/port slice

This slice does not move HTTP routing, authZ middleware, SSE, metrics, store
actor implementation, snapshot/operation replay, DynamoDB/EventBridge adapters,
environment parsing, Terraform, secret values, or deployment evidence.

### RIID-4667 — agent catalog HTTP adapter migration

This slice moves the stdlib-only agent catalog HTTP adapter into the public
control-plane repository.

This slice does:

- add the minimal `ServerConfig`, `Server`, and `Handler` boundary for agent
  catalog routes
- add `GET /v1/agent-catalog` and `POST /v1/agent-catalog`
- add `GET`, `PATCH`, and `DELETE /v1/agent-catalog/{agent_id}`
- call `RequestAuthorizer` before every catalog read or mutation
- stamp create owners from the authorization principal instead of request JSON
- reuse owner/public/private/admin RBAC decisions on read/update/delete
- reject unknown JSON fields and trailing request data
- add black-box HTTP tests for user, owner, admin, public-read, external
  authorizer, and create-owner stamping scenarios
- extend `docs/20-domain/agent-catalog-rbac.md` with the HTTP adapter boundary
- add a focused public CI workflow for the agent catalog HTTP slice

This slice does not move assignment polling, heartbeat, event append, SSE,
metrics, provider status HTTP routes, store actor implementation,
snapshot/operation replay, outbox, DynamoDB/EventBridge adapters,
environment parsing, review account seed, Terraform, secret values, or
deployment evidence.

### RIID-4668 — assignment contract/type migration

This slice moves the executable assignment contract and public assignment DTO
surface into the public control-plane repository.

This slice does:

- add `assignment_contract.riido.json`
- add generated assignment schema version, state, poll action, task event, and
  transition helpers
- move `PollRequest` into the broader assignment API DTO surface
- add `AssignRequest`, `Assignment`, heartbeat, poll response, agent event,
  task event, health, and metrics DTOs
- add tests that the JSON contract matches generated constants, terminal
  classification, active classification, and legal transitions
- add tests that assignment DTO JSON shapes match the former private contract
- add `docs/20-domain/saas-control-plane.md` as the public SSOT for this slice
- add a focused public CI workflow for the assignment contract/type slice

This slice does not move the store actor, assignment queue/claim/lease logic,
assignment HTTP routes, daemon poll/heartbeat/event handlers, SSE, metrics
route wiring, provider status store/API, review account seed, environment
parsing, Docker, Terraform, secret values, AWS adapters, or deployment
evidence.

### RIID-4669 — assignment operation journal port migration

This slice moves the durable assignment operation journal and claim-port
contract into the public control-plane repository.

This slice does:

- add assignment operation, projection, and active-assignment schema constants
- add assignment operation type constants
- add `AssignmentOperationStore`, `AssignmentOperationLoader`,
  `AssignmentQueueReader`, `AssignmentClaimer`, `AssignmentActiveLeaseStore`,
  and `AssignmentProjectionReader`
- add `AssignmentProjection`, `AssignmentActiveLease`,
  `AssignmentClaimResult`, and `AssignmentOperationRecord`
- add executable validation for operation records
- add deterministic helpers for operation IDs, event sequence extraction,
  assignment sequence parsing, queue sort keys, and active lease expiry
- add `docs/20-domain/assignment-operation-journal.md` as the public SSOT for
  this slice
- add a focused public CI workflow for the assignment operation journal slice

This slice does not move `stateFromAssignmentOperations`, store actor replay,
assignment queue/claim runtime behavior, active lease recovery runtime behavior,
assignment HTTP routes, daemon poll/heartbeat/event handlers, SSE, metrics
route wiring, provider status store/API, review account seed, environment
parsing, Docker, Terraform, secret values, AWS adapters, or deployment
evidence.

## Validation Gates

Required before a control-plane migration PR is mergeable:

```bash
go test ./...
go list -m all
```

Server black-box tests should cover:

- daemon assignment polling
- event append and read model visibility
- SSE fan-out behavior
- RBAC: admin can view/edit/delete public and private agents
- RBAC: owner is admin only for owned agents
- RBAC: non-admin users see own agents plus other users' public agents

## Infra Boundary

`riido-control-plane` may build an image and publish a release artifact.
Deployment is not owned here.

`riido-infra` consumes immutable artifacts by:

- Git tag
- module version
- container image digest

The control-plane repository must not need AWS credentials for normal pull
request verification.

## Contract Boundary

Use `riido-contracts` for shared API/fixture facts only after duplication exists
or after daemon and control-plane both consume the same schema. Until then,
keep implementation-local types inside `riido-control-plane`.
