# Control Plane Integration Matrix

> Riido task: RIID-4712 `[Control Plane] Architecture SSOT docs migration`

Control-plane verification is split into deterministic public CI and
operator/private infra validation.

## Public Deterministic Gates

| Surface | Verification | External dependency |
| --- | --- | --- |
| module dependency boundary | `go list -m all` allowlist | none |
| full backend behavior | `go test ./...` | none |
| agent catalog RBAC and HTTP | focused `internal/riidoaiserver` black-box tests | none |
| AI Agent client mock API | focused `internal/riidoaiserver` black-box tests over generated contract paths | none |
| generated React Query client | `tools/reactquerygen` drift test from checked-in OpenAPI | none |
| future `riido-client` generated delivery | tag-triggered control-plane workflow, target-path allowlist, generated manifest/history diff | GitHub API only after delivery secrets are configured |
| request authorization | static tokens and `httptest` external authorizer tests | none |
| assignment polling/heartbeat/events | in-memory store and HTTP tests | none |
| SSE | `httptest` streaming tests | none |
| web frontend API transport | CORS allowlist/preflight HTTP tests and env parser tests | none |
| metrics HTTP and stdout EMF | metrics read-model and writer tests | none |
| DynamoDB/EventBridge adapters | fake endpoint HTTP tests with fake credentials | no live AWS |
| `awsadapters` facade | compile and usage tests | none |
| container image contract | `tools/containercontract` and optional local Docker build | Docker only for image build check |
| AI Agent testnet runtime CD | tag-triggered `deploy-ai-agent-testnet` workflow: build, ECR push, ECS service wait, live smoke | GitHub OIDC, masked AWS account boundary, configured testnet secrets, no live URL dispatch input, no live-value step output, same-job temp handoff cleanup |
| infra-output-gated CodeDeploy runtime CD | same workflow creates/waits/smokes CodeDeploy deployment only after infra supplies topology evidence and both CodeDeploy variables are configured from infra outputs | GitHub OIDC, masked deployment target config, same-job temp AppSpec/request files with traps, no service-role/target-group/listener ARN input, no uploaded deployment artifacts or live-value step output |

Public PR checks must not require AWS credentials, Terraform state, ECR access,
production secret material, or write access to `riido-client`. Testnet runtime CD
is not a pull-request gate; it runs only for `v*` tag pushes or explicit manual
dispatch.

## Private / Operator Gates

| Surface | Owner | Evidence |
| --- | --- | --- |
| Terraform plan/apply | `riido-infra` | typed plan/apply evidence and Terraform work-unit records |
| ECR repository/topology | `riido-infra` | Terraform plan/apply evidence |
| testnet ECR image push | `riido-control-plane` tag CD | immutable tag, image digest, masked workflow logs |
| ECS/Fargate topology | `riido-infra` | Terraform plan/apply evidence and drift policy |
| testnet ECS service deployment | `riido-control-plane` tag CD | task-definition revision, service-stability wait, AI Agent smoke |
| CodeDeploy application/deployment group topology | `riido-infra` | RIID-4822 Terraform plan/apply evidence, rollback/traffic-shift policy, and exported application/deployment group names before the control-plane optional mode is configured |
| DNS/ACM/WAF/public ingress | `riido-infra` | traffic, certificate, and ingress evidence |
| production secret wiring | `riido-infra` plus secret manager | redacted runtime secret evidence |
| live DynamoDB/EventBridge behavior | `riido-infra` | backend bootstrap and traffic/evidence tools |

Private gates may consume public module tags, container image digests, workflow
run URLs, and the `awsadapters` facade. They must not push account-specific
state, raw secret values, or unredacted live evidence back into this public
repository.

## Optional Local Commands

```bash
go test ./...
go list -m all
go test ./awsadapters -count=1
go test ./internal/riidoaiserver -run 'WebFrontendCORS' -count=1
go test ./internal/riidoaiserver -run 'AIAgentClient' -count=1
go test ./cmd/riido_ai_server -run 'WebAllowedOrigins|ConfigFromEnv' -count=1
go test ./tools/reactquerygen -count=1
git diff --check
go test ./tools/containercontract -count=1
go run ./tools/containercontract -contract packaging/containers/riido_ai_server_container.riido.json -out -
docker build -f packaging/containers/riido_ai_server.Dockerfile -t riido-control-plane:local .
```

The Docker build is useful before release but may be skipped where Docker is not
available. The Go tests and contract verifier are the minimum public gate.
