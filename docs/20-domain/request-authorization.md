# Request Authorization

> Riido task: RIID-4664 `[Control Plane] external HTTP authorizer migration`

This document is the SSOT for the public control-plane request authorization
domain slice.

## Ownership

`internal/riidoaiserver.RequestAuthorizer` owns the server-side decision of
whether a request token can perform a resource/action request. The authorizer
returns a server-side `principal_id` plus optional control-plane roles. Client
payloads cannot directly submit an owner or role.

This slice owns:

- static token request-scope authorization
- SHA-256 token hash credentials for public-safe runtime configuration
- external HTTP request authorization behind the same port
- unauthenticated-only fallback chaining
- external authorizer JSON contract versions
- fail-closed response mapping for provider errors

This slice does not own:

- production IdP rollout
- tenant claim mapping, JWKS, or OIDC validation
- HTTP route wiring in `cmd/riido_ai_server`
- Terraform, SSM, secret rotation, or deployment evidence
- production request-token values

## Resources And Actions

Authorization requests use a resource/action pair plus optional agent or task
IDs. The current public domain slice supports these resource groups:

- `agent`
- `ai_agent_client`
- `agent_catalog`
- `component_task`
- `component_task_events`
- `metrics`

Scopes are endpoint-call gates only. Agent catalog reads and writes must still
re-run the RBAC decision described in
[`agent-catalog-rbac.md`](agent-catalog-rbac.md).

The `ai_agent_client` development API accepts these static scope families:

- `ai-agent:*`
- `ai-agent:read`
- `ai-agent:device:read`
- `ai-agent:stream`
- `ai-agent:write`
- `ai-agent:{agent_id}:read`
- `ai-agent:{agent_id}:update`
- `ai-agent:{agent_id}:delete`
- `task:{task_id}:read` for task participant dropdown reads
- `task:{task_id}:comment` for task-thread AI Agent comment submit
- `task:{task_id}:stop` or `task:{task_id}:write` for task-thread AI Agent stop

Those scopes only gate HTTP access. The development API still evaluates principal
ownership/admin roles before returning private agents or accepting mutations.

## External Authorizer Contract

`ExternalHTTPAuthorizer` sends `riido-external-authorizer-request.v1` as a JSON
POST to the configured endpoint.

The request includes:

- raw request token
- optional audience
- resource
- action
- optional workspace ID
- optional agent ID
- optional task ID

When the external authorizer is the existing Riido API server, the raw request
token is the existing Riido user JWT supplied by the web or desktop webview
client through `X-Riido-AI-Agent-Token`. The control plane still treats it as an
opaque request token: it does not decode Riido JWT claims locally. The existing
API server owns JWT verification, session/membership interpretation, and the
mapping from workspace membership to the control-plane `principal_id` and
optional `admin` role.

The current internal API server authorizer endpoint is
`POST /internal/control-plane/authorize`. That URL is deployment config, not a
generated frontend API. For browser-originated AI Agent client requests,
`request.resource` must remain `ai_agent_client` and `request.workspace_id` must
be present so the API server can evaluate the existing workspace membership
model. Daemon/device routes continue to use daemon/device credentials and must
not be authorized by a browser user JWT.

## DevicePrincipal Transport

Desktop-launched daemons are authorized as a DevicePrincipal, not as the browser
or desktop-webview UserPrincipal. The canonical ownership and secret handling
rules are in
`riido-contracts/docs/20-domain/device-principal.md`.

The desktop main process registers an account-owned device after an established
UserPrincipal session exists:

- `POST /v2/desktop/workspaces/{workspace_id}/devices/enroll`
- input token: `X-Riido-AI-Agent-Token`
- output credential: `device_id` plus one-time `device_secret`

The `workspace_id` in the enrollment URL is the selected workspace context for
authorization and audit. It does not make the device workspace-owned; devices
and runtimes remain account-owned.

When Desktop already has a local `device_id`, enrollment may send that
`device_id` again. If it belongs to the same user/workspace, control-plane keeps
the device principal stable and rotates only the one-time `device_secret`.
Stale credential recovery must not create a new device principal by accident,
because agent/runtime bindings are keyed through the device/runtime read model.

Daemon polling, heartbeat, progress, provider-status sync, and SaaS command
reads use:

- `X-Riido-Device-ID`
- `X-Riido-Device-Secret`

Those headers are daemon/server transport headers. They are not generated
frontend headers and must not be exposed through `riido-client`, browser
storage, webview JavaScript, task-thread events, logs, or status responses. The
server stores only a secret hash and returns the raw `device_secret` only in the
enrollment response.

A daemon runtime snapshot also treats `runtime_id` as a unique live runtime
identity. If the same `runtime_id` is reported by a different device, the
runtime moves to the latest reporting device and is removed from the previous
device projection. This keeps daemon `agent-bindings` deterministic and avoids
binding an agent to a stale device row.

Workspace admin authorization does not expose every device persisted in the
control-plane snapshot. AI Agent device/runtime reads remain scoped to the
current workspace: viewer-owned devices, same-workspace device credentials
allowed by admin RBAC, and device runtime rows exposed through a visible
workspace agent access path. A device enrolled for another workspace must stay
hidden from generated client bootstrap/devices responses and from replayed SSE
device events.

Development persistence uses the DynamoDB AI Agent client snapshot as the
durable authorization/projection SSOT. ECS task memory is a reloadable cache, so
device credential authorization and daemon `agent-bindings` reads must reload
the latest snapshot before deciding whether a daemon belongs to the user or
which device owns a runtime. This prevents a valid Desktop-enrolled daemon from
receiving a false 401 or a stale binding when ALB routes related requests to
different tasks.

The external authorizer hop may be protected by
`X-Riido-Control-Plane-Authorizer-Key`. That header is server-to-server only and
is never part of the generated frontend API. A browser client that lacks a
request token cannot be identified, so missing token input remains
unauthenticated.

The response must use `riido-external-authorizer-response.v1`.

When `allowed:true`, the response must include `principal_id`. The only role
accepted by this slice is `admin`; unsupported roles are service errors.

## Fail-Closed Rules

External authorizer response mapping is:

- HTTP 401 -> unauthenticated
- HTTP 403 -> forbidden
- HTTP 2xx with `allowed:false` -> forbidden
- non-2xx other than 401/403 -> service error
- malformed JSON -> service error
- unsupported schema version -> service error
- missing principal on `allowed:true` -> service error
- invalid role -> service error
- request timeout or network error -> service error

Service errors are intentionally not converted into allow decisions.

## Fallback Chain

`FallbackAuthorizer` evaluates authorizers in order. It calls the next
authorizer only when the current authorizer returns unauthenticated.

If a static token authorizer returns forbidden because a known token lacks the
requested scope, fallback must stop. This prevents a scoped deny from becoming
an external-provider allow.

## Browser Frontend Transport

Browser frontends call the same public HTTP endpoints as other clients. CORS is
therefore a transport allowlist, not an authorization decision.

Generated AI Agent clients must send the token through
`X-Riido-AI-Agent-Token`. This header name is intentionally different from the
legacy Riido app `token`/`Authorization` path so frontend call sites can tell
which SaaS token they are passing. Server-side authorization still treats the
value as the same opaque request token after extraction. If this header is
absent, and no compatibility `Authorization: Bearer ...` token is present, the
request is unauthenticated because the server has no principal to authorize.

`cmd/riido_ai_server` may configure exact browser origins through
`RIIDO_AI_SERVER_WEB_ALLOWED_ORIGINS`. When configured, `ServerConfig` handles
`OPTIONS` preflight before route authorization and allows only the HTTP methods
and headers required by the existing API surface: `GET`, `POST`, `PATCH`,
`DELETE`, `X-Riido-AI-Agent-Token`, `Authorization`, `Content-Type`, `Accept`,
and `Last-Event-ID`.

The server does not use wildcard origins and does not enable browser
credentials. Protected endpoints still require request-token authorization and,
where applicable, the agent-catalog RBAC decision described in
[`agent-catalog-rbac.md`](agent-catalog-rbac.md).

## Migration State

RIID-4664 moves the stdlib-only external HTTP authorizer adapter and
fail-closed tests from the former private `riido_daemon/internal/riidoaiserver`
package into this public repository.

RIID-4679 moves runtime environment parsing and HTTP route wiring into this
public repository. RIID-4691 reuses the static token hash path for review
account provisioning without storing raw token values.

RIID-4717 adds browser frontend CORS transport configuration over the existing
public HTTP API without changing request-token authorization, RBAC, or endpoint
payload contracts.

RIID-4721 adds the request-token-protected AI Agent client development API. It reuses
the same static/external authorizer port and keeps owner/public/private
visibility checks inside the route handler/store boundary.

RIID-4795 adds `X-Riido-AI-Agent-Token` as the canonical generated-client
transport header. `Authorization: Bearer ...` remains a compatibility input for
non-generated/internal callers, but generated web and desktop-webview clients no
longer use `token` or `Authorization` in their API wrapper.

This slice adds optional server-to-server authentication for the external
authorizer hop through `X-Riido-Control-Plane-Authorizer-Key`. This enables the
existing Riido API server to validate existing user JWTs for web/desktop-webview
requests without issuing a second browser token.

RIID-4869 adds the Desktop enrollment route and daemon DevicePrincipal header
verification for the control-plane slice. RIID-4872 promotes device credential
and AI Agent client state to the development DynamoDB snapshot store. Production
tenant claim mapping, JWKS/OIDC validation, credential rotation/revocation, and
production request-token values remain separate migration units. Daemons must
not reuse browser user JWTs.

Unresolved production identity mapping questions are tracked in
[`../50-roadmap/open-questions.md`](../50-roadmap/open-questions.md).
