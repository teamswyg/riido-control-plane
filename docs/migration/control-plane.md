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
