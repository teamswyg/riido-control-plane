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
| `RIIDO_AI_SERVER_EXTERNAL_AUTHZ_TIMEOUT_SECONDS` | authorizer default when unset | `cmd/riido_ai_server` | positive integer timeout override for external authorizer requests |
| `RIIDO_AI_SERVER_REVIEW_ACCOUNT_TOKEN_SHA256` | empty | `cmd/riido_ai_server` | enables public-safe review/demo seed provisioning using only a token hash |
| `RIIDO_AI_SERVER_METRICS_LOG_INTERVAL_SECONDS` | disabled | `cmd/riido_ai_server` | positive integer interval for stdout CloudWatch EMF JSON Lines |
| `RIIDO_AI_SERVER_WEB_ALLOWED_ORIGINS` | empty | `cmd/riido_ai_server` | comma-separated exact `http://` or `https://` browser origins allowed to call the public HTTP API with CORS preflight support |
| `RIIDO_AI_SERVER_AI_AGENT_CLIENT_DYNAMODB_TABLE` | falls back to `RIIDO_AI_SERVER_DYNAMODB_ASSIGNMENT_TABLE` | `cmd/riido_ai_server` | DynamoDB table used by the development AI Agent client store for agents, devices, daemon snapshots, device credentials, task threads, and client events |
| `RIIDO_AI_SERVER_DYNAMODB_ASSIGNMENT_TABLE` | empty | `cmd/riido_ai_server` | DynamoDB table used by assignment snapshots and operations; also the AI Agent client table fallback |
| `RIIDO_AI_SERVER_DYNAMODB_OUTBOX_TABLE` | empty | `cmd/riido_ai_server` | DynamoDB table used by the assignment event outbox |
| `RIIDO_AI_SERVER_DYNAMODB_ENDPOINT` | empty | `cmd/riido_ai_server` | optional DynamoDB HTTP endpoint override for adapter tests or explicitly configured non-production AWS-compatible endpoints |
| `RIIDO_AI_SERVER_AWS_REGION` | `AWS_REGION` | `cmd/riido_ai_server` | AWS region used by DynamoDB adapters |
| `RIIDO_AI_SERVER_ASSIGNMENT_ACTIVE_LEASE_SECONDS` | assignment default | `cmd/riido_ai_server` | active assignment lease window used when opening the assignment store |
| `RIIDO_AI_SERVER_TASK_CONTEXT_BASE_URL` | empty | `cmd/riido_ai_server` | existing Riido API server base URL used for server-to-server task context lookup |
| `RIIDO_AI_SERVER_TASK_CONTEXT_WORKSPACE_ID` | empty | `cmd/riido_ai_server` | workspace id used in the existing API server task context path |
| `RIIDO_AI_SERVER_TASK_CONTEXT_TEAM_ID` | empty | `cmd/riido_ai_server` | team id used in the existing API server task context path |
| `RIIDO_AI_SERVER_TASK_CONTEXT_WORKSPACE_API_KEY` | empty | `cmd/riido_ai_server` | existing API server Open API key sent as `X-Workspace-Api-Key`; secret, server-only |
| `RIIDO_AI_SERVER_TASK_CONTEXT_TIMEOUT_SECONDS` | task-context client default when unset | `cmd/riido_ai_server` | positive integer timeout override for task context lookup requests |

All JSON env vars are decoded with unknown-field rejection and trailing-data
rejection. Empty values disable optional auth/task-context features rather than
selecting a production default.

`RIIDO_AI_SERVER_WEB_ALLOWED_ORIGINS` is not an authorization source. It only
enables browser transport from explicitly listed origins. Every protected API
still requires the bearer-token authorization rules owned by
[`../20-domain/request-authorization.md`](../20-domain/request-authorization.md).

The AI Agent client runtime is no longer selected by a local feature flag.
`cmd/riido_ai_server` opens the development AI Agent client store from DynamoDB
and fails startup when no AI Agent client table can be resolved. Static
`AgentRuntimeBinding` env injection is not part of the current runtime path:
agents, devices, daemon runtime snapshots, and DevicePrincipal credentials flow
through the persisted store.

When DynamoDB stores are configured in ECS/Fargate, the process obtains AWS
credentials from the container credential endpoint
(`AWS_CONTAINER_CREDENTIALS_FULL_URI` or
`AWS_CONTAINER_CREDENTIALS_RELATIVE_URI`). Raw AWS access keys are not part of
this config catalog.

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
