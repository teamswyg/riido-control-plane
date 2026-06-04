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

`AssignRequest.agent_instruction` and `Assignment.agent_instruction` are the
assignment-created snapshots of the saved agent instruction. The store trims the
value, enforces the 1000-character contract limit, persists it with the
assignment, and returns it in poll/heartbeat/event responses. Provider-specific
placement belongs to `riido-daemon`, not this repository.

`AssignRequest.allow_experimental_runtime` and
`Assignment.allow_experimental_runtime` are assignment-created snapshots derived
from the selected runtime's `RuntimeRecord.requires_experimental_opt_in` fact.
The control plane derives this at assignment creation time; daemon execution
must consume the snapshot instead of inferring opt-in from provider names or
environment variables.

Runtime progress intended for the client task thread is ingested as bounded
daemon batches on `POST /v1/agents/{agent_id}/thread-progress`. The endpoint
stores each accepted line as an assignment `riido_log` task event and, when the
AI Agent client event store is configured, fans out the same batch as
`agent_thread_progress` on the client SSE surface.

The daemon's standard assignment event path,
`POST /v1/agents/{agent_id}/events`, is also part of the client task-thread
projection. `riido_log` events append to the generated client's thread lines.
The daemon's parsed progress batch path, `POST /v1/agents/{agent_id}/thread-progress`,
may omit `thread_id`; in that case the batch is reconciled to the active
client-facing thread for the same `(task_id, agent_id)` rather than materializing
a second thread from the assignment run id.
`completed`, `failed`, and `cancelled` assignment states close the thread's
`active_stream` read-model state so a completed historical thread remains a
cold collection row instead of a live stream candidate. This projection is
derived from the assignment/device principal boundary; it does not use team id
or an Open API workspace key.

## Multi-Agent Task Expansion Boundary

> Riido task: RIID-4913 `CONTROL PLANE MULTI AGENT TASK EXPANSION SSOT REGRESSION TESTS`

Existing participant assignment routes remain compatibility/demo routes. The
store method behind those routes keeps the previous replacement handoff: when a
different agent is assigned to the same task, the current task assignment is
marked cancelling and the new assignment is blocked until the old assignment is
terminal.

New multi-agent task work is additive and must use the v2
`agent-assignments` routes. The additive store path preserves the same
validation gates as compatibility assignment: task id, agent id, runtime
provider, prompt, agent/runtime binding, store-safe provider routing,
experimental runtime snapshot, and assignment-created agent instruction
snapshot. It differs only in task-level replacement behavior: another agent's
active assignment on the same task is not cancelled.

The concurrency invariant is therefore:

- an agent is still globally polled through its own assignment queue and can
  actively execute at most one assignment at a time
- one task may have multiple active agents only through the additive v2
  `agent-assignments` routes
- legacy `tasks.threads.active_stream` remains a singular handoff for UI
  compatibility and targets the latest visible active thread
- multi-agent clients must read
  `riido.v2.aiAgent.tasks.threadStreamSubscription`, then apply
  `active_thread_filters[]` to the shared client SSE stream
- stop/removal for multi-agent tasks must carry the target `agent_id` in the
  path through `riido.v2.aiAgent.tasks.agentAssignments.stop` or
  `riido.v2.aiAgent.tasks.agentAssignments.delete`

## Assignment Prompt Composer Boundary

> Riido task: RIID-4799 `contracts server assignment prompt composer ssot`

`riido-control-plane` owns the provider-neutral assignment prompt composer.
When an AI Agent is assigned to a task from the client-facing participant flow,
the client sends only the selected `agent_id`; it does not send task body,
branch, repository, or agent instruction text.

The prompt composer consumes a read-only task context snapshot whose source is
the existing Riido API server endpoint added by RIID-4798:

- task/component id, type, title, key number, and branch name
- task document content, projected as markdown by the existing API server
- project, milestone, and parent-task hierarchy labels
- connected GitHub repository candidates, with pull-request-connected
  repositories preferred over workspace-connected fallback repositories

The composed `AssignRequest.prompt` is an assignment-time immutable snapshot.
Later task document edits, repository relation edits, or agent setting edits do
not rewrite queued or running assignments. The saved agent `instruction` is
snapshotted separately into `AssignRequest.agent_instruction`.

This boundary owns only the deterministic, provider-neutral prompt scaffold and
repository selection rule. It does not own:

- frontend participant dropdown shape or `agent_id` request body
- existing Riido API server authorization, database queries, or document HTML to
  markdown conversion
- daemon/provider-specific placement of `Assignment.prompt` versus
  `Assignment.agent_instruction`
- production secret names for server-to-server calls

> Riido task: RIID-4800 `server task context http client assignment prompt wiring`

The production wiring path reads task context from the existing API server as a
server-to-server call. For browser/desktop-webview generated assignment,
`team_id`, `teamId`, OpenAPI task-context paths, and Open API key transport such
as `X-Workspace-Api-Key` are outside the problem entirely. They are not
generated request fields, not agent fields, not daemon polling inputs, not
deployment prerequisites, and not smoke-test acceptance criteria. For generated
`tasks.assign`, control-plane uses the request's already-authorized user token,
resolves the task's team location from `task_id` through the existing private
component workspace lookup, reads the private component document, composes
`AssignRequest.prompt`, and then calls the assignment store. The selected agent
is not made team-aware; any team value observed during lookup is a transient
API-server location result and must not be persisted into the agent or exposed
to generated clients. If lookup or composition fails, the HTTP handler fails
before a daemon can lease provider work.
The task-location lookup result is not an identity bridge. Device enrollment,
daemon polling, generated assignment acceptance, and staging E2E verification
must still use UserPrincipal/DevicePrincipal plus workspace-scoped agent facts,
never `team_id`, `teamId`, OpenAPI task-context paths, Open API keys, or
`X-Workspace-Api-Key`.

The legacy Open API key task-context reader is a separate compatibility adapter
for automation surfaces outside the generated AI Agent assignment flow. It is
not the SSOT for generated client behavior and must not be used to reason about,
or validate, the daemon-facing assignment path.

Development fixture responses may still use deterministic synthetic thread
responses, but those responses are not evidence that a daemon assignment prompt
exists.

The generated AI Agent client assignment path is not allowed to stop at the
client read model. `POST
/v1/client/ai-agent/tasks/{task_id}/assignment` and its v2 workspace-scoped
alias first resolve the visible agent, load the agent runtime binding from the
AI Agent client registry, compose the provider-neutral task prompt through the
same task-context reader, and enqueue the resulting `AssignRequest` into the
assignment store. Only after that succeeds may the handler project the
client-facing task-thread row. A successful generated assignment response must
therefore be observable by the owning daemon as a later `poll` `start` or
`active` action.

## Daemon Runtime Snapshot Boundary

`POST /v1/daemon/runtime-snapshot` is a daemon-owned runtime batch upsert into
the SaaS read model. It is not a full replacement of every runtime previously
seen on the device. The customer-PC daemon may detect and report provider
runtime actors independently, so Codex, Claude Code, OpenClaw, and Cursor
snapshot batches can arrive at different times for the same `device_id` and
`daemon_id`.

The control plane merges incoming runtime records by `runtime_id`, updates the
matching runtime's availability, detection state, model catalog, and assigned
agent flag, and preserves other runtimes on the same device. This is required
for daemon `agent-bindings` polling: a later Cursor snapshot must not erase an
already assigned Codex runtime before the Codex actor can poll and claim its
work. Explicit daemon stop/offline handling remains the path that marks the
device's known runtimes offline together.

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

`DynamoDBAssignmentOperationStore` must page replay and queue `Query` calls with
a small explicit `Limit` instead of asking DynamoDB for the largest possible
page. Assignment operation records can contain task context and event payloads;
bounded query pages keep the stdlib HTTP JSON reader from decoding truncated
large responses during ECS startup replay.

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
- `RIIDO_AI_SERVER_AUTHZ_TOKENS_JSON`
- `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_URL`
- `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_AUDIENCE`
- `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_API_KEY`
- `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_TIMEOUT_SECONDS`
- `RIIDO_AI_SERVER_REVIEW_ACCOUNT_TOKEN_SHA256`
- `RIIDO_AI_SERVER_METRICS_LOG_INTERVAL_SECONDS`
- `RIIDO_AI_SERVER_WEB_ALLOWED_ORIGINS`
- `RIIDO_AI_SERVER_AI_AGENT_CLIENT_DEVELOPMENT`
- `RIIDO_AI_SERVER_AI_AGENT_CLIENT_DYNAMODB_TABLE`
- `RIIDO_AI_SERVER_AWS_REGION`
- `RIIDO_AI_SERVER_DYNAMODB_ENDPOINT`
- `RIIDO_AI_SERVER_TASK_CONTEXT_BASE_URL`
- `RIIDO_AI_SERVER_TASK_CONTEXT_WORKSPACE_ID`
- `RIIDO_AI_SERVER_TASK_CONTEXT_TEAM_ID`
- `RIIDO_AI_SERVER_TASK_CONTEXT_WORKSPACE_API_KEY`
- `RIIDO_AI_SERVER_TASK_CONTEXT_TIMEOUT_SECONDS`
- `AWS_CONTAINER_CREDENTIALS_FULL_URI`
- `AWS_CONTAINER_CREDENTIALS_RELATIVE_URI`
- `AWS_CONTAINER_AUTHORIZATION_TOKEN`

Static-token JSON values use strict decoding, so unknown fields and trailing JSON
are rejected. Static-token authorization may be combined with the external HTTP
authorizer through the existing fallback
authorizer rule: only unauthenticated results fall through to the next
authorizer, while forbidden results stop evaluation.

The external authorizer API key is a server-to-server secret for the
control-plane to authorizer hop. It is sent as
`X-Riido-Control-Plane-Authorizer-Key`, never as a generated frontend token, and
does not replace the request token supplied by the web or desktop webview
client.

`RIIDO_AI_SERVER_REVIEW_ACCOUNT_TOKEN_SHA256` enables only the public-safe
review account provisioning path. The environment value is a SHA-256 hash of an
externally supplied review token; the raw token remains outside this
repository.

`RIIDO_AI_SERVER_AI_AGENT_CLIENT_DEVELOPMENT` enables the development AI Agent
client API described in [`ai-agent-client-api.md`](ai-agent-client-api.md). The
development API uses a DynamoDB snapshot item for AI Agent client read/write
state and a DynamoDB assignment operation journal/queue for generated
assignment-to-daemon polling. Client thread projection and daemon work leasing
must share this durable boundary so an accepted generated assignment can be
observed by a later daemon `poll` even when the HTTP requests land on different
ECS tasks or after a deployment restart.

The generated AI Agent assignment path configures only the production
server-to-server base URL for the existing Riido API server. `team_id`,
`teamId`, OpenAPI task-context paths, and Open API key transport such as
`X-Workspace-Api-Key` are not part of that problem; they must stay out of
generated client requests, agent records, daemon polling, smoke criteria, and
deployment reasoning for this flow. The legacy Open API key task-context
environment variables remain a compatibility adapter outside the generated AI
Agent assignment SSOT; when both paths are present, generated assignment
behavior is judged only by the private user-token task context and
DevicePrincipal assignment boundary.

This boundary does not own legacy broad bearer-token compatibility, assignment
snapshot/outbox stores, durable operation save/claim wiring, EventBridge,
Terraform, AWS credentials, CloudWatch API wiring, Prometheus
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

This boundary does not own ECR repositories, Terraform-created ECS/Fargate
topology, AWS account topology, raw runtime secret values, production
environment values, DNS/ACM/WAF resources, or Terraform state. Those remain
`riido-infra` responsibilities.

For the AI Agent testnet, `riido-control-plane` owns the tag-triggered runtime
artifact CD workflow: build the checked-in image contract, push an immutable ECR
image tag, resolve it to a digest, register a new ECS task-definition revision,
wait for ECS service stability, and run live `healthz`/`readyz`/v2 bootstrap
smoke. The workflow uses GitHub OIDC and named secrets/variables only; no AWS
account values, raw tokens, Terraform state, live URLs, task-definition ARNs,
image digests, workflow run URLs, or live evidence payloads are committed to
this public repository. The live smoke target comes only from the configured
GitHub environment variable, not from workflow_dispatch URL inputs. Live values
that are needed between deploy steps stay in `$RUNNER_TEMP` files inside the
same job and are re-masked before use rather than published as GitHub step
outputs.

If production switches to CodeDeploy blue/green, the same ownership rule holds:
`riido-control-plane` owns the runtime artifact CD workflow and post-shift
smoke, while `riido-infra` owns the CodeDeploy topology, IAM, rollback policy,
Terraform drift, and operator evidence. RIID-4822 makes the workflow
infra-output-gated: if the optional CodeDeploy application and deployment-group
variables are both configured from infra evidence, the workflow creates same-job
temporary AppSpec/request JSON, creates the CodeDeploy deployment, waits for
success, and runs the same smoke checks. Public control-plane docs/workflows may
name stable configuration keys, but they must not commit or upload deployment
IDs, AppSpec/task-definition JSON, image digests, live URLs, ARNs, smoke
payloads, service role ARNs, target group/listener ARNs, or environment-specific
examples.
RIID-4839 narrows that public configuration surface further: deploy and smoke
workflows may reference only the stable `RIIDO_AI_SERVER_*` GitHub secret and
variable names listed in the runtime CD ownership manifest. Adding another
public key is a CD surface change; key values, live examples, generated deploy
payloads, image/task-definition values, CodeDeploy generated JSON, deployment
IDs, smoke payloads, and detailed evidence stay outside public repositories.
RIID-4842 adds the ratchet: public key names are a managed sensitivity budget.
The existing names can be documented only because operators need to configure
them, and any new `RIIDO_AI_SERVER_*` name must be added to the runtime CD
ownership manifest before public docs or workflows reference it.
RIID-4855 keeps CodeDeploy activation on the same side of the boundary:
topology and private evidence come from `riido-infra`, but the activation path
is still a `riido-control-plane` workflow mode. Public docs may describe the
operator/environment gate and stable categories, not live CodeDeploy values,
generated deployment payloads, image/task-definition values, smoke payloads, or
Terraform/operator evidence.

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
