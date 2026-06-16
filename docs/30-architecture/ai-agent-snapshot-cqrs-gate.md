# AI Agent Snapshot CQRS Gate

> Riido task: RIID-4964 — AI Agent control-plane cost attribution and
> snapshot persistence review.

## Decision

Do not promote the whole control-plane Store to CQS/CQRS yet.

The CQRS candidate is limited to the AI Agent client snapshot. The current
evidence shows that the hot read/write path is the monolithic
`AIAgentClientSnapshot`, not the assignment queue/projection store as a whole.
The next structural change must therefore stay scoped to the AI Agent client
snapshot unless fresh measurements prove another Store operation is dominant.

## Evidence Shape

Sampled development traces now classify the path with domain Store operation
vocabulary:

- `/v1/daemon/runtime-snapshot`
  - `ai_agent_client_snapshot_load`
  - `ai_agent_client_snapshot_save`
- `/v1/daemon/agent-bindings`
  - `ai_agent_client_snapshot_load`
- `/v1/agents/{agent_id}/poll`
  - `store_poll_assignment`

This separates infrastructure calls such as DynamoDB `GetItem`/`PutItem` from
their domain intent. The important distinction is that `agent-bindings` is a
query path while `runtime-snapshot` is a command/sync path, but both currently
touch the same snapshot item.

## Measurement Gate

After the cadence fixes are live for a full 24 hour window, compare these
signals against the pre-change baseline:

- `ai_agent_client_snapshot_load_calls_total`
- `ai_agent_client_snapshot_save_calls_total`
- `ai_agent_client_snapshot_load_bytes_last`
- `ai_agent_client_snapshot_save_bytes_last`
- DynamoDB `ConsumedReadCapacityUnits`
- DynamoDB `ConsumedWriteCapacityUnits`
- X-Ray route/store-operation samples for `runtime-snapshot`, `agent-bindings`,
  and `poll`

The current cadence hardening keeps the monolithic model but moves the default
snapshot reload and no-change heartbeat save intervals to 15 seconds. This
stays below the 20 second daemon-runtime stale projection window while reducing
the repeated same-item reads/writes observed in development.

If read/write request units drop by at least 50% and snapshot operations no
longer dominate sampled Store traces, keep the monolithic snapshot and revisit
only when the API surface or data volume changes.

If request units drop by less than 50%, or the snapshot operations remain the
dominant Store cost after the cadence reduction, split only the AI Agent client
snapshot into smaller models.

## Candidate Split

The candidate split is micro-CQRS, not a platform-wide event-sourcing rewrite.

Command model:

- daemon runtime snapshot sync
- daemon heartbeat/status save
- device credential and daemon command mutations
- cold recovery checkpoint for full AI Agent client state

Query models:

- agent bindings read model for `/v1/daemon/agent-bindings`
- runtime/device status read model for client and settings views
- task-thread participant/progress projection only where the client needs it

The split should continue using low-cardinality trace vocabulary and must not
put task ids, agent ids, DynamoDB keys, table names, prompts, credentials,
profile URLs, or payload documents into trace attributes or public docs.

## Non-Goals

- Do not split assignment polling, lease, event append, or task event store
  paths without separate evidence.
- Do not introduce Kafka, a queue service, or a broad event-sourcing framework
  for this decision.
- Do not add a new public API contract merely to expose persistence internals.
- Do not encode live AWS account details, trace ids, log excerpts, or raw
  operator evidence in the public repository.
