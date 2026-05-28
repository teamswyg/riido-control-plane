# riido-control-plane

SaaS control-plane backend for Riido.

This repository is the public backend boundary. It will own assignment polling,
event ingest, SSE/read-model APIs, authorization ports, RBAC behavior, and
control-plane black-box/domain tests.

It consumes shared public contracts from
`github.com/teamswyg/riido-contracts` at `v0.3.0`. Runtime provider execution,
Terraform deployment wiring, and secret material stay outside this repository.

## Module

```text
github.com/teamswyg/riido-control-plane
```

## Repository Boundary

This repository may contain:

- HTTP/SSE control-plane server code
- assignment, event, RBAC, provider status, and read-model domain logic
- store-review/demo seed artifacts that contain no raw tokens or provider
  execution grants
- public API contracts implemented by the control plane
- public Docker image contracts that do not publish or deploy artifacts
- black-box and domain scenario tests

This repository must not contain:

- Terraform state, AWS account details, or deployment secrets
- environment-specific production values
- customer data exports
- private release artifacts
- ECR push configuration, image digest deployment evidence, or private
  Fargate task-definition wiring

## Verification

Optional store-review/demo access is enabled by setting only
`RIIDO_AI_SERVER_REVIEW_ACCOUNT_TOKEN_SHA256`. The raw review token is never
committed or read from this repository. Optional CloudWatch Embedded Metric
Format JSONL logs can be enabled with
`RIIDO_AI_SERVER_METRICS_LOG_INTERVAL_SECONDS`; the writer uses stdout only and
does not require AWS SDKs, credentials, log groups, or Terraform state.
The temporary AI Agent client mock API is enabled with
`RIIDO_AI_SERVER_AI_AGENT_CLIENT_MOCK=true` and remains bearer-token protected.

```bash
go test ./...
go list -m all
go test ./internal/riidoaiserver -run 'AIAgentClient' -count=1
go test ./tools/reactquerygen -count=1
go run ./tools/reactquerygen -openapi contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json -out web/generated/aiAgentClient.ts
go test ./tools/containercontract -count=1
go run ./tools/containercontract -contract packaging/containers/riido_ai_server_container.riido.json -out -
docker build -f packaging/containers/riido_ai_server.Dockerfile -t riido-control-plane:local .
```

The public GitHub Actions workflow runs the lightweight verification suite
outside the private infrastructure repository billing pool. CI allows only
Riido-owned Go module dependencies; any third-party dependency requires a new
documented decision before it is introduced.

The deployed AI Agent mock testnet is checked by the separate
`ai-agent-client-testnet-smoke` workflow. It calls the ALB URL from the
`RIIDO_AI_SERVER_TESTNET_BASE_URL` repository variable or a manual workflow
input, and reads the bearer token from the `RIIDO_AI_SERVER_TESTNET_TOKEN`
repository secret.

## License

Apache-2.0.
