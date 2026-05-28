# SaaS Control Plane SSOT

> Riido tasks: RIID-4668 `[Control Plane] assignment contract/type migration`,
> RIID-4688 `[Control Plane] riido-contracts v0.3.0 assignment import migration`,
> RIID-4691 `[Control Plane] review account seed runtime wiring migration`,
> RIID-4692 `[Control Plane] CloudWatch EMF metrics publisher migration`,
> RIID-4704 `[Control Plane] DynamoDB/EventBridge adapter migration`,
> RIID-4706 `[Control Plane] AWS adapter public facade migration`

This document is the public SSOT for the SaaS control-plane assignment contract
surface that can be verified without AWS credentials.

## Responsibility

`riido-control-plane` owns the server-side behavior for assigning component
tasks to SaaS agent identities, daemon polling, daemon heartbeat handling,
agent event sync, task event storage/streaming, health responses, and metrics
snapshots. It also owns stdout CloudWatch Embedded Metric Format publication
from those metrics snapshots and stdlib-only DynamoDB/EventBridge adapter
behavior that can be verified with local black-box HTTP tests. Shared
assignment polling DTOs and state vocabulary are imported from
`github.com/teamswyg/riido-contracts/assignment`.
It also owns the public-safe store review seed artifact and runtime
provisioning path that can be verified without raw review tokens, provider
execution grants, AWS credentials, or Terraform state.

This document does not own customer-PC provider process execution, local daemon
configuration, Terraform, AWS account/resource configuration, private
deployment evidence, live AWS evidence collection, or production secret values.

The split-repo context map is
[`context-map.md`](context-map.md). Package decomposition, runtime config,
integration gates, and release hand-off are owned by
[`../30-architecture/module-decomposition.md`](../30-architecture/module-decomposition.md),
[`../30-architecture/config-reference.md`](../30-architecture/config-reference.md),
[`../30-architecture/integration-matrix.md`](../30-architecture/integration-matrix.md),
and
[`../30-architecture/runtime-deployment-boundary.md`](../30-architecture/runtime-deployment-boundary.md).

## Executable Contract

The executable assignment polling contract is owned by
`github.com/teamswyg/riido-contracts/assignment` and documented in
`riido-contracts/docs/20-domain/assignment-polling.md`.

That contract owns:

- `service_schema_version`
- assignment state values
- assignment terminal classification
- assignment agent-active classification
- legal assignment transitions
- daemon poll action values
- task event type values
- assignment/poll/heartbeat/event/task-event DTO JSON field names
- agent runtime binding DTO JSON field names

The local Go surface in `internal/riidoaiserver/assignment_contract_gen.go` and
`assignment_api.go` is an alias/import layer over that shared package so that
existing control-plane store, HTTP, SSE, and metrics code can preserve its
internal API while the cross-repository contract lives in `riido-contracts`.
Markdown must link to the shared executable contract instead of redefining the
transition matrix.

## Public DTO Surface

The shared assignment DTO surface imported from `riido-contracts/assignment` is:

- `AssignRequest`
- `Assignment`
- `PollRequest`
- `PollResponse`
- `AgentHeartbeatRequest`
- `AgentHeartbeatResponse`
- `AgentEventRequest`
- `AgentEventResponse`
- `TaskEvent`

Runtime progress intended for the client task thread is ingested as bounded
daemon batches on `POST /v1/agents/{agent_id}/thread-progress`. The endpoint
stores each accepted line as an assignment `riido_log` task event and, when the
AI Agent client event store is configured, fans out the same batch as
`agent_thread_progress` on the client SSE surface. The batch may carry
`thread_id`; when omitted, the server derives `thread_id` from `assignment_id`
so historical task-thread collections and live stream events target the same
thread deterministically.

The control-plane-local DTO surface is:

- `Health`
- `MetricsSnapshot`

These types are API/adapter contracts only. They do not own HTTP routing, SSE
fan-out, outbox, snapshots, or AWS adapters. `Health` and `MetricsSnapshot`
remain local because they are control-plane adapter/read-model contracts rather
than daemon polling contracts.

## Store Actor Boundary

`internal/riidoaiserver.Store` is the public in-memory assignment actor. It owns
the stdlib-only runtime behavior that can be verified without AWS credentials:

- `AssignmentStore` command serialization through one actor goroutine
- assignment creation and reassignment cancellation handoff
- daemon poll actions (`none`, `start`, `active`, `cancel`)
- heartbeat-based active assignment timestamp refresh
- agent event transition validation and task event append
- metrics read-model counters for tasks, assignments, poll actions, and events
- in-memory provider status sync/read state used by store-safe routing
- configurable snapshot and task-event outbox port calls after assignment
  mutations
- configurable assignment operation journal save/replay/claim calls after
  assignment mutations
- durable active-assignment lease reads and heartbeat refreshes when the
  configured operation store implements the lease/projection ports

This actor does not own HTTP assignment routes, SSE fan-out,
DynamoDB/EventBridge adapter payload construction, Terraform, AWS credentials,
or deployment evidence.

## Store Snapshot And File Outbox Boundary

The public store snapshot and file outbox boundary owns the stdlib-only
persistence adapters that can be verified without AWS credentials:

- `SnapshotStore`
- `StoreSnapshot`
- `StoreSnapshotTask`
- `FileStoreSnapshot`
- `OpenStoreWithConfig`
- `EventSink`
- `OutboxRecord`
- `FileOutbox`

`StoreSnapshot` is a point-in-time assignment store snapshot. It preserves
tasks, assignments, agent-assignment indexes, task event history, and the next
assignment/event sequence counters. `FileStoreSnapshot` writes that snapshot as
strict JSON using atomic replace. Loading rejects unknown fields, unsupported
schema versions, trailing JSON, blank task ids, blank assignment ids, and agent
assignment references that do not exist in the snapshot assignment set.

`FileOutbox` appends task events as JSON Lines `OutboxRecord` values. The store
actor calls the outbox after task events are appended for assignment queue,
lease, and agent-event mutations. Outbox append errors do not fail the
assignment mutation; they increment the public `outbox_errors_total` metrics
counter and still record event-append latency counters.

This boundary does not own `DynamoDBStoreSnapshot`, `DynamoDBOutbox`, DynamoDB
Streams relays, EventBridge publishers, assignment operation durable save/claim
adapter implementation, Terraform, AWS credentials, Docker image contracts,
review account seed data, or deployment evidence.

## Assignment HTTP Adapter Boundary

The public HTTP assignment adapter routes request/response JSON into the
`AssignmentStore` port. It owns these stdlib-only routes:

- `POST /v1/component-tasks/{task_id}/assignment`
- `POST /v1/agents/{agent_id}/poll`
- `POST /v1/agents/{agent_id}/heartbeat`
- `POST /v1/agents/{agent_id}/events`

Every route must use `RequestAuthorizer` before it reaches the store. Request
bodies use the strict JSON decoder, so unknown fields are rejected and private
provider-path/token material cannot be accepted by accident.

This adapter does not own the task event SSE route, `/metrics`, health/ready
routes, `cmd/riido_ai_server` environment parsing, snapshot stores, file outbox
adapters, durable operation save/claim wiring, DynamoDB, EventBridge,
Terraform, AWS credentials, or deployment evidence.

## Task Event SSE Adapter Boundary

The public task event SSE adapter streams `TaskEvent` records from the
`AssignmentStore` subscription port. It owns this stdlib-only route:

- `GET /v1/component-tasks/{task_id}/events`

The route must use `RequestAuthorizer` with `component_task_events` /
`events:read` scope before subscribing. On connection it replays existing task
event history as SSE messages. With `replay=1`, the adapter flushes history and
returns without holding the stream open. Without `replay=1`, it keeps the stream
open and forwards later task events until the request context is cancelled.

The SSE adapter does not own `/metrics`, health/ready routes,
`cmd/riido_ai_server` environment parsing, snapshot stores, file outbox
adapters, durable operation save/claim wiring, DynamoDB, EventBridge,
Terraform, AWS credentials, daemon/GUI SSE consumers, or deployment evidence.

## Metrics HTTP Adapter Boundary

The public metrics HTTP adapter exposes the `MetricsSnapshot` read model from
the `AssignmentStore` port. It owns this stdlib-only route:

- `GET /metrics`

The route must use `RequestAuthorizer` with `metrics` / `read` scope before
reading the store. The response is the `MetricsSnapshot` DTO, including
`riido-ai-server-metrics.v1` schema version and the in-memory assignment,
poll, event, subscriber, outbox-error, and event-latency counters that are
available without AWS credentials.

The metrics adapter does not own health/ready routes, Prometheus conversion,
production tuning calibration, `cmd/riido_ai_server` environment parsing,
snapshot stores, file outbox adapters, durable operation save/claim wiring,
DynamoDB, EventBridge, Terraform, AWS credentials, dashboards, daemon
consumers, or deployment evidence.

## CloudWatch EMF Metrics Boundary

The public CloudWatch EMF metrics boundary emits the same `MetricsSnapshot`
read model as stdout JSON Lines in CloudWatch Embedded Metric Format. It owns:

- `CloudWatchEMFConfig`
- `PublishCloudWatchEMF`
- `RunCloudWatchEMFPublisher`
- `WriteCloudWatchEMF`
- `RIIDO_AI_SERVER_METRICS_LOG_INTERVAL_SECONDS`

When `RIIDO_AI_SERVER_METRICS_LOG_INTERVAL_SECONDS` is a positive integer,
`cmd/riido_ai_server` starts the publisher, writes one metrics record
immediately, and then writes at the configured interval until shutdown. The EMF
record includes assignment, poll, agent event, task event, SSE subscriber,
outbox error, and event-append-latency counters.

This boundary owns only stdout EMF serialization and runtime scheduling. It
does not own AWS SDK calls, CloudWatch PutMetricData, credentials, log group or
dashboard creation, production tuning samples, Prometheus conversion,
DynamoDB, EventBridge, Terraform, or deployment evidence.

## DynamoDB/EventBridge Adapter Boundary

The public DynamoDB/EventBridge adapter boundary owns stdlib-only AWS request
construction, SigV4 signing, serialization, and local fake-endpoint behavior
for the control-plane durable adapters:

- `AWSCredentials`, `AWSCredentialsProvider`, `StaticAWSCredentialsProvider`,
  and `ECSContainerCredentialsProvider`
- `DynamoDBOutbox`
- `DynamoDBStoreSnapshot`
- `DynamoDBAssignmentOperationStore`
- DynamoDB table stream discovery
- DynamoDB Streams relay and checkpoint handling
- EventBridge stream relay publishing

These adapters are production adapter code, but their public verification must
use only fake endpoints, fake credentials, `httptest`, and deterministic local
black-box scenarios. They may construct AWS JSON payloads and sign HTTP
requests; they must not require live AWS credentials in pull-request CI.

This boundary does not own AWS account ids, Terraform, IAM/VPC/ECS/EventBridge
rule resources, Route53/ACM/WAF resources, tfvars, Terraform backend/state,
live DynamoDB/EventBridge smoke evidence, stream-relay evidence artifacts,
runtime secret values, ECR image push evidence, dashboards, or private
deployment evidence. It also does not add or require the external Go AWS SDK.

## AWS Adapter Public Facade Boundary

The `awsadapters` package is a narrow public Go module facade over the
`internal/riidoaiserver` DynamoDB/EventBridge adapter implementation. It owns
only type aliases and constructor/function re-exports needed by private
operational tooling, such as `riido-infra` stream-relay evidence collection.

The facade may expose:

- AWS credential provider DTOs and constructors
- DynamoDB outbox, snapshot, operation-store, table-stream, stream-relay, and
  checkpoint adapter DTOs and constructors
- EventBridge stream relay publisher DTOs and constructor
- assignment/task-event DTO aliases needed to build smoke events

The facade must not fork or redefine adapter behavior. Behavioral code stays in
`internal/riidoaiserver`; the facade compiles against it so external repos can
consume the same production adapter surface through the public module path.

This boundary does not own live AWS evidence collection, Terraform, AWS
credentials, release evidence artifacts, runtime secret values, or deployment
automation. Those remain private infra responsibilities.

## Health/Ready And Runtime Command Boundary

The public health/ready adapter exposes liveness and readiness responses that do
not require request authorization:

- `GET /healthz`
- `GET /readyz`

Both routes return the `Health` DTO with the current control-plane schema
version. Non-`GET` methods must fail with `405`.

`cmd/riido_ai_server` is the minimal stdlib-only runtime entrypoint for this
public repository. It owns only these environment variables:

- `RIIDO_AI_SERVER_ADDR`
- `RIIDO_AI_SERVER_SHUTDOWN_TIMEOUT_SECONDS`
- `RIIDO_AI_SERVER_AGENT_BINDINGS_JSON`
- `RIIDO_AI_SERVER_AUTHZ_TOKENS_JSON`
- `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_URL`
- `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_AUDIENCE`
- `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_TIMEOUT_SECONDS`
- `RIIDO_AI_SERVER_REVIEW_ACCOUNT_TOKEN_SHA256`
- `RIIDO_AI_SERVER_METRICS_LOG_INTERVAL_SECONDS`
- `RIIDO_AI_SERVER_WEB_ALLOWED_ORIGINS`
- `RIIDO_AI_SERVER_AI_AGENT_CLIENT_MOCK`

The agent binding and static-token JSON values use strict decoding, so unknown
fields and trailing JSON are rejected. Static-token authorization may be
combined with the external HTTP authorizer through the existing fallback
authorizer rule: only unauthenticated results fall through to the next
authorizer, while forbidden results stop evaluation.

`RIIDO_AI_SERVER_REVIEW_ACCOUNT_TOKEN_SHA256` enables only the public-safe
review account provisioning path. The environment value is a SHA-256 hash of an
externally supplied review token; the raw token remains outside this
repository.

`RIIDO_AI_SERVER_AI_AGENT_CLIENT_MOCK` enables only the temporary AI Agent
client mock API described in [`ai-agent-client-api.md`](ai-agent-client-api.md).

This boundary does not own legacy broad bearer-token compatibility,
snapshot/outbox stores, durable operation save/claim wiring, DynamoDB,
EventBridge, Terraform, AWS credentials, CloudWatch API wiring, Prometheus
adapters, Docker image contracts, raw review token values, production secrets,
or deployment evidence.

## Review Account Seed Boundary

The review account seed boundary owns the public-safe App Store/MS Store review
and demo control-plane bootstrap data:

- `riido-review-account-seed.v1`
- a non-admin `store-reviewer` principal
- static-token credential provisioning from
  `RIIDO_AI_SERVER_REVIEW_ACCOUNT_TOKEN_SHA256`
- seeded agent catalog records that demonstrate owner/private, owner/public,
  other-user/public, and other-user/private RBAC visibility
- a synthetic `store-review-agent` provider-status snapshot for the
  `mac-app-store` distribution channel
- non-routable provider statuses only: `login-required` or `unsupported`

The seed artifact must not contain raw tokens, passwords, provider executable
paths, workspace root paths, API keys, AWS credentials, or provider execution
grants. The review principal may read metrics, read the agent catalog, read the
synthetic provider status, assign component tasks, and read component-task
events. It must not poll, heartbeat, write agent events, or write provider
status as a daemon.

This boundary does not own production IdP rollout, raw review token issuance,
real provider execution, daemon/provider bundling, DynamoDB/EventBridge
adapters, Terraform, AWS credentials, production secrets, or deployment
evidence.

## Container Image Contract Boundary

The public container image contract owns the buildable `riido_ai_server`
artifact shape that can be verified without AWS credentials:

- `packaging/containers/riido_ai_server.Dockerfile`
- `packaging/containers/riido_ai_server_container.riido.json`
- `tools/containercontract`

The executable contract requires a two-stage Go build, `CGO_ENABLED=0`, the
`./cmd/riido_ai_server` package, a `scratch` final image, copied CA
certificates, `EXPOSE 8080`, `RIIDO_AI_SERVER_ADDR=:8080`, non-root
`65532:65532`, and `ENTRYPOINT ["/riido_ai_server"]`.

`tools/containercontract` is the stdlib-only verifier for
`riido-container-image-contract.v1`. It emits
`riido-container-image-contract-check.v1` evidence and may optionally validate a
private Fargate task-definition IR when another repository supplies that path.

This boundary does not own ECR repositories, image push permissions, immutable
image digest evidence, Terraform/Fargate task definitions, AWS credentials,
runtime secret values, production environment variables, or deployment
evidence. Those remain `riido-infra` responsibilities.

## Durable Operation Boundary

The assignment operation journal and claim-port contract is owned by
[`assignment-operation-journal.md`](assignment-operation-journal.md).

That boundary owns operation records, assignment projection records,
active-assignment lease records, and durable claim/read ports. It does not own
the store actor, HTTP routes, SSE, DynamoDB payload construction, Terraform, or
deployment evidence.

It also owns the pure operation replay reducer that reconstructs internal
assignment projection state from operation records before a later store actor
slice consumes it. RIID-4681 is that store actor runtime-consumption slice:
the store actor can now save operation records, replay them when no snapshot is
available, claim the next assignment through an `AssignmentClaimer`, and consult
durable active-assignment lease/projection ports during poll and heartbeat.

## Provider Status Boundary

The provider status sync/read contract is owned by
[`provider-status.md`](provider-status.md).

That boundary owns provider status DTOs, normalization, read/write ports, and
the `GET`/`POST /v1/agents/{agent_id}/provider-status` HTTP adapter. It also
owns the pure store-safe routing guard that evaluates synced provider routing
status before a later assignment integration calls it. It does not own provider
executable detection, customer-PC provider process execution, durable store
actors, DynamoDB payloads, Terraform, or deployment evidence.

## Migration State

RIID-4668 moved the executable assignment contract and DTO surface from the
former private `riido_daemon/internal/riidoaiserver` package into this public
repository.

RIID-4688 moves the shared assignment polling contract SSOT into
`riido-contracts v0.3.0` and changes this repository to consume that tagged
contract through aliases/imports. Control-plane health/metrics DTOs and all
store/HTTP/SSE behavior remain local.

RIID-4692 moves the stdout CloudWatch EMF metrics publisher into this public
repository and wires optional command startup through
`RIIDO_AI_SERVER_METRICS_LOG_INTERVAL_SECONDS`, while leaving AWS resources,
dashboards, Terraform, and deployment evidence in private infra.

RIID-4704 moves stdlib-only DynamoDB/EventBridge adapter behavior into this
public repository, including DynamoDB outbox/snapshot/operation stores,
DynamoDB Streams relay/checkpoint handling, and EventBridge publisher request
construction. Live AWS configuration, Terraform, stream-relay evidence
collection, credentials, and deployment evidence remain private infra
responsibilities.

RIID-4706 adds the public `awsadapters` facade so private infra tooling can
consume RIID-4704 adapter behavior through the `riido-control-plane` Go module
without importing an `internal` package or duplicating adapter behavior.

RIID-4712 adds the public architecture SSOT set for the split-repo control-plane
boundary: context map, module decomposition, config reference, integration
matrix, runtime/deployment hand-off, open questions, and a focused public docs
workflow.

RIID-4669 moves the operation journal port and record surface into this public
repository.

RIID-4673 moves the assignment operation replay reducer into this public
repository.

RIID-4671 moves the provider status DTO/port/HTTP contract into this public
repository, using `riido-contracts v0.2.0+` for shared provider/distribution
vocabulary.

RIID-4672 moves the pure store-safe routing guard into this public repository.

RIID-4674 moves the stdlib-only in-memory assignment store actor into this
public repository.

RIID-4675 moves the assignment HTTP adapter into this public repository.

RIID-4677 moves the task event SSE adapter into this public repository.

RIID-4678 moves the metrics HTTP adapter into this public repository.

RIID-4679 moves health/ready routes and the minimal `cmd/riido_ai_server`
environment/runtime entrypoint into this public repository.

RIID-4680 moves stdlib-only store snapshot and file outbox adapters into this
public repository.

RIID-4681 wires durable assignment operation journal ports into the public
store actor runtime without moving DynamoDB adapters or Terraform.

RIID-4682 moves the public Docker image contract, Dockerfile, container contract
verifier, and focused CI into this repository. ECR push, Terraform/Fargate task
definition IR, image digest deployment evidence, AWS credentials, and runtime
secret values remain private infra responsibilities.

RIID-4691 moves the public-safe review account seed artifact, provisioning
domain, in-memory agent catalog store actor commands, command env wiring, and
black-box review account HTTP scenarios into this repository. Raw review token
values, production IdP rollout, AWS adapters, Terraform, image digest evidence,
and production deployment evidence remain separate migration units.

## Open Questions

Unresolved control-plane decisions are owned by
[`../50-roadmap/open-questions.md`](../50-roadmap/open-questions.md).
