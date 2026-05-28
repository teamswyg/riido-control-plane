# Control Plane Config Reference

> Riido task: RIID-4712 `[Control Plane] Architecture SSOT docs migration`

This file is the public Factor 12 configuration catalog for
`cmd/riido_ai_server`.

## Runtime Env

| Variable | Default | Owner | Meaning |
| --- | --- | --- | --- |
| `RIIDO_AI_SERVER_ADDR` | `:8080` | `cmd/riido_ai_server` | HTTP listen address passed to `http.Server.Addr` |
| `RIIDO_AI_SERVER_SHUTDOWN_TIMEOUT_SECONDS` | `10` | `cmd/riido_ai_server` | graceful shutdown timeout after SIGINT/SIGTERM |
| `RIIDO_AI_SERVER_AGENT_BINDINGS_JSON` | empty | `cmd/riido_ai_server` | strict JSON array of `AgentRuntimeBinding` records for static daemon/runtime binding validation |
| `RIIDO_AI_SERVER_AUTHZ_TOKENS_JSON` | empty | `cmd/riido_ai_server` | strict JSON array of static token credentials for public-testable request authorization |
| `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_URL` | empty | `cmd/riido_ai_server` | optional external HTTP authorizer endpoint |
| `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_AUDIENCE` | empty | `cmd/riido_ai_server` | optional audience forwarded to the external authorizer |
| `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_TIMEOUT_SECONDS` | authorizer default when unset | `cmd/riido_ai_server` | positive integer timeout override for external authorizer requests |
| `RIIDO_AI_SERVER_REVIEW_ACCOUNT_TOKEN_SHA256` | empty | `cmd/riido_ai_server` | enables public-safe review/demo seed provisioning using only a token hash |
| `RIIDO_AI_SERVER_METRICS_LOG_INTERVAL_SECONDS` | disabled | `cmd/riido_ai_server` | positive integer interval for stdout CloudWatch EMF JSON Lines |
| `RIIDO_AI_SERVER_WEB_ALLOWED_ORIGINS` | empty | `cmd/riido_ai_server` | comma-separated exact `http://` or `https://` browser origins allowed to call the public HTTP API with CORS preflight support |

All JSON env vars are decoded with unknown-field rejection and trailing-data
rejection. Empty values disable the optional feature rather than selecting a
production default.

`RIIDO_AI_SERVER_WEB_ALLOWED_ORIGINS` is not an authorization source. It only
enables browser transport from explicitly listed origins. Every protected API
still requires the bearer-token authorization rules owned by
[`../20-domain/request-authorization.md`](../20-domain/request-authorization.md).

## Non-Config Facts

The following are not owned by this repository's runtime config catalog:

- AWS account IDs, subnet IDs, security group IDs, domain names, ACM
  certificates, Route53 zones, WAF IDs, ECR repositories, ECS task definitions
- Terraform backend settings, state files, tfvars, plans, or apply logs
- raw bearer tokens, review account raw token values, IdP client secrets, or AWS
  credentials
- daemon provider executable paths or customer workspace roots

Those values belong to `riido-infra`, secret managers, or the daemon
configuration catalog.

## Validation

The public checks must prove that every env var parsed by `cmd/riido_ai_server`
is documented here. If a new variable is added in code, this file and the
architecture-doc workflow must change in the same PR.
