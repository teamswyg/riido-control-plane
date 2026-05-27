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
   daemon and control-plane need them. RIID-4688 completes this handoff for
   assignment polling DTOs by consuming `riido-contracts v0.3.0`.

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

### RIID-4688 — riido-contracts v0.3.0 assignment import migration

This slice replaces the local assignment contract SSOT with the tagged shared
contract from `riido-contracts`.

This slice does:

- update `github.com/teamswyg/riido-contracts` to `v0.3.0`
- import `github.com/teamswyg/riido-contracts/assignment`
- replace local assignment state, poll action, task event, and polling DTO
  declarations with aliases/imports
- replace `AgentRuntimeBinding` DTO declaration with the shared assignment
  contract type
- keep local wrapper predicates for existing store code while delegating to the
  shared contract predicates
- remove the local `assignment_contract.riido.json` fixture so the executable
  contract has one SSOT
- update public docs and the focused assignment contract workflow

This slice does not move store actor behavior, assignment HTTP routes, SSE,
metrics route wiring, health routes, authorization, provider status store/API,
daemon control-plane adapters, environment parsing, Docker, Terraform, secret
values, AWS adapters, or deployment evidence.

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

### RIID-4673 — assignment operation replay reducer migration

This slice moves the pure operation replay reducer into the public control-plane
repository.

This slice does:

- add the replay-relevant internal `storeState` projection shape
- add `stateFromAssignmentOperations`
- add operation replay ordering by last event sequence, recorded time, and
  operation id
- add duplicate event suppression by `(task_id, seq)`
- preserve same-sequence events across different tasks
- rebuild task current-assignment indexes
- rebuild agent assignment indexes
- update `next_event_seq` and `next_assignment_seq`
- extend `docs/20-domain/assignment-operation-journal.md`
- add focused assignment operation replay tests and public CI

This slice does not move the store actor loop, assignment queue/claim runtime,
active lease recovery runtime, assignment HTTP routes, daemon
poll/heartbeat/event handlers, SSE, metrics route wiring, snapshot store, file
outbox, DynamoDB/EventBridge adapters, Terraform, secret values, AWS adapters,
or deployment evidence.

### RIID-4674 — in-memory assignment store actor migration

This slice moves the stdlib-only in-memory assignment store actor into the
public control-plane repository.

This slice does:

- add the `AssignmentStore` port
- add `Store`, `StoreConfig`, `NewStore`, and `NewStoreWithClock`
- serialize assignment mutations through one goroutine-backed command channel
- preserve assign-task validation and agent runtime binding checks
- connect store-safe routing to synced in-memory provider status before
  assignment creation
- preserve reassignment cancellation handoff and blocker behavior
- preserve daemon poll actions for queued, active, and cancelling assignments
- refresh active assignment timestamps on daemon heartbeat
- validate agent event assignment FSM transitions and append task events
- expose metrics read-model counters for tasks, assignments, poll actions, and
  task events
- expose in-memory provider status sync/read through the existing ports
- add focused BDD/domain tests and public CI

This slice does not move HTTP assignment routes, SSE stream routes, snapshot
stores, file outbox adapters, assignment operation durable save/claim wiring,
stale durable active lease recovery, DynamoDB/EventBridge adapters, Terraform,
secret values, AWS adapters, `cmd/riido_ai_server`, Docker, or deployment
evidence.

### RIID-4675 — assignment HTTP adapter migration

This slice moves the assignment HTTP adapter into the public control-plane
repository.

This slice does:

- add `AssignmentStore` to `ServerConfig`
- auto-use assignment stores that also implement provider status ports for the
  existing provider status routes
- add `POST /v1/component-tasks/{task_id}/assignment`
- add `POST /v1/agents/{agent_id}/poll`
- add `POST /v1/agents/{agent_id}/heartbeat`
- add `POST /v1/agents/{agent_id}/events`
- gate each route with `RequestAuthorizer` resource/action scopes
- keep strict JSON decoding for assignment, poll, heartbeat, and event payloads
- add black-box HTTP BDD tests against the public in-memory store actor
- add a focused public CI workflow for assignment HTTP adapter behavior

This slice does not move `GET /v1/component-tasks/{task_id}/events` SSE,
`/metrics`, health/ready routes, `cmd/riido_ai_server` environment parsing,
snapshot stores, file outbox adapters, assignment operation durable save/claim
wiring, DynamoDB/EventBridge adapters, Terraform, secret values, AWS adapters,
Docker, or deployment evidence.

### RIID-4677 — task SSE stream adapter migration

This slice moves the component-task event SSE adapter into the public
control-plane repository.

This slice does:

- add `SubscribeTask(ctx, taskID)` to the `AssignmentStore` port
- add subscribe/unsubscribe commands to the in-memory store actor
- replay existing task event history on SSE connection
- support `replay=1` as a history-only response that does not hold the stream
  open
- stream new task events to active subscribers after agent event append
- add `GET /v1/component-tasks/{task_id}/events`
- gate the route with `RequestAuthorizer` using `component_task_events` /
  `events:read` scopes
- preserve SSE framing as `id`, `event`, and JSON `data`
- verify subscriber count metrics decrease after cancellation
- add black-box HTTP/domain tests and focused public CI

This slice does not move `/metrics`, health/ready routes,
`cmd/riido_ai_server` environment parsing, snapshot stores, file outbox
adapters, assignment operation durable save/claim wiring, DynamoDB/EventBridge
adapters, Terraform, secret values, AWS adapters, Docker, daemon SSE clients,
GUI SSE consumers, or deployment evidence.

### RIID-4678 — metrics HTTP adapter migration

This slice moves the `/metrics` HTTP adapter into the public control-plane
repository.

This slice does:

- add `GET /metrics`
- read `MetricsSnapshot` from `AssignmentStore.Metrics`
- gate the route with `RequestAuthorizer` using `metrics` / `read` scope
- return the `riido-ai-server-metrics.v1` JSON DTO
- verify assignment, poll, event, task-event, and state counters through a
  black-box HTTP test
- verify missing metrics scope is forbidden
- verify an unconfigured assignment store fails closed with 503
- add focused public CI for metrics HTTP adapter behavior

This slice does not move health/ready routes, `cmd/riido_ai_server`
environment parsing, CloudWatch EMF, Prometheus conversion, production tuning
metrics calibration, snapshot stores, file outbox adapters, assignment
operation durable save/claim wiring, DynamoDB/EventBridge adapters, Terraform,
secret values, AWS adapters, Docker, dashboards, daemon consumers, or
deployment evidence.

### RIID-4679 — health/ready and cmd env migration

This slice moves the public liveness/readiness HTTP routes and the minimal
runtime entrypoint into the public control-plane repository.

This slice does:

- add unauthenticated `GET /healthz`
- add unauthenticated `GET /readyz`
- return the `Health` DTO with the current control-plane schema version
- reject non-`GET` health/ready requests with `405`
- add `cmd/riido_ai_server`
- parse `RIIDO_AI_SERVER_ADDR` with default `:8080`
- parse `RIIDO_AI_SERVER_SHUTDOWN_TIMEOUT_SECONDS`
- parse strict `RIIDO_AI_SERVER_AGENT_BINDINGS_JSON`
- connect agent bindings to `StoreConfig.AgentRegistry`
- parse strict `RIIDO_AI_SERVER_AUTHZ_TOKENS_JSON`
- parse `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_URL`
- parse `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_AUDIENCE`
- parse `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_TIMEOUT_SECONDS`
- compose static-token and external HTTP authorizers through fallback authZ
- add black-box health/cmd parser tests
- add focused public CI for health/ready and command environment behavior

This slice does not move legacy `RIIDO_AI_SERVER_BEARER_TOKEN` compatibility,
snapshot stores, file outbox adapters, assignment operation durable save/claim
wiring, DynamoDB/EventBridge adapters, Terraform, secret values, AWS adapters,
CloudWatch/Prometheus adapters, Docker, review account seed data, dashboards,
daemon consumers, or deployment evidence.

### RIID-4680 — store snapshot/file outbox migration

This slice moves the stdlib-only store snapshot and file outbox adapters into
the public control-plane repository.

This slice does:

- add `SnapshotStore` and `EventSink` ports to the store configuration
- add `OpenStoreWithConfig` to restore an in-memory store from a snapshot
- add `StoreSnapshot`, `StoreSnapshotTask`, and
  `StoreSnapshotSchemaVersion`
- add `FileStoreSnapshot` with strict JSON decode, trailing-data rejection, and
  atomic replace writes
- preserve tasks, assignments, agent-assignment indexes, task event history,
  and next sequence counters in the snapshot
- save snapshots after assign, poll-start, and agent-event mutations
- close configured snapshot and outbox adapters when the store closes
- add `OutboxRecord`, `OutboxRecordSchemaVersion`, and `FileOutbox`
- append queued and leased task events as JSON Lines outbox records
- keep assignment mutations successful when outbox append fails
- expose outbox error and event-append latency counters through metrics
- add focused persistence tests and public CI

This slice does not move `DynamoDBStoreSnapshot`, `DynamoDBOutbox`, DynamoDB
Streams relay code, EventBridge publishers, assignment operation durable
save/claim runtime wiring, active lease recovery, Terraform, AWS credentials,
Docker image contracts, review account seed data, dashboards, daemon consumers,
or deployment evidence.

### RIID-4681 — durable operation runtime wiring migration

This slice wires the public assignment operation journal ports into the
stdlib-only store actor runtime.

This slice does:

- add `OperationStore` and `ActiveLeaseDuration` to `StoreConfig`
- save assignment operation records after assign, poll-start, and agent-event
  mutations
- replay assignment operation records from `AssignmentOperationLoader` when no
  snapshot is available
- claim the next assignment through `AssignmentClaimer` without duplicating an
  already-recorded poll-start operation
- load durable active-assignment leases and projections before returning
  `active` or `cancel` poll responses
- refresh durable active-assignment leases during heartbeat
- fail stale active assignments before leasing the next queued assignment
- close the configured operation store when the actor closes
- add BDD-style domain tests for save, replay, claim, active lease refresh,
  stale lease failure, and durable cancellation projection scenarios
- add focused public CI for the operation runtime wiring slice

This slice does not move DynamoDB assignment operation adapters, DynamoDB
Streams relay code, EventBridge publishers, Terraform, AWS credentials, Docker
image contracts, review account seed data, dashboards, daemon consumers, or
deployment evidence.

### RIID-4682 — Docker image contract migration

This slice moves the public, buildable control-plane container image contract
into the public control-plane repository.

This slice does:

- add `packaging/containers/riido_ai_server.Dockerfile`
- add `packaging/containers/riido_ai_server_container.riido.json`
- add `tools/containercontract` as a stdlib-only executable verifier for
  `riido-container-image-contract.v1`
- require a static `CGO_ENABLED=0` Go build of `./cmd/riido_ai_server`
- require a `scratch` final image with copied CA certificates
- require `EXPOSE 8080`, `RIIDO_AI_SERVER_ADDR=:8080`, non-root
  `65532:65532`, and `ENTRYPOINT ["/riido_ai_server"]`
- emit `riido-container-image-contract-check.v1` verification evidence
- add focused public CI for the contract verifier and Docker build

This slice does not move ECR repository creation, image push permissions,
immutable image digest evidence, Terraform/Fargate task definitions, AWS
credentials, runtime secret values, production environment values, private
deployment evidence, review account seed data, dashboards, or daemon consumers.

### RIID-4671 — provider status contract migration

This slice moves the provider status sync/read contract into the public
control-plane repository after `riido-contracts v0.2.0` exposed the shared
distribution/provider routing vocabulary.

This slice does:

- update the public contracts module to `v0.2.0`
- add `ProviderStatusRecord`, `ProviderStatusSyncRequest`, and
  `ProviderStatusSyncResponse`
- add the executable provider status normalization gate
- add `ProviderStatusStore` and `ProviderStatusReader`
- add `POST /v1/agents/{agent_id}/provider-status`
- add `GET /v1/agents/{agent_id}/provider-status`
- route write/read authorization through scoped agent provider-status actions
- reject unknown/private JSON fields with the existing strict decoder
- add `docs/20-domain/provider-status.md` as the public SSOT
- add focused provider status black-box/domain tests and public CI

This slice does not move daemon provider detectors, executable path/provenance
collection, store-safe routing decisions, assignment routing integration,
durable store actor behavior, review account seed runtime wiring, DynamoDB
adapters, Terraform, secret values, AWS adapters, or deployment evidence.

### RIID-4672 — store-safe routing guard migration

This slice moves the pure provider-status routing guard into the public
control-plane repository.

This slice does:

- add `StoreSafeRoutingInput` and `StoreSafeRoutingDecision`
- add `EvaluateStoreSafeRouting`
- allow `available`
- block `login-required`, `unsupported`, and `store-blocked`
- block a requested provider missing from an already-synced snapshot
- preserve legacy assignment behavior when no provider status snapshot exists
- reject blank runtime providers and unknown routing statuses
- extend `docs/20-domain/provider-status.md` with the routing guard boundary
- add focused store-safe routing domain tests and public CI

This slice does not connect the guard to assignment creation, daemon status
sync, review/demo seed runtime wiring, durable store actor behavior, DynamoDB
adapters, Terraform, secret values, AWS adapters, or deployment evidence.

### RIID-4691 — review account seed runtime wiring migration

This slice moves the public-safe store review/demo bootstrap path into the
public control-plane repository.

This slice does:

- add the embedded `riido-review-account-seed.v1` artifact
- add review seed loading, validation, and provisioning domain behavior
- provision a static-token credential from
  `RIIDO_AI_SERVER_REVIEW_ACCOUNT_TOKEN_SHA256` only
- keep the raw review token outside the repository and outside env parsing
- add in-memory `Store` actor commands for the `AgentCatalogStore` port
- let `NewServer` auto-use assignment stores that also implement the agent
  catalog and provider status ports
- seed owner/private, owner/public, other-user/public, and other-user/private
  agent catalog records for RBAC review
- seed a synthetic, non-routable provider-status snapshot for the
  `store-review-agent`
- wire `cmd/riido_ai_server` startup to apply review provisioning before
  serving HTTP
- add black-box HTTP tests proving the review token can read seeded catalog and
  synthetic provider status but cannot poll as a daemon agent
- add a focused public CI workflow for review seed domain, command wiring,
  import boundary, and no-raw-secret artifact checks

This slice does not move raw review token values, production IdP rollout, real
provider execution, Claude/Codex/OpenClaw/Cursor bundling, DynamoDB/EventBridge
adapters, Terraform, AWS credentials, production secrets, image digest
evidence, or deployment evidence.

### RIID-4692 — CloudWatch EMF metrics publisher migration

This slice moves the stdout CloudWatch Embedded Metric Format publisher into
the public control-plane repository.

This slice does:

- add a stdlib-only CloudWatch EMF metrics writer over `MetricsSnapshot`
- publish one EMF JSON Lines record immediately and then on a configured
  interval
- include assignment, poll, agent event, task event, SSE subscriber, outbox
  error, and event-append-latency counters
- parse `RIIDO_AI_SERVER_METRICS_LOG_INTERVAL_SECONDS` in `cmd/riido_ai_server`
- wire the optional metrics publisher to stdout during server startup
- keep the server shutdown path responsible for stopping the publisher
- add focused domain/command tests and public CI for the EMF slice

This slice does not move AWS SDK integration, CloudWatch PutMetricData,
credentials, log group/dashboard creation, production tuning samples,
Prometheus conversion, DynamoDB/EventBridge adapters, Terraform, AWS
credentials, private deployment evidence, or deployment evidence.

## Validation Gates

Required before a control-plane migration PR is mergeable:

```bash
go test ./...
go list -m all
go test ./internal/riidoaiserver -run 'ReviewAccount|HTTPReviewAccount|AgentCatalogStore' -count=1
go test ./cmd/riido_ai_server -run 'ReviewAccount|ConfigFromEnv|AuthorizerFromEnv' -count=1
go test ./internal/riidoaiserver -run 'CloudWatch|Metrics|EventAppend' -count=1
go test ./cmd/riido_ai_server -run 'MetricsLog|ConfigFromEnv|EnvOptionalDuration' -count=1
go test ./tools/containercontract -count=1
go run ./tools/containercontract -contract packaging/containers/riido_ai_server_container.riido.json -out -
docker build -f packaging/containers/riido_ai_server.Dockerfile -t riido-control-plane:local .
```

Server black-box tests should cover:

- daemon assignment polling
- event append and read model visibility
- SSE fan-out behavior
- RBAC: admin can view/edit/delete public and private agents
- RBAC: owner is admin only for owned agents
- RBAC: non-admin users see own agents plus other users' public agents

## Infra Boundary

`riido-control-plane` may build an image and verify the executable image
contract. Deployment is not owned here.

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
