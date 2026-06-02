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
| `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_API_KEY` | empty | `cmd/riido_ai_server` | optional server-to-server key sent to the external authorizer as `X-Riido-Control-Plane-Authorizer-Key` |
| `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_TIMEOUT_SECONDS` | authorizer default when unset | `cmd/riido_ai_server` | positive integer timeout override for external authorizer requests |
| `RIIDO_AI_SERVER_REVIEW_ACCOUNT_TOKEN_SHA256` | empty | `cmd/riido_ai_server` | enables public-safe review/demo seed provisioning using only a token hash |
| `RIIDO_AI_SERVER_METRICS_LOG_INTERVAL_SECONDS` | disabled | `cmd/riido_ai_server` | positive integer interval for stdout CloudWatch EMF JSON Lines |
| `RIIDO_AI_SERVER_WEB_ALLOWED_ORIGINS` | empty | `cmd/riido_ai_server` | comma-separated exact `http://` or `https://` browser origins allowed to call the public HTTP API with CORS preflight support |
| `RIIDO_AI_SERVER_AI_AGENT_CLIENT_DEVELOPMENT` | `false` | `cmd/riido_ai_server` | enables the development AI Agent client API backed by DynamoDB snapshot persistence |
| `RIIDO_AI_SERVER_AI_AGENT_CLIENT_MOCK` | `false` | `cmd/riido_ai_server` | deprecated compatibility alias for `RIIDO_AI_SERVER_AI_AGENT_CLIENT_DEVELOPMENT` |
| `RIIDO_AI_SERVER_AI_AGENT_CLIENT_DYNAMODB_TABLE` | empty | `cmd/riido_ai_server` | DynamoDB table name for the AI Agent client development snapshot item; required when development mode is enabled |
| `RIIDO_AI_SERVER_AWS_REGION` | empty | `cmd/riido_ai_server` | AWS region used by stdlib-only DynamoDB request signing; required when development mode is enabled |
| `RIIDO_AI_SERVER_DYNAMODB_ENDPOINT` | AWS default for region | `cmd/riido_ai_server` | optional DynamoDB endpoint override for fake-endpoint tests or local development |
| `AWS_CONTAINER_CREDENTIALS_FULL_URI` | empty | AWS/ECS runtime | optional ECS credential endpoint used to sign DynamoDB requests |
| `AWS_CONTAINER_CREDENTIALS_RELATIVE_URI` | empty | AWS/ECS runtime | optional ECS relative credential endpoint; used with `http://169.254.170.2` when the full URI is absent |
| `AWS_CONTAINER_AUTHORIZATION_TOKEN` | empty | AWS/ECS runtime | optional authorization token forwarded to the ECS credential endpoint |
| `RIIDO_AI_SERVER_TASK_CONTEXT_BASE_URL` | empty | `cmd/riido_ai_server` | existing Riido API server base URL used for server-to-server task context lookup |
| `RIIDO_AI_SERVER_TASK_CONTEXT_WORKSPACE_ID` | empty | `cmd/riido_ai_server` | workspace id used in the existing API server task context path |
| `RIIDO_AI_SERVER_TASK_CONTEXT_TEAM_ID` | empty | `cmd/riido_ai_server` | team id used in the existing API server task context path |
| `RIIDO_AI_SERVER_TASK_CONTEXT_WORKSPACE_API_KEY` | empty | `cmd/riido_ai_server` | existing API server Open API key sent as `X-Workspace-Api-Key`; secret, server-only |
| `RIIDO_AI_SERVER_TASK_CONTEXT_TIMEOUT_SECONDS` | task-context client default when unset | `cmd/riido_ai_server` | positive integer timeout override for task context lookup requests |

All JSON env vars are decoded with unknown-field rejection and trailing-data
rejection. Empty values disable the optional feature rather than selecting a
production default.

`RIIDO_AI_SERVER_WEB_ALLOWED_ORIGINS` is not an authorization source. It only
enables browser transport from explicitly listed origins. Every protected API
still requires the bearer-token authorization rules owned by
[`../20-domain/request-authorization.md`](../20-domain/request-authorization.md).

`RIIDO_AI_SERVER_AI_AGENT_CLIENT_DEVELOPMENT` is the development/testnet switch.
It does not enable unauthenticated access; all AI Agent client endpoints still
require bearer scopes. When enabled, the server fails during startup unless the
DynamoDB table, AWS region, and ECS credential endpoint configuration are
available. The development store persists the whole AI Agent client read/write
state as a schema-versioned DynamoDB snapshot item. `RIIDO_AI_SERVER_AI_AGENT_CLIENT_MOCK`
is kept only so older deployment configuration fails forward into the same
development mode.

The external authorizer API key is server-to-server authentication for the
configured authorizer endpoint. It is not a generated frontend token. If set,
`RIIDO_AI_SERVER_EXTERNAL_AUTHZ_URL` must also be set, and the value is sent only
through `X-Riido-Control-Plane-Authorizer-Key` on the control-plane to
authorizer hop.

The task-context base URL, workspace id, team id, and workspace API key must be
set together. If only part of the group is set, the server fails during startup.
The workspace API key is never a generated client token; frontend clients should
continue using the AI Agent SaaS bearer header defined in the generated API
surface.

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
