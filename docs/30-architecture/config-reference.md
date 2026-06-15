# Control Plane Config Reference

> Riido task: RIID-4712 `[Control Plane] Architecture SSOT docs migration`

This file is the public Factor 12 configuration catalog for
`cmd/riido_ai_server`.

## Runtime Env

| Variable | Default | Owner | Meaning |
| --- | --- | --- | --- |
| `RIIDO_AI_SERVER_ADDR` | `:8080` | `cmd/riido_ai_server` | HTTP listen address passed to `http.Server.Addr` |
| `RIIDO_AI_SERVER_SHUTDOWN_TIMEOUT_SECONDS` | `10` | `cmd/riido_ai_server` | graceful shutdown timeout after SIGINT/SIGTERM |
| `RIIDO_AI_SERVER_AUTHZ_TOKENS_JSON` | empty | `cmd/riido_ai_server` | strict JSON array of static token credentials for public-testable request authorization |
| `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_URL` | empty | `cmd/riido_ai_server` | optional external HTTP authorizer endpoint |
| `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_AUDIENCE` | empty | `cmd/riido_ai_server` | optional audience forwarded to the external authorizer |
| `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_API_KEY` | empty | `cmd/riido_ai_server` | optional server-to-server key sent to the external authorizer as `X-Riido-Control-Plane-Authorizer-Key` |
| `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_TIMEOUT_SECONDS` | authorizer default when unset | `cmd/riido_ai_server` | positive integer timeout override for external authorizer requests |
| `RIIDO_AI_SERVER_REVIEW_ACCOUNT_TOKEN_SHA256` | empty | `cmd/riido_ai_server` | enables public-safe review/demo seed provisioning using only a token hash |
| `RIIDO_AI_SERVER_METRICS_LOG_INTERVAL_SECONDS` | disabled | `cmd/riido_ai_server` | positive integer interval for stdout CloudWatch EMF JSON Lines |
| `RIIDO_AI_SERVER_WEB_ALLOWED_ORIGINS` | empty | `cmd/riido_ai_server` | comma-separated exact `http://` or `https://` browser origins allowed to call the public HTTP API with CORS preflight support |
| `RIIDO_AI_SERVER_ASSIGNMENT_ACTIVE_LEASE_SECONDS` | store default (`20`) | `cmd/riido_ai_server` | optional active assignment lease duration shared by the store actor and DynamoDB active-lease adapter |
| `RIIDO_AI_SERVER_LONGPOLL_MAX_HOLD_SECONDS` | `25` | `cmd/riido_ai_server` | max time a daemon claim poll (`PollRequest.wait_ms`) is held open; must stay under the ALB idle timeout. See [`../20-domain/saas-control-plane.md`](../20-domain/saas-control-plane.md) |
| `RIIDO_AI_SERVER_LONGPOLL_TICK_SECONDS` | `2` | `cmd/riido_ai_server` | fallback re-evaluation interval during a held poll; bounds cross-instance assignment discovery latency |
| `RIIDO_AI_SERVER_AI_AGENT_CLIENT_DEVELOPMENT` | `false` | `cmd/riido_ai_server` | enables the development AI Agent client API backed by DynamoDB snapshot and assignment-operation persistence |
| `RIIDO_AI_SERVER_AI_AGENT_CLIENT_DYNAMODB_TABLE` | empty | `cmd/riido_ai_server` | DynamoDB table name for the AI Agent client development snapshot item and assignment operation journal/queue; required when development mode is enabled |
| `RIIDO_AI_SERVER_AWS_REGION` | empty | `cmd/riido_ai_server` | AWS region used by stdlib-only DynamoDB request signing; required when development mode is enabled |
| `RIIDO_AI_SERVER_DYNAMODB_ENDPOINT` | AWS default for region | `cmd/riido_ai_server` | optional DynamoDB endpoint override for fake-endpoint tests or local development |
| `RIIDO_AI_SERVER_AGENT_PROFILE_THUMBNAIL_BUCKET` | empty | `cmd/riido_ai_server` | existing deployment-owned S3 bucket used to sign AI Agent profile thumbnail POST upload intents; required with the CDN base URL to enable the endpoint |
| `RIIDO_AI_SERVER_AGENT_PROFILE_THUMBNAIL_PREFIX` | `thumbnail/ai/profile/` | `cmd/riido_ai_server` | object key prefix used when issuing AI Agent profile thumbnail upload intents |
| `RIIDO_AI_SERVER_AGENT_PROFILE_THUMBNAIL_CDN_BASE_URL` | empty | `cmd/riido_ai_server` | HTTPS CDN base URL returned to clients as the saved `profile_thumbnail_url`; required with the bucket to enable the endpoint |
| `RIIDO_AI_SERVER_AGENT_PROFILE_THUMBNAIL_MAX_BYTES` | `5242880` | `cmd/riido_ai_server` | positive maximum image size accepted by the upload-intent policy |
| `RIIDO_AI_SERVER_AGENT_PROFILE_THUMBNAIL_UPLOAD_EXPIRES_SECONDS` | `300` | `cmd/riido_ai_server` | positive lifetime for the one-time S3 POST upload intent |
| `RIIDO_AI_SERVER_AGENT_PROFILE_THUMBNAIL_S3_ENDPOINT` | AWS bucket endpoint for region | `cmd/riido_ai_server` | optional S3-compatible endpoint override for tests; production should use the regional S3 endpoint |
| `AWS_CONTAINER_CREDENTIALS_FULL_URI` | empty | AWS/ECS runtime | optional ECS credential endpoint used to sign DynamoDB requests |
| `AWS_CONTAINER_CREDENTIALS_RELATIVE_URI` | empty | AWS/ECS runtime | optional ECS relative credential endpoint; used with `http://169.254.170.2` when the full URI is absent |
| `AWS_CONTAINER_AUTHORIZATION_TOKEN` | empty | AWS/ECS runtime | optional authorization token forwarded to the ECS credential endpoint |
| `RIIDO_AI_SERVER_TASK_CONTEXT_BASE_URL` | empty | `cmd/riido_ai_server` | existing Riido API server base URL used for server-to-server task context lookup |
| `RIIDO_AI_SERVER_TASK_CONTEXT_WORKSPACE_ID` | empty | `cmd/riido_ai_server` | legacy Open API task-context adapter input; outside generated AI Agent assignment |
| `RIIDO_AI_SERVER_TASK_CONTEXT_TEAM_ID` | empty | `cmd/riido_ai_server` | legacy Open API task-context adapter input; outside generated AI Agent assignment |
| `RIIDO_AI_SERVER_TASK_CONTEXT_WORKSPACE_API_KEY` | empty | `cmd/riido_ai_server` | legacy Open API task-context adapter key sent as `X-Workspace-Api-Key`; outside generated AI Agent assignment |
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
state as a schema-versioned DynamoDB snapshot item and persists assignment
operations in the same table so generated assignment requests and daemon poll
requests can cross ECS instance or restart boundaries.

Profile thumbnail upload intents are optional and fail closed. If any profile
thumbnail upload env var is set, the bucket, CDN base URL, AWS region, and ECS
credential endpoint must be present. The endpoint returns signed S3 POST form
fields plus the CDN URL that clients save as `profile_thumbnail_url`; the server
does not proxy image bytes and public docs must not include bucket values,
account IDs, credentials, or live upload evidence.

The external authorizer API key is server-to-server authentication for the
configured authorizer endpoint. It is not a generated frontend token. If set,
`RIIDO_AI_SERVER_EXTERNAL_AUTHZ_URL` must also be set, and the value is sent only
through `X-Riido-Control-Plane-Authorizer-Key` on the control-plane to
authorizer hop.

The generated AI Agent assignment/auth path does not depend on `team_id`,
`teamId`, OpenAPI task-context paths, or Open API key transport such as
`X-Workspace-Api-Key`. The task-context base URL can run by itself through the
private JWT task-context reader. The workspace id, team id, and workspace API
key are only the legacy Open API task-context reader group; if any one of that
group is set, all three must be set. The workspace API key is never a generated
client token; frontend clients should continue using the AI Agent SaaS bearer
header defined in the generated API surface. Daemon poll/heartbeat/event and
client assignment/thread projection bugs must be debugged through the
device-principal and assignment-store path, not through this legacy Open API
configuration group.

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
