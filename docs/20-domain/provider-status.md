# Provider Status SSOT

> Riido task: RIID-4671 `[Control Plane] provider status contract migration`

This document is the public SSOT for the control-plane provider status sync
contract. It is intentionally limited to server-facing request/response shapes,
validation, read/write authorization, and the narrow persistence ports required
by HTTP adapters.

## Responsibility

`riido-control-plane` owns the server contract for receiving and reading the
latest provider routing snapshot for an agent identity.

This boundary owns:

- `ProviderStatusRecord`
- `ProviderStatusSyncRequest`
- `ProviderStatusSyncResponse`
- `ProviderStatusStore`
- `ProviderStatusReader`
- `POST /v1/agents/{agent_id}/provider-status`
- `GET /v1/agents/{agent_id}/provider-status`

This boundary does not own customer-PC provider process execution, executable
path detection, login detection, provider provenance collection, store-channel
policy decisions, assignment routing decisions, Terraform, AWS, deployment
state, or raw credentials.

## Shared Vocabulary

Provider status uses shared public contracts from `riido-contracts`:

- `provider/capability.ProviderKind`
- `hostintegration.DistributionChannel`
- `hostintegration.ProviderRoutingStatus`

`ProviderRoutingStatus` accepts only:

- `available`
- `login-required`
- `unsupported`
- `store-blocked`

`DistributionChannel` accepts only:

- `developer-id`
- `mac-app-store`
- `msix-sideload`
- `msix-store`
- `dev-local`

## Sync Validation

`normalizeProviderStatusSync(agent_id, request)` is the executable validation
gate for this contract. It must:

- reject blank `agent_id`
- trim `daemon_id`, `device_id`, `runtime_id`, and `app_version`
- reject blank `daemon_id`
- reject blank `runtime_id`
- reject unknown `distribution_channel`
- reject missing provider rows
- trim each `provider_kind`
- reject blank or duplicated provider kinds
- reject unknown `routing_status`
- sort providers by `provider_kind` before persistence/response

The request/response intentionally exclude executable paths, workspace absolute
paths, provider tokens, API keys, raw environment, and private host evidence.

## Authorization

Provider status sync and read use the generic agent authorization resource:

- write: `agent:{agent_id}:provider-status:write`
- read: `agent:{agent_id}:provider-status:read`

Wildcard scopes such as `agent:*` and `riido:*` are interpreted by the request
authorization SSOT. Review/demo credentials may have read-only provider status
scope but must not receive daemon poll, heartbeat, event-write, or provider
status write scope unless a separate SSOT explicitly allows it.

## Migration State

RIID-4671 moves the provider status DTO, normalization gate, read/write port,
and HTTP provider-status route from the former private
`riido_daemon/internal/riidoaiserver` package into this public repository.

Runtime detector implementation, store-safe routing, assignment integration,
review account seed runtime wiring, durable store actors, DynamoDB adapters,
Terraform, AWS configuration, and deployment evidence remain separate migration
units.
