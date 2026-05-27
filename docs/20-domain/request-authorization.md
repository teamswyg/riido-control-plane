# Request Authorization

> Riido task: RIID-4664 `[Control Plane] external HTTP authorizer migration`

This document is the SSOT for the public control-plane request authorization
domain slice.

## Ownership

`internal/riidoaiserver.RequestAuthorizer` owns the server-side decision of
whether a bearer token can perform a resource/action request. The authorizer
returns a server-side `principal_id` plus optional control-plane roles. Client
payloads cannot directly submit an owner or role.

This slice owns:

- static token request-scope authorization
- external HTTP request authorization behind the same port
- unauthenticated-only fallback chaining
- external authorizer JSON contract versions
- fail-closed response mapping for provider errors

This slice does not own:

- production IdP rollout
- tenant claim mapping, JWKS, or OIDC validation
- HTTP route wiring in `cmd/riido_ai_server`
- Terraform, SSM, secret rotation, or deployment evidence
- production bearer token values

## Resources And Actions

Authorization requests use a resource/action pair plus optional agent or task
IDs. The current public domain slice supports these resource groups:

- `agent`
- `agent_catalog`
- `component_task`
- `component_task_events`
- `metrics`

Scopes are endpoint-call gates only. Agent catalog reads and writes must still
re-run the RBAC decision described in
[`agent-catalog-rbac.md`](agent-catalog-rbac.md).

## External Authorizer Contract

`ExternalHTTPAuthorizer` sends `riido-external-authorizer-request.v1` as a JSON
POST to the configured endpoint.

The request includes:

- raw bearer token
- optional audience
- resource
- action
- optional agent ID
- optional task ID

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

## Migration State

RIID-4664 moves the stdlib-only external HTTP authorizer adapter and
fail-closed tests from the former private `riido_daemon/internal/riidoaiserver`
package into this public repository.

Runtime environment parsing, production IdP rollout, and HTTP server route
wiring remain separate migration units.
