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
   External AWS SDK integration remains outside the public migration unless a
   new ADR accepts that dependency. Stdlib-only AWS JSON/SigV4 adapters may
   move when they are verified with local black-box HTTP tests and no live AWS
   credentials.

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
- parse `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_API_KEY`
- parse `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_TIMEOUT_SECONDS`
- compose static-token and external HTTP authorizers through fallback authZ
- add black-box health/cmd parser tests
- add focused public CI for health/ready and command environment behavior

This slice does not move legacy `RIIDO_AI_SERVER_BEARER_TOKEN` compatibility,
snapshot stores, file outbox adapters, assignment operation durable save/claim
wiring, DynamoDB/EventBridge adapters, Terraform, secret values, AWS adapters,
CloudWatch/Prometheus adapters, Docker, review account seed data, dashboards,
daemon consumers, or deployment evidence.

The external authorizer API key extends the same authorizer hop for production
integration with the existing Riido API server. It protects the server-to-server
authorizer request while the original browser request token remains opaque to
the control plane.

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

### RIID-4807 — tag-based AI Agent testnet CD

This slice moves the executable **testnet runtime artifact CD** into the public
control-plane repository without moving Terraform topology or secret values.

This slice does:

- add `.github/workflows/deploy-ai-agent-testnet.yml`
- trigger deployment only from `v*` tags or explicit manual dispatch
- use GitHub OIDC via the configured deploy-role secret
- build the checked-in `riido_ai_server` container image contract
- push an immutable ECR tag derived from the Git ref, commit SHA, and workflow
  run attempt
- resolve the pushed image to an ECR digest
- register a new ECS task-definition revision by changing only the configured
  container image
- update the configured ECS service and wait for service stability
- smoke `healthz`, `readyz`, and v2 workspace-scoped AI Agent bootstrap after
  deployment
- document the public secret/variable names without committing their values
- keep live URL, AWS account id, task-definition ARN, image digest, workflow run
  URL, smoke payload, Terraform state, plan output, and raw evidence out of
  checked-in public docs

This slice does not move ECR repository creation, ECS cluster/service topology,
IAM role topology, ALB/Route53/ACM/WAF resources, DynamoDB/EventBridge
resources, Terraform state, production secret payloads, or private live
evidence. Those remain `riido-infra` responsibilities. It also does not add
CodeDeploy blue/green; rolling ECS deployment is the current testnet CD model.

RIID-4812 tightens that public boundary: the deploy workflow may own runtime
artifact CD, but durable deployment evidence and environment values remain
private/operator-owned. Public docs name only configuration keys and behavior.

### RIID-4814 — CodeDeploy ownership and public redaction SSOT

This slice makes the runtime CD ownership rule explicit before any CodeDeploy
blue/green implementation work lands.

This slice does:

- add `docs/30-architecture/runtime-cd-ownership.md`
- add `docs/30-architecture/runtime-cd-ownership.riido.json`
- keep current testnet ECS rolling CD under the existing
  `deploy-ai-agent-testnet` workflow
- state that future CodeDeploy create/wait/smoke execution also stays in
  `riido-control-plane`
- require CodeDeploy application/deployment group, target groups/listeners,
  rollback policy, IAM, Terraform drift, and operator evidence to stay in
  `riido-infra`
- mask configured deploy target values in the public workflow before AWS calls
- keep live URL values, AWS account ids, ARNs, deployment ids, task-definition
  revision values, image digests, AppSpec/task-definition JSON, smoke payloads,
  Terraform plans, state, tfvars, apply logs, and raw evidence out of checked-in
  public artifacts
- avoid workflow_dispatch live URL inputs and avoid GitHub step outputs for
  image URI / task-definition ARN handoff; use masked `$RUNNER_TEMP` values
  inside the single deploy job instead

This slice does not create CodeDeploy AWS resources, alter Terraform topology,
change production traffic shifting, run a live deployment, or upload deployment
evidence artifacts.

### RIID-4815 — topology-gated CodeDeploy workflow mode

This slice turns the CodeDeploy handoff from documentation-only into a dormant
workflow mode while preserving the same ownership boundary. The default
`deploy-ai-agent-testnet` path remains ECS rolling deployment. If the optional
CodeDeploy application/deployment-group GitHub environment keys from the
machine-readable manifest are configured together, the workflow uses the
already-built immutable image and registered ECS task-definition revision to
create a CodeDeploy deployment, wait for deployment success, and run the same AI
Agent smoke checks.

Changed:

- add optional CodeDeploy application/deployment group variables to the public
  deploy workflow
- generate CodeDeploy AppSpec content and create-deployment request JSON only in
  same-job `$RUNNER_TEMP` files with restrictive permissions
- keep the CodeDeploy deployment id masked and stored only as a same-job temp
  value
- fail configuration early if exactly one CodeDeploy variable is present
- update the runtime CD ownership manifest, boundary docs, README, integration
  matrix, and deploy-policy tests

This slice does not create CodeDeploy application/deployment group resources,
blue/green target groups/listeners, IAM roles, rollback policy, Terraform
topology, live deployments, uploaded deployment artifacts, or checked-in live
values. Those remain `riido-infra` topology/evidence responsibilities before
the optional workflow mode can be configured.

### RIID-4822 — infra-output-gated CodeDeploy topology

This slice records the follow-up once `riido-infra` owns optional CodeDeploy
ECS blue/green topology. The public workflow ownership still does not move:
`riido-control-plane` creates/waits/smokes deployments, while `riido-infra`
creates/exposes the application and deployment group names after plan/apply
evidence.

Changed:

- update `runtime-cd-ownership.riido.json` so RIID-4822 is the current CD
  ownership task and RIID-4814/RIID-4815 are superseded history
- document that the optional workflow mode is infra-output-gated, not merely
  repo-gated
- keep service role ARNs, target group/listener ARNs, task-definition JSON,
  generated AppSpec/request JSON, deployment IDs, image digests, live URLs,
  smoke payloads, and environment-specific examples out of public docs,
  workflow inputs, reusable step outputs, and uploaded artifacts
- remove live host examples from generated-client comments

### RIID-4824 — testnet smoke URL redaction gate

This slice tightens the public redaction rule for the companion
`ai-agent-client-testnet-smoke` workflow. CD execution still belongs to
`riido-control-plane`, and `riido-infra` still owns topology/evidence only. The
public smoke workflow may verify the configured testnet target, but it must not
accept a live base URL through manual dispatch metadata.

Changed:

- remove the `base_url` `workflow_dispatch` input from
  `ai-agent-client-testnet-smoke`
- read the smoke target only from the configured GitHub environment variable
- mask both the smoke target and AI Agent token before issuing smoke requests
- extend the runtime CD ownership manifest and deploy-policy gate so both
  public workflows obey the same redaction rule

### RIID-4825 — CD ownership remodel and public redaction SSOT

This slice records the settled ownership model after the CodeDeploy discussion:
runtime artifact CD execution stays in `riido-control-plane`, while
`riido-infra` must know the topology/evidence contract without receiving public
workflow live values.

Changed:

- update `runtime-cd-ownership.riido.json` so RIID-4825 is the current
  ownership/redaction work unit
- add an explicit same-job handoff policy: image URI, task-definition ARN,
  container port, CodeDeploy AppSpec/request JSON, and deployment IDs are temp
  deploy implementation values only
- add an `always()` cleanup step for long-lived deploy temp files and require
  CodeDeploy generated files to stay under same-step traps
- document that infra consumes stable output names and redaction categories, not
  generated deploy payloads, image values, task-definition JSON, deployment IDs,
  or smoke payloads
- keep public docs focused on stable key names and behavior rather than
  environment-specific examples

This slice does not create or modify AWS resources, Terraform topology,
production traffic shifting, GitHub environment values, live deployment
evidence, or uploaded deployment artifacts.

### RIID-4833 — CD public redaction hardening

This slice hardens the RIID-4825 ownership model at workflow implementation
level. Runtime artifact CD execution still belongs to `riido-control-plane`,
and `riido-infra` still owns topology/evidence interpretation, but public
workflow temp payloads are made more local and short-lived.

Changed:

- set restrictive `umask 077` in deploy/smoke shell steps that write live
  handoff, task-definition, CodeDeploy, or smoke replay temp files
- explicitly `chmod 600` task-definition JSON files before they are reused by
  later commands
- remove companion smoke SSE replay capture with a same-step trap
- extend `runtime-cd-ownership.riido.json` and `tools/deploypolicy` so future
  workflow changes must preserve those file-permission and cleanup rules

This slice does not create or modify AWS resources, Terraform topology,
GitHub environment values, live deployment execution, release evidence files,
or uploaded workflow artifacts.

### RIID-4835 — CD public export contract

This slice narrows the CD ownership remodel into an explicit public export
contract. Runtime artifact CD execution and CodeDeploy create/wait/smoke remain
owned by `riido-control-plane`, while `riido-infra` must know only the stable
topology/evidence contract.

Changed:

- add `public_export_contract` to `runtime-cd-ownership.riido.json`
- state that public CD exports are limited to workflow names, stable secret and
  variable key names, stable infra output names, git identifiers, and aggregate
  pass/fail status
- state that public repos must not upload or output live deployment payloads,
  including URLs, AWS identifiers, image values, task-definition/AppSpec JSON,
  deployment IDs, smoke payloads, Terraform plans/state/tfvars/apply logs, or
  raw operator evidence
- extend `tools/deploypolicy` so the manifest, workflow, and docs preserve this
  public/export boundary

This slice does not create or modify AWS resources, Terraform topology,
GitHub environment values, live deployment execution, release evidence files,
or uploaded workflow artifacts.

### RIID-4836 — CD public surface redaction scan

This slice makes the public CD export boundary executable as a focused scan.
Runtime artifact CD execution, CodeDeploy create/wait/smoke, and public
redaction enforcement remain owned by `riido-control-plane`; `riido-infra`
remains the topology, IAM, Terraform drift, and private evidence owner.

Changed:

- add `public_surface_scan_contract` to
  `runtime-cd-ownership.riido.json`
- scan the explicit public CD surface for live host literals, AWS account
  number literals, checked-in ARN literals, live ALB/API Gateway/CloudFront URL
  literals, and public handoff mechanisms for live deploy values
- keep AWS CLI response field names and angle-bracket placeholder host examples
  allowed because they are non-live behavior descriptions
- state that infra must know the scan exists, but must not consume scan output
  as release evidence or a live workflow handoff

This slice does not create or modify AWS resources, Terraform topology,
GitHub environment values, live deployment execution, release evidence files,
or uploaded workflow artifacts.

### RIID-4837 — CD ownership final guard and public surface minimization

This slice closes the CD ownership remodel after the final ownership discussion.
Runtime artifact CD execution, CodeDeploy create/wait/smoke, and public redaction
gates stay in `riido-control-plane`. `riido-infra` must know the complete
hardening sequence, but it must still consume only stable names, redaction
categories, and private/operator evidence summaries.

Changed:

- add RIID-4837 to the runtime CD ownership hardening sequence
- extend the public surface scan to generated-client delivery docs and the
  generated React wrapper
- list every infra no-diff hardening work unit in the manifest's infra awareness
  path
- clarify that image digests and workflow run references can appear only in
  private/operator evidence summaries, not public workflow outputs, uploaded
  artifacts, checked-in examples, or generated client guidance

This slice does not create or modify AWS resources, Terraform topology,
GitHub environment values, live deployment execution, release evidence files,
uploaded workflow artifacts, or generated deployment payload handoffs.

### RIID-4839 — CD public config key minimization SSOT

This slice narrows the public CD ownership remodel into an explicit key-name
surface. Runtime artifact CD execution, CodeDeploy create/wait/smoke, and public
redaction gates still stay in `riido-control-plane`. `riido-infra` must know the
stable key categories, but public repositories should disclose only the minimum
`RIIDO_AI_SERVER_*` GitHub secret and variable names needed to configure deploy
and smoke.

Changed:

- add `public_config_key_minimization` to
  `docs/30-architecture/runtime-cd-ownership.riido.json`
- verify that deploy/smoke workflows reference only the allowed
  `RIIDO_AI_SERVER_*` GitHub configuration keys
- document that adding another public CD configuration key must update the
  runtime CD ownership manifest before workflow use
- keep key values, live examples, generated deploy payloads, image values,
  task-definition values, CodeDeploy generated JSON, deployment IDs, smoke
  payloads, and detailed evidence outside public repositories

This slice does not create or modify AWS resources, Terraform topology,
GitHub environment values, live deployment execution, release evidence files,
uploaded workflow artifacts, or generated deployment payload handoffs.

### RIID-4842 — CD owner SSOT and public sensitive surface guard

This slice keeps the settled remodel: runtime artifact CD execution and
CodeDeploy create/wait/smoke remain in `riido-control-plane`; `riido-infra`
owns topology, IAM, drift policy, and private/operator evidence. It adds the
extra rule that public CD configuration key names are a managed sensitivity
budget, not an open glossary.

Changed:

- add `public_sensitive_surface_guard` to
  `docs/30-architecture/runtime-cd-ownership.riido.json`
- record that new `RIIDO_AI_SERVER_*` public key names must update the
  ownership manifest before README, docs, or workflows reference them
- scan the public CD surface for unknown `RIIDO_AI_SERVER_*` key names, not only
  live values
- document that `riido-infra` consumes stable key categories/source names only
  and must not ask public workflows for live payload handoffs

This slice does not create or modify AWS resources, Terraform topology,
GitHub environment values, live deployment execution, release evidence files,
uploaded workflow artifacts, or generated deployment payload handoffs.

### RIID-4844 — CD public surface minimal exposure ratchet

This slice keeps the same remodel: runtime artifact CD execution and
CodeDeploy create/wait/smoke remain in `riido-control-plane`, while
`riido-infra` owns topology, IAM, drift policy, and private/operator evidence.
It narrows the public disclosure pattern by keeping the exact deploy/smoke
GitHub key list in the runtime CD ownership SSOT and workflow files instead of
repeating it in broad README/client-facing docs.

Changed:

- add RIID-4844 to `runtime-cd-ownership.riido.json` hardening tasks
- declare canonical CD key-list paths and broad docs that must link instead of
  listing deploy/smoke keys
- remove repeated deploy/smoke key names from README/client API prose while
  preserving operator pointers to the ownership SSOT
- extend deploy-policy tests so public summary docs cannot grow another CD key
  list by accident

This slice does not create or modify AWS resources, Terraform topology,
GitHub environment values, live deployment execution, release evidence files,
uploaded workflow artifacts, or generated deployment payload handoffs.

### RIID-4845 — CD public key-name docs minimization ratchet

This slice keeps the same ownership model: runtime artifact CD execution,
CodeDeploy create/wait/smoke, and public workflow redaction stay in
`riido-control-plane`; `riido-infra` owns topology, IAM, drift policy, and
private/operator evidence. It narrows public human-readable docs so exact
deploy/smoke key-name lists do not spread beyond the machine-readable manifest
and the workflow files that consume them.

Changed:

- add RIID-4845 to `runtime-cd-ownership.riido.json` hardening tasks
- keep the exact deploy/smoke key-name list canonical in the manifest and
  workflow files
- update runtime CD ownership, deployment-boundary, and migration prose to
  describe key categories and link to the manifest instead of repeating the list
- extend deploy-policy tests so these human-readable public docs cannot grow a
  duplicate CD key list again

This slice does not create or modify AWS resources, Terraform topology,
GitHub environment values, live deployment execution, release evidence files,
uploaded workflow artifacts, generated deployment payload handoffs, or workflow
key names.

### RIID-4853 — CD ownership settled public minimization guard

This slice records the settled CodeDeploy/CD remodel: runtime artifact CD
execution, CodeDeploy create/wait/smoke, and public workflow redaction stay in
`riido-control-plane`; `riido-infra` owns topology, IAM, drift policy, and
private/operator evidence interpretation. It narrows the public posture one
more time: even stable non-secret operational details are public only when they
are needed for workflow wiring, review, or operator setup.

Changed:

- add RIID-4853 to `runtime-cd-ownership.riido.json` hardening tasks
- add `public_operational_detail_minimization` as the canonical public
  disclosure posture for CD ownership
- link the expected infra awareness work unit so infra knows the SSOT without
  taking CD execution ownership
- update runtime CD ownership and deployment-boundary prose to tell broad public
  docs to link to the manifest when restating operational detail is unnecessary

This slice does not create or modify AWS resources, Terraform topology,
GitHub environment values, live deployment execution, release evidence files,
uploaded workflow artifacts, generated deployment payload handoffs, workflow
key names, or production traffic shifting.

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

### RIID-4704 — DynamoDB/EventBridge adapter migration

This slice moves the stdlib-only DynamoDB/EventBridge production adapter
behavior into the public control-plane repository while keeping live AWS
configuration and evidence private.

This slice does:

- add the DynamoDB request/signing helper and static/ECS credential provider
  boundary
- add `DynamoDBOutbox` for task-event outbox writes
- add `DynamoDBStoreSnapshot` for store snapshot load/save
- add the DynamoDB assignment operation durable store, queue claim,
  projection, and active-lease adapter
- add DynamoDB table stream discovery and DynamoDB Streams relay/checkpoint
  behavior
- add the EventBridge stream relay publisher adapter
- verify every AWS adapter through `httptest`/local black-box tests with fake
  endpoints and fake credentials only
- add focused public CI for the DynamoDB/EventBridge adapter test slice

This slice does not move AWS account ids, Terraform state/backend config,
tfvars, IAM/VPC/ECS/EventBridge rule resources, live AWS credentials, runtime
secret payloads, raw release evidence, live stream-relay evidence collection,
ECR image push evidence, dashboards, or private deployment evidence. It also
does not add the external Go AWS SDK.

### RIID-4706 — AWS adapter public facade migration

This slice adds a narrow public Go module facade for the RIID-4704
DynamoDB/EventBridge adapters so private infra tooling can consume the adapter
surface without importing `internal/riidoaiserver`.

This slice does:

- add the `awsadapters` package
- re-export only adapter DTOs, ports, constructors, and functions from
  `internal/riidoaiserver`
- add facade compile/usage tests for static credentials and stream-relay event
  types
- document that the facade must not fork or redefine adapter behavior
- add focused public CI for the facade package

This slice does not move implementation out of `internal/riidoaiserver`, add
the external Go AWS SDK, move live AWS evidence collection into public CI,
move Terraform, or expose credentials/runtime secret values.

### RIID-4712 — architecture SSOT docs migration

This slice restores the public architecture SSOT set for the split-repo
control-plane boundary.

This slice does:

- add `docs/20-domain/context-map.md` for public context ownership
- add `docs/30-architecture/module-decomposition.md` for package/import rules
- add `docs/30-architecture/config-reference.md` for `cmd/riido_ai_server`
  Factor 12 env ownership
- add `docs/30-architecture/integration-matrix.md` for public CI vs private
  infra/operator gates
- add `docs/30-architecture/runtime-deployment-boundary.md` for container,
  AWS adapter, and deployment hand-off facts
- add `docs/50-roadmap/open-questions.md` for unresolved public
  control-plane decisions
- add a focused public architecture-docs GitHub Actions workflow
- update domain docs and package comments to align with the RIID-4704/RIID-4706
  AWS adapter facade ownership

This slice does not move Terraform, AWS account wiring, live plan/apply
evidence, production secrets, daemon runtime code, or external AWS SDK
dependencies.

### RIID-4717 — web frontend API endpoint config

This slice configures the public control-plane HTTP API for browser-based
frontends without changing endpoint payload contracts.

This slice does:

- add `ServerConfig.WebAllowedOrigins` as the CORS transport allowlist
- handle `OPTIONS` preflight for the existing public HTTP endpoints
- allow only the methods and request headers needed by current API calls
- parse `RIIDO_AI_SERVER_WEB_ALLOWED_ORIGINS` as exact comma-separated browser
  origins in `cmd/riido_ai_server`
- document that CORS is transport configuration, not authorization
- add black-box HTTP tests and a focused public CI workflow for the web
  frontend API boundary

This slice does not add new product UI routes, change bearer-token
authorization, change agent-catalog RBAC, enable browser credentials, use
wildcard origins, add raw tokens, move production IdP rollout, or add external
dependencies.

### RIID-4721 — AI Agent client mock API and React Query generation

This slice wires the v1.22 AI Agent client contract into the public
control-plane repository as a mockable HTTP surface.

This slice does:

- add the checked-in AI Agent client DSL/IR/OpenAPI projection from
  `riido-contracts`
- add stdlib-only mock handlers for the AI Agent web/desktop webview endpoints
- enable the mock surface only when `RIIDO_AI_SERVER_AI_AGENT_CLIENT_MOCK=true`
- add black-box tests for bootstrap, devices, assignable agents, editability,
  task-thread cold collection, task-thread comment submit, task-thread stop,
  mutation, deletion, and SSE replay
- add `tools/reactquerygen` to generate `web/generated/aiAgentClient.ts` from
  the checked-in OpenAPI projection
- add a focused CI workflow for mock API and generated-client drift

This slice does not add production persistence, daemon runtime probing, final
Route53 records, IdP rollout, Terraform state, AWS credentials, or frontend app
implementation.

### RIID-4746 — AI Agent client API delivery SSOT

This slice defines how the control-plane AI Agent client API will be delivered
as generated React Query code to `riido-client` without editing the client
repository in this PR.

This slice does:

- define `riido-contracts` as canonical vocabulary and lifecycle owner
- define `riido-control-plane` as owner of the AI Agent client API sub-DSL,
  OpenAPI projection handoff, generated-client delivery boundary, and release
  manifest
- require business-meaning changes to escalate from control-plane to
  `riido-contracts`
- require generated-client delivery to run only from API release tags
- define the target client branch shape. This slice originally used
  `react-query-{tag}-{shortsha}`; A-60 supersedes that historical shape with the
  Riido task response `branchName`.
- define the target generated path allowlist
  `src/generated/react-query/riido-control-plane/**`
- require generated `apiHistory.generated.ts` and
  `contractManifest.generated.ts` artifacts to preserve lifecycle/deprecation
  context for frontend developers

This slice does not edit `teamswyg/riido-client`, run control-plane codegen in
the client repository, configure delivery secrets, publish npm packages, or
implement the cross-repository delivery workflow.

### RIID-4826 — riido-client/riido-desktop consumer boundary wording

This slice tightens the control-plane context map after the AI Agent client
consumer boundary became concrete.

This slice does:

- replace the old deferred-client context-map wording with `riido-client` web
  and `riido-desktop` webview as the current user-facing consumers
- state that those clients consume public HTTP contracts and generated AI Agent
  client artifacts while owning screen composition and route wiring
- keep `riido-control-plane` as the owner of HTTP/SSE behavior, OpenAPI
  projection, and generated-client delivery boundaries
- add a regression gate so stale future-client wording does not return to docs

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add deployment secrets, introduce live endpoint
examples, or move client UI ownership into this repository.

### RIID-4827 — bootstrap scenario wording SSOT projection

This slice mirrors the contracts-owned BDD wording clarification into the
control-plane generated-client projection.

This slice does:

- update the mirrored AI Agent client DSL/IR fixtures from `riido-contracts`
- keep the generated OpenAPI/API surface unchanged
- extend the projection regression gate so ambiguous future-client/bootstrap
  wording cannot remain in mirrored contracts or generated-client projection docs

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### RIID-4828 — task-thread v2 submitComment coverage projection gate

This slice tightens the top-down Figma coverage projection gate.

This slice does:

- mirror the contracts-owned Figma AI Agent coverage manifest into the
  AI Agent client contract mirror directory
- require each control-plane projection `required_generated_paths` entry to be
  named by the mirrored upstream coverage entry for the same Figma node
- absorb the upstream task-thread coverage fix for
  `v2.aiAgent.tasks.submitComment`

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### RIID-4829 — Figma non-UI page coverage projection

This slice absorbs the contracts-owned whole-file Figma coverage expansion into
the control-plane projection gate.

This slice does:

- update the mirrored Figma AI Agent coverage manifest from `riido-contracts`
- require the mirror to include the three inspected Figma pages and four
  non-UI top-level evidence nodes
- document that control-plane generated-client projection consumes whole-file
  coverage evidence but still only projects HTTP/SSE/generated-client behavior

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### RIID-4830 — Figma coverage inspection method projection

This slice mirrors the contracts-owned Figma inspection method so the
control-plane projection cannot accidentally treat metadata XML expansion as
page-level child-count drift.

This slice does:

- update the mirrored Figma AI Agent coverage manifest from `riido-contracts`
- require the mirror to preserve `figma.root.children` and
  `page.children.length` as the Figma Plugin API authority for page registry
  and top-level child counts
- document that metadata XML/read output is supporting evidence only
- extend the generated-client projection gate so the mirror and human doc agree
  on the inspection method

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### RIID-4832 — Figma lazy page load coverage projection

This slice mirrors the contracts-owned loaded-page Figma coverage correction.

The upstream contracts coverage now distinguishes passive page registry reads
from loaded page child counts. `figma.root.children` remains the page registry,
but non-current page top-level counts must come after
`await figma.setCurrentPageAsync(page); page.children.length`. This matters
because page `0:1` (`Wireframe`) can appear as one child in passive/lazy reads
but has 28 top-level children after loading the page.

This slice does:

- update the mirrored Figma AI Agent coverage manifest from `riido-contracts`
- preserve the loaded-page child-count rule in the control-plane projection doc
- require the projection gate to verify `non_ui_top_level_inventory` length
  against each mirrored page's loaded `child_count`
- keep generated-client coverage unchanged because this is an evidence/projection
  correction, not an API surface change

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### RIID-4834 — Figma API Generated annotation projection gate

This slice mirrors the contracts-owned Figma `API Generated` annotation
normalization.

This slice does:

- update the mirrored Figma AI Agent coverage manifest from `riido-contracts`
- preserve `api_generated_annotations` for Dev Mode category `700:0`
  `API Generated`, matching the mirror field names to the current category
  vocabulary
- normalize Figma facade examples such as `riido.aiAgent.events.stream` and
  `riido.aiAgent.tasks.stop` to canonical generated paths
  `aiAgent.events.stream` and `aiAgent.tasks.stop`
- require the projection gate to prove both the canonical generated path and the
  Korean generated-client access example exist in generated TypeScript/React
  comments
- document that `상세내용은 작업중입니다` is stale Figma handoff copy, not a new
  endpoint requirement

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### RIID-4838 — Legacy Wireframe semantic absorption projection

This slice mirrors the contracts-owned legacy Wireframe semantic-node coverage
gate from `teamswyg/riido-contracts#51`.

The upstream contracts manifest now promotes meaningful page `0:1`
(`Wireframe`) frames from inventory-only evidence to explicit coverage
decisions: runtime, agent, agent edit, agent add, daemon detail, and runtime
detail. Control-plane does not create new endpoints for those legacy frames.
Instead, this repo records that each legacy frame is absorbed by the current UI
coverage entry that already owns the generated-client surface.

This slice does:

- copy the updated contracts Figma coverage mirror
- add `legacy_non_ui_absorptions` to the local projection manifest
- require each legacy absorption to point to a covered upstream non-UI node and a
  covered current UI node
- require each inherited generated path to exist in OpenAPI and generated
  TypeScript/React comments

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### RIID-4840 — Figma metadata page-list limitation projection

This slice mirrors a contracts-owned Figma tooling limitation into the
control-plane projection gate.

On 2026-06-02, no-`nodeId` Figma `get_metadata` listed only `129:5215` `UI`,
while the Figma Plugin API page registry returned `129:5215`, `42:3014`, and
`0:1`. Control-plane needs this fact because the generated-client projection
must keep the whole-file coverage mirror, including non-UI inventories and
legacy Wireframe absorptions.

This slice does:

- copy the updated contracts Figma coverage mirror
- add `mirrored_supporting_tool_limitations` to the local projection manifest
- require the projection gate to prove the limitation exists in the upstream
  mirror and preserves the three authoritative page IDs
- document that no-`nodeId` metadata output must not remove `expected_pages`,
  `non_ui_top_level_inventory`, or `legacy_non_ui_absorptions`

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### Figma onboarding timeout provenance mirror catch-up

This slice consumes the contracts-owned provenance catch-up after the onboarding
page load timeout limitation was added.

`teamswyg/riido-contracts#60` changed executable Figma coverage meaning by
adding `figma-onboarding-page-load-timeout.v1` to
`supporting_tool_limitations`. The control-plane mirror must carry that entry
both in the mirrored source coverage fixture and in local
`source_contracts_manifest.stabilized_by`.

This slice does:

- copy the updated contracts Figma coverage mirror
- append `teamswyg/riido-contracts#60` to
  `source_contracts_manifest.stabilized_by`
- require the projection gate to expect the extended upstream provenance list

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### Figma API Generated annotation content policy mirror

This slice consumes `teamswyg/riido-contracts#62`, which records the live Figma
annotation content rule for generated-client handoff.

The mirrored contracts coverage now carries
`api_generated_annotation_content_policy`: every live `riido.*` Figma annotation
belongs to `700:0` / `API Generated`, keeps the facade path first, includes
`종류: Query | Mutation | SSE Stream`, and includes Korean `배경:` text. The live
inspection counts remain 53 annotations on the UI page, 6 on onboarding, and 0
on the legacy wireframe page, for 59 total generated handoff annotations.

This slice does:

- copy the updated contracts Figma coverage mirror
- append `teamswyg/riido-contracts#62` to
  `source_contracts_manifest.stabilized_by`
- require the projection gate to preserve the annotation content policy and
  page-level live inspection totals
- document that generated TypeScript comments still derive from OpenAPI and the
  generated-client manifest, not from hard-coded Figma label strings

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### Figma retired client-delivery annotation category mirror

This slice consumes `teamswyg/riido-contracts#63`, which records old Figma
category `39:0` / `클라이언트 전달` as retired and unused.

The control-plane mirror keeps that fact so generated-client delivery does not
accidentally treat the old category as an active handoff category again. The
current live usage count is 0. The category definition may still exist in Figma
because the current Figma MCP exposes category data without callable `remove` or
`setLabel` methods.

This slice does:

- copy the updated contracts Figma coverage mirror
- append `teamswyg/riido-contracts#63` to
  `source_contracts_manifest.stabilized_by`
- require the projection gate to preserve the retired category id, label,
  `unused_not_deleted` status, zero usage count, and tool limitation

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### Figma onboarding direct-node fallback evidence mirror

This slice consumes `teamswyg/riido-contracts#64`, which refines the onboarding
page-load timeout limitation with direct registered-node fallback evidence.

Full `42:3014` `Wireframe - 온보딩` traversal can still time out, but direct
Figma Plugin API reads for `236:33845` and `236:33847` preserve the six
onboarding `riido.*` `API Generated` annotations. Control-plane mirrors that
fact so generated-client projection does not treat a full-page timeout as proof
that onboarding generated paths disappeared.

This slice does:

- copy the updated contracts Figma coverage mirror
- append `teamswyg/riido-contracts#64` to
  `source_contracts_manifest.stabilized_by`
- require the projection gate to preserve `236:33845`, `236:33847`, and
  `onboarding_api_generated_annotations=6` in the source limitation

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### Figma headless file-key placeholder limitation mirror

This slice consumes `teamswyg/riido-contracts#65`, which records a Figma Plugin
API runtime limitation: live `use_figma` inspection can return
`figma.fileKey=headless` while still reading the real AI Agent file content.

Control-plane mirrors that evidence so generated-client projection never treats
the headless runtime placeholder as the contracts source identity. The
authoritative source remains the mirrored contracts manifest's `figma.file_key`
and the local `source_contracts_manifest` values.

This slice does:

- copy the updated contracts Figma coverage mirror
- append `teamswyg/riido-contracts#65` to
  `source_contracts_manifest.stabilized_by`
- add a local mirrored limitation that forbids replacing the upstream file
  identity with `headless`
- require the projection gate to preserve the source authoritative results
  `MUOd9lctoEHASUStN3vUuK` and `v.1.22 AI Agent`

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### Figma API Generated operation-kind transport guard mirror

This slice consumes `teamswyg/riido-contracts#66`, which tightens the contracts
Figma API Generated annotation guard.

The mirrored source policy now requires `operation_kind` to match generated
OpenAPI transport: `text/event-stream` responses are `SSE Stream`, non-stream
`GET` operations are `Query`, and non-`GET` operations are `Mutation`.
Control-plane mirrors that source rule so generated-client projection cannot
preserve a Figma annotation kind that contradicts the OpenAPI operation shape.

This slice does:

- copy the updated contracts Figma coverage mirror
- append `teamswyg/riido-contracts#66` to
  `source_contracts_manifest.stabilized_by`
- require the projection gate to preserve the transport-derived operation-kind
  rule in the mirrored source policy and human projection doc

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### Figma onboarding page load timeout limitation mirror

This slice mirrors the contracts-owned Figma onboarding page load timeout
limitation into the control-plane projection gate.

Current live reads of `node-id=42:3014` (`Wireframe - 온보딩`) can time out
after 120s when using Figma `get_metadata(nodeId=42:3014)` or `use_figma`
scripts that attempt `await figma.setCurrentPageAsync(page)`. Control-plane
must not treat that timeout as proof that onboarding generated-client coverage
disappeared.

This slice does:

- copy `figma-onboarding-page-load-timeout.v1` from the contracts Figma coverage
  mirror
- add a local `mirrored_supporting_tool_limitations` entry for the timeout
- require the projection gate to keep page `42:3014`, non-UI inventory, and
  onboarding generated paths covered despite the timeout

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### RIID-4900 — Figma API Generated handoff refresh mirror

This slice mirrors the latest contracts-owned Figma handoff audit.

The live `API Generated` annotation coverage did not add or remove generated
paths. All 90 `riido.*` annotations remain in the `API Generated` category with
operation kind and background handoff text. The mirrored source change is only
the page `42:3014` count split: `child_count=84`,
`known_inventory_count=83`, and `unresolved_extra_top_level_node=1`.

This slice does:

- copy the updated contracts Figma coverage mirror
- require the local projection limitation to preserve `child_count=84`,
  `known_inventory_count=83`, and `unresolved_extra_top_level_node=1`
- keep onboarding generated paths and API Generated annotation coverage
  unchanged

This slice does not edit `teamswyg/riido-client`, change OpenAPI, change
handlers, change authorization, change daemon runtime behavior, add Terraform,
or deploy AWS resources.

### Figma API Generated provenance mirror catch-up

This slice consumes the contracts-owned provenance catch-up after the API
Generated annotation passes.

`teamswyg/riido-contracts#56`, `#57`, and `#58` changed executable Figma
coverage meaning: #56 registered the screen-level API Generated annotation
inventory, #57 moved the Figma category to `700:0` / `API Generated`, and #58
renamed the manifest fields to `api_generated_annotations` and
`api_generated_annotation_inventory`. The control-plane mirror must carry those
entries both in the mirrored source coverage fixture and in local
`source_contracts_manifest.stabilized_by`.

This slice does:

- copy the updated contracts Figma coverage mirror
- append `teamswyg/riido-contracts#56`, `#57`, and `#58` to
  `source_contracts_manifest.stabilized_by`
- require the projection gate to expect the extended upstream provenance list

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### RIID-4841 — Figma coverage upstream provenance guard

This slice closes the provenance gap left after RIID-4840.

The control-plane mirror now contains the contracts-owned
`supporting_tool_limitations` entry introduced by `teamswyg/riido-contracts#52`.
However, the local projection manifest's `source_contracts_manifest.stabilized_by`
list still stopped at `teamswyg/riido-contracts#51`. That made the mirror
content newer than its recorded upstream stabilization boundary.

This slice does:

- add `teamswyg/riido-contracts#52` to
  `source_contracts_manifest.stabilized_by`
- require the projection gate to fail when the metadata page-list limitation is
  mirrored without the matching upstream stabilization PR
- document that upstream provenance is part of the mirror contract

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### RIID-4848 — Figma coverage upstream provenance full mirror guard

This slice closes the remaining downstream provenance gap after RIID-4841.

The local projection manifest already carried the full contracts coverage
history in `source_contracts_manifest.stabilized_by`, but the human document
and executable gate only required the metadata limitation slice
`teamswyg/riido-contracts#52`. That made the generated-client projection look
like a full mirror while the test only proved one local limitation provenance.

This slice does:

- require `source_contracts_manifest.stabilized_by` to equal the full upstream
  contracts coverage provenance:
  `teamswyg/riido-contracts#38`, `#39`, `#45`, `#46`, `#51`, and `#52`
- document that the full upstream coverage provenance is distinct from
  limitation-local provenance
- keep `figma-metadata-page-list-underreports-pages.v1` tied to
  `teamswyg/riido-contracts#52` while preserving the full source history used by
  the control-plane projection

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### RIID-4850 — Figma coverage provenance mirror source field sync

This slice consumes the contracts-owned provenance field introduced by
`teamswyg/riido-contracts#53`.

RIID-4848 made the local projection manifest require the full upstream coverage
history, but that list still lived only in the local projection manifest. After
contracts added top-level `stabilized_by` to the canonical Figma coverage
manifest, the control-plane mirror must copy that field and compare it with
local `source_contracts_manifest.stabilized_by`.

This slice does:

- add top-level `stabilized_by` to the mirrored contracts coverage fixture under
  `contracts/ai-agent-client/figma-ai-agent-coverage.riido.json`
- require the projection gate to fail when the mirrored source coverage
  `stabilized_by` and local `source_contracts_manifest.stabilized_by` diverge
- document that control-plane now mirrors the contracts field instead of
  preserving upstream history from local memory

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change generated API shape, add runtime behavior, add deployment secrets,
introduce live endpoint examples, or move client UI ownership into this
repository.

### RIID-4855 — CodeDeploy activation ownership gate SSOT

This slice records the final activation rule after the CodeDeploy ownership
discussion.

CodeDeploy blue/green topology can exist in `riido-infra`, but activation is
still a `riido-control-plane` runtime artifact CD workflow mode. Operators wire
the infra-provided stable application/deployment-group names into GitHub
environment configuration out of band; the public workflow then registers the
task-definition revision, creates/waits for CodeDeploy, and runs smoke in the
same job.

This slice does:

- add `codedeploy_activation_gate` to the runtime CD ownership manifest
- require the gate to keep `riido-control-plane` as canonical CD owner and
  `riido-infra` as awareness/topology owner
- define activation requirements without adding another public key-name list to
  human-readable docs
- tighten the public deny list around live CodeDeploy values, generated deploy
  payloads, image/task-definition values, smoke payloads, workflow-run evidence,
  and Terraform/operator evidence

This slice does not add AWS topology, change Terraform, add deployment secrets,
upload workflow artifacts, expose live environment values, edit generated
client API shape, or move CD execution into `riido-infra`.

### RIID-4825 follow-up — infra awareness back-reference guard

This follow-up keeps the settled CD ownership remodel readable from both sides
without widening the public surface.

The runtime CD ownership manifest already says `riido-control-plane` owns
runtime artifact CD and CodeDeploy create/wait/smoke execution, while
`riido-infra` owns topology, IAM, drift, and private/operator evidence. This
follow-up adds the latest infra-side no-diff awareness work units to the
manifest back-reference list so operators can trace the policy sequence:
RIID-4854 public minimization awareness, RIID-4856 CodeDeploy activation gate
awareness, and RIID-4860 infra-local awareness guard.

This follow-up does not add AWS topology, change Terraform, add GitHub
configuration keys, expose key values, export workflow payloads, upload
artifacts, publish live URLs, image/task-definition values, CodeDeploy
AppSpec/request JSON, deployment IDs, smoke payloads, Terraform evidence, or
move CD execution into `riido-infra`.

### Figma API Generated v2 counterpart mirror

This slice consumes `teamswyg/riido-contracts#67`, which makes the Figma API
Generated handoff guard preserve both the short facade path and the
workspace-scoped v2 generated path.

Control-plane still mirrors Figma labels such as `riido.aiAgent.tasks.stop`
because those labels are searchable handoff text for frontend developers. The
actual current API surface also exposes `riido.v2.aiAgent.tasks.stop`.
Therefore each mirrored API Generated annotation/inventory item must prove that
the canonical facade path and the `v2.*` counterpart both exist in OpenAPI, in
the same source coverage entry, and in generated TypeScript comments.

This slice does:

- mirror the contracts Figma coverage manifest through
  `teamswyg/riido-contracts#67`
- append `teamswyg/riido-contracts#67` to
  `source_contracts_manifest.stabilized_by`
- require generated core and React clients to keep Korean comments for both the
  facade path and the `riido.v2.*` access example

This slice does not edit `teamswyg/riido-client` or `teamswyg/riido-desktop`,
change runtime behavior, add deployment secrets, expose live endpoint values,
or move client UI ownership into this repository.

### Figma coverage mirror strict decode guard

This follow-up tightens the downstream Figma coverage mirror reader used by the
generated-client projection gate.

Control-plane keeps a copied contracts coverage manifest only to prove that the
local OpenAPI/generated-client projection is consuming the upstream SSOT. A
plain JSON unmarshal could silently ignore fields that contracts added to the
manifest shape, which would let this repository keep passing while consuming
only part of the upstream coverage record. The projection gate now decodes both
the local projection manifest and the mirrored contracts coverage manifest with
unknown-field and trailing-document rejection.

This slice updates the mirrored source coverage test type to include the full
contracts-owned fields that the current mirror carries, including `riido_task`,
`human_doc`, `related_manifests`, `figma`, `coverage_policy`,
`expected_top_level_nodes`, inspection authority/supporting tools, supporting
tool limitation provenance, and entry ownership/direction fields.

This slice does not edit API DSL/OpenAPI shape, generated client output,
frontend delivery branches, runtime behavior, Figma annotations, deployment
configuration, or live endpoint values.

### Runtime CD ownership strict manifest decode guard

This slice tightens the control-plane-owned runtime CD ownership SSOT. The
deploy-policy test now decodes
`docs/30-architecture/runtime-cd-ownership.riido.json` with unknown-field and
trailing-document rejection, so the public CD ownership manifest cannot silently
accept misspelled or unmodeled decision fields.

This slice does not edit API DSL/OpenAPI shape, generated client output,
frontend delivery branches, runtime behavior, Figma annotations, deployment
configuration, Terraform topology, GitHub environment values, or live endpoint
values.

### RIID-4872 — AI Agent development persistence

This slice promotes the AI Agent client surface from a throwaway in-memory mock
runtime to a development server runtime with durable state.

It adds:

- `PersistentAIAgentClientStore`, a snapshot-saving wrapper around the
  deterministic development store
- `AIAgentClientSnapshot` with schema version
  `riido-ai-agent-client-persistence.v2`
- `DynamoDBAIAgentClientSnapshot`, a stdlib-only DynamoDB `PutItem`/`GetItem`
  adapter that stores the development state as one `pk/sk` snapshot item
- runtime env parsing for `RIIDO_AI_SERVER_AI_AGENT_CLIENT_DEVELOPMENT`,
  `RIIDO_AI_SERVER_AI_AGENT_CLIENT_DYNAMODB_TABLE`,
  optional `RIIDO_AI_SERVER_DYNAMODB_ENDPOINT`, AWS region configuration, and
  ECS container credential endpoint variables
- compatibility handling for the older
  `RIIDO_AI_SERVER_AI_AGENT_CLIENT_MOCK` flag as an alias for development mode
- black-box tests proving created agents, task-thread events, and device
  credentials survive reopening the store
- DynamoDB fake-endpoint tests proving the snapshot item key and schema
  metadata

The snapshot stores only `device_secret` hashes, never one-time raw
`device_secret` values. Development mode now fails during startup unless the
DynamoDB snapshot table and signing credential path are configured.

This slice does not create AWS tables, change Terraform topology, implement
production single-table projections, rotate/revoke device credentials, alter the
OpenAPI/generated client surface, or edit `riido-client`.

### RIID-4881 — Generated assignment team/OpenAPI exclusion mirror

This slice mirrors the upstream contracts decision that generated AI Agent
assignment and DevicePrincipal polling must not depend on team/OpenAPI-key
inputs.

This slice does:

- document that private task-context lookup may resolve a task's location from
  `task_id`, but the resulting team value is transient lookup context
- keep generated client requests, agent records, daemon polling, deployment
  reasoning, and smoke-test acceptance free of `team_id`, `teamId`, OpenAPI
  task-context paths, Open API keys, and `X-Workspace-Api-Key`
- clarify that legacy Open API key task-context environment variables remain a
  compatibility adapter outside the generated AI Agent assignment SSOT

This slice does not edit endpoint shapes, generated OpenAPI, DynamoDB schema,
Terraform topology, task-context HTTP implementation, or `riido-client`.

### A-22 — assignment ready read model and generated-client delivery automation

This slice closes two frontend-handoff gaps found while testing the development
AI Agent flow.

This slice does:

- map daemon assignment `ready` events to the task-thread client read model as
  `comment_kind=assignment_started`, `assignment_state=running`, and
  `work_status=running`
- keep busy-agent queue presentation reserved for actual queued/busy assignment
  cases instead of reusing it for daemon/runtime handoff
- add `tools/generatedclienthandoff` so generated README, history, manifest,
  barrels, and PR body are produced from the same OpenAPI/DSL/IR inputs as the
  React Query client
- add `.github/workflows/generated-client-delivery.yml` so control-plane can
  package generated artifacts and open or update a draft `riido-client` PR
  without auto-merging it from API release tags, explicit manual dispatches, or
  path-filtered `main` pushes that change the AI Agent client contract/generator
  boundary
- guard the client handoff so only
  `src/generated/react-query/riido-control-plane/**` can be changed in the
  target branch, and so no-diff runs stop before creating or refreshing a PR
- normalize generated files with `riido-client`'s pinned Prettier setup before
  regenerating manifest hashes and PR body for the final client branch
- preserve the previous `contractManifest.generated.ts` from `riido-client`
  `main` and include added/removed/changed/no-surface-diff generated operation
  summary in the generated client PR body
- run the target branch generated-path Prettier check and `pnpm run type-check`
  before opening or updating the client PR
- extend the AI Agent client CI workflow to test both `tools/reactquerygen` and
  `tools/generatedclienthandoff`

This slice does not edit handwritten `riido-client` application code, merge a
client PR, change endpoint payload shapes, introduce npm/Orval delivery, add
delivery secrets, or change daemon provider execution behavior.

### RIID-4899 — generated client delivery token CI gate clarification

This slice absorbs the 2026-06-03 generated-client-delivery failures that came
from the legacy delivery workflow before the GitHub App token boundary was
introduced. Those failed runs required a raw `RIIDO_CLIENT_DELIVERY_TOKEN` and
also synthesized `react-query-*` target branches. Both behaviors are outside
the current SSOT.

This slice does:

- keep `generated-client-delivery.yml` as a package-first manual workflow:
  `create_pr=false` builds and uploads the generated-client handoff artifact
  without requiring cross-repository write credentials
- keep `create_pr=true` as the only currently enabled `riido-client` PR handoff
  path; it must resolve a short-lived GitHub App installation token first, with
  `RIIDO_CLIENT_DELIVERY_TOKEN` accepted only as a temporary fallback
- keep the missing-credential failure scoped to the delivery job, before
  checkout of `teamswyg/riido-client`
- keep the target branch as the Riido task `branchName`; the workflow must not
  synthesize `react-query-*` branch names
- add a regression gate so the old raw-token-only error string and synthetic
  branch naming cannot silently return

This slice does not configure repository secrets, open or update a
`riido-client` PR, edit `teamswyg/riido-client`, or change AI Agent endpoint
shape.

### RIID-4902 — Workspace assigned-agent profile map execution

This slice executes the contracts-owned assigned-agent profile read model in the
control plane.

This slice does:

- mirror the updated AI Agent client contract fixture from `riido-contracts`
- expose `GET /v2/client/workspaces/{workspace_id}/ai-agent/tasks/assigned-agent-profiles`
- generate `riido.v2.aiAgent.tasks.assignedAgentProfiles` and searchable
  generated comments
- preserve onboarding fixture `tmp_color` values on fixture-created agents
- return active queued/running/stopping assignments as a component_id keyed
  `assigned_agent_profiles` map
- extend the local generator so typed OpenAPI `additionalProperties.$ref` becomes
  `Record<string, AssignedAgentProfile>` instead of `Record<string, unknown>`

This slice does not add a v1 compatibility route, change participant dropdown
selection, change task-thread history, change daemon polling, or edit
`teamswyg/riido-client`.

### RIID-4913 — multi-agent task expansion SSOT regression tests

Riido task:
`RIID-4913-CONTROL-PLANE-MULTI-AGENT-TASK-EXPANSION-SSOT-REGRESSION-TESTS`.

Scope:

- keep existing v1/v2 `tasks.assignment`, `tasks.stop`, and
  `tasks.threads.active_stream` behavior intact for the current demo/client
  compatibility surface
- add v2-only additive task assignment generated paths:
  `riido.v2.aiAgent.tasks.agentAssignments.create`,
  `riido.v2.aiAgent.tasks.agentAssignments.delete`, and
  `riido.v2.aiAgent.tasks.agentAssignments.stop`
- add `riido.v2.aiAgent.tasks.threadStreamSubscription` so clients can receive
  one shared SSE stream handoff plus explicit active thread filters
- keep the daemon/actor invariant that an agent polls its own queue and can
  actively execute at most one assignment at a time
- allow one task to have multiple active agents only through the additive v2
  routes, not through the legacy/demo `tasks.assignment` route
- update the assignment-store GitHub Action to run both store actor and client
  assignment-domain regression tests when the relevant domain files change

Validation gates:

```bash
go test ./internal/riidoaiserver -run 'StoreActor|AssignmentStore|AIAgentClient.*Assignment|AIAgentClient.*ThreadStream' -count=1
go test ./internal/riidoaiserver -run 'AIAgentClient' -count=1
go test ./tools/reactquerygen -count=1
```

### RIID-4917 — assignment status SSE dedupe for provider heartbeats

This slice absorbs the 2026-06-05 development Codex smoke evidence where one
Go Hello World assignment completed successfully, but the SSE stream emitted
hundreds of repeated `agent_work_status_changed` events while the assignment
remained `(running, running, runtime_progress)`.

This slice does:

- keep `agent_thread_progress` as the live text/progress event for daemon
  `riido_log` batches
- fan out `agent_work_status_changed` only when daemon/provider assignment
  events change `(work_status, assignment_state, comment_kind)`
- strip internal `<riido_log>...<end>` transport blocks from client-visible
  terminal task-thread `message` while keeping rendered progress in `lines[]`
- enqueue a real assignment when a viewer sends a follow-up message to a
  completed AI Agent task thread; the assignment prompt includes a
  `## Follow-up Thread Message` section and daemon progress continues to update
  the same client-facing `thread_id`
- preserve terminal status events such as completed, failed, stopped, queued,
  and the first transition from assignment-started into runtime-progress
- add regression tests that duplicate running assignment events do not grow the
  client event stream, a later completed event still does without exposing raw
  progress transport in the cold thread message, and a completed-thread
  follow-up becomes pollable daemon work instead of a read-model-only state

This slice does not change endpoint shape, generated OpenAPI, frontend code,
assignment replacement behavior, or the v2 multi-agent
`thread-stream-subscription` contract. It does intentionally change daemon
polling behavior for completed-thread follow-up messages from no-op/read-model
only to a queued assignment.

### RIID-4917 — development seed device visibility boundary

This slice keeps deterministic AI Agent client seed devices for static
fixture/generated API tests, but prevents those seed devices from leaking into
external-authorized development workspaces. Real development workspaces now see
device/runtime rows from desktop enrollment plus daemon runtime sync only. This
preserves the existing static fixture behavior while treating
the development AI API host as a real persistent development environment rather
than a mixed mock surface.

### RIID-4917 — public progress stream renders structured payload fallback

This slice closes a live development finding where a runtime progress line could
reach the client event stream as raw JSON-shaped text such as
`{"code":1102,"args":{"count":1}}`. The shared progress-message SSOT already
requires fixed, translated, integer-coded progress messages to be rendered
before public SSE delivery, so the control-plane now treats a bare JSON progress
payload in `message` as a defensive fallback input and renders it into the same
Korean `message` string shape the frontend already consumes.

This slice does:

- parse bare JSON progress payloads during `AgentThreadProgressLine`
  normalization when `message_code` is absent
- accept primitive JSON arg values, including numeric `count`, and normalize
  them into `message_args`
- keep daemon-provided `message_code` / `message_key` / `message_args`
  precedence when they are already present
- preserve existing public client shape and generated endpoint paths

This slice does not add new SSE event variants, change frontend behavior,
change the progress catalog, or require generated-client updates.

### RIID-4917 — active provider heartbeat does not own task-thread copy

Live development verification with daemon `v0.0.10` showed that Codex could
complete a follow-up Rust Hello World assignment and emit fixed rendered
progress `lines[]`, while an active provider heartbeat briefly overwrote the
cold task-thread representative `message` with provider-internal text such as
`codex unknown notification: item/started` or a JSON warning log. That conflicts
with the AI Agent client API SSOT: active non-`riido_log` daemon/provider events
are lifecycle heartbeats, not client copy.

This slice does:

- preserve the existing task-thread representative `message` for active
  non-`riido_log` assignment events
- keep rendered progress `lines[]` as the owner of visible in-progress copy
- keep terminal assignment messages visible, with `<riido_log>...<end>` blocks
  stripped as before
- add a store regression test proving provider heartbeat/log events do not
  replace a rendered progress message

This slice does not change generated endpoint paths, add frontend filtering
requirements, or alter terminal completed/failed message handling.

## Validation Gates

Required before a control-plane migration PR is mergeable:

```bash
go test ./...
go list -m all
go test ./internal/riidoaiserver -run 'ReviewAccount|HTTPReviewAccount|AgentCatalogStore' -count=1
go test ./cmd/riido_ai_server -run 'ReviewAccount|ConfigFromEnv|AuthorizerFromEnv' -count=1
go test ./internal/riidoaiserver -run 'CloudWatch|Metrics|EventAppend' -count=1
go test ./internal/riidoaiserver -run 'DynamoDB|EventBridge' -count=1
go test ./awsadapters -count=1
go test ./cmd/riido_ai_server -run 'MetricsLog|ConfigFromEnv|EnvOptionalDuration' -count=1
go test ./tools/containercontract -count=1
go test ./internal/riidoaiserver -run 'WebFrontendCORS' -count=1
go test ./internal/riidoaiserver -run 'AIAgentClient' -count=1
go test ./cmd/riido_ai_server -run 'AIAgentClient|WebAllowedOrigins|ConfigFromEnv' -count=1
go test ./tools/reactquerygen -count=1
go test ./tools/generatedclienthandoff -count=1
go run ./tools/containercontract -contract packaging/containers/riido_ai_server_container.riido.json -out -
docker build -f packaging/containers/riido_ai_server.Dockerfile -t riido-control-plane:local .
```

Architecture-doc migration PRs must also pass:

```bash
test -f docs/20-domain/context-map.md
test -f docs/30-architecture/module-decomposition.md
test -f docs/30-architecture/api-client-delivery.md
test -f docs/30-architecture/config-reference.md
test -f docs/30-architecture/integration-matrix.md
test -f docs/30-architecture/runtime-deployment-boundary.md
test -f docs/50-roadmap/open-questions.md
go test ./...
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
contract. After RIID-4807 it also owns tag-triggered testnet runtime artifact CD:
immutable image push, ECS task-definition revision registration, ECS service
stability wait, and live AI Agent development smoke.

`riido-infra` still owns AWS topology and consumes immutable artifacts by:

- Git tag
- module version
- container image digest

The control-plane repository must not need AWS credentials for normal pull
request verification.

## Contract Boundary

Use `riido-contracts` for shared API/fixture facts only after duplication exists
or after daemon and control-plane both consume the same schema. Until then,
keep implementation-local types inside `riido-control-plane`.
