# riido-control-plane

`riido-control-plane`은 Riido SaaS control plane의 공개 backend boundary입니다.
web client와 desktop app webview가 호출하는 HTTP/SSE API, assignment polling,
provider status, authorization port, RBAC read model, mock/testnet API를
소유합니다.

이 레포는 provider CLI를 실행하지 않습니다. 런타임 실행과 로컬 디바이스 제어는
`riido-daemon`의 책임이고, Terraform과 AWS 배포 구성은 `riido-infra`의
책임입니다. 이 레포는 공개 가능한 서버 코드와 검증 가능한 API surface만
담습니다.

## 이 레포가 하는 일

- Riido SaaS HTTP/SSE endpoint를 구현합니다.
- request-token authorization port와 static/external authorizer adapter를
  제공합니다.
- agent catalog, AI Agent client API, provider status, assignment polling
  같은 control-plane domain slice를 검증합니다.
- `riido-contracts`의 canonical DSL/IR을 기준으로 AI Agent client API
  sub-DSL -> OpenAPI projection -> generated React Query client가 drift 나지
  않도록 합니다.
- 공개 GitHub Actions에서 black-box/domain/generator 검증을 실행합니다.

## 이 레포가 하지 않는 일

- Terraform state, AWS 계정 topology, ECR push 설정, Fargate task-definition
  wiring을 커밋하지 않습니다.
- raw request token, IdP secret, customer data export를 소유하지 않습니다.
- provider runtime process를 실행하거나 provider CLI binary를 번들하지 않습니다.
- production persistence와 DNS 운영 evidence를 SSOT로 삼지 않습니다.

## 왜 이 작업을 여기서 했나

AI Agent client-facing endpoint는 Riido web과 desktop webview가 직접 호출하는
SaaS API입니다. 따라서 handler, auth scope gate, mock store, SSE replay,
React Query generated client는 `riido-control-plane`에서 함께 검증해야 합니다.

다만 canonical business vocabulary와 lifecycle/deprecation grammar의 최종
권한은 `riido-contracts`에 있습니다. 이 레포는 그 언어를 import해 AI Agent
client API sub-DSL과 handler/generator delivery boundary를 소유합니다.
client 사용성 때문에 API surface 변경이 필요하면 `riido-control-plane`에서
시작하고, 비즈니스 의미가 바뀌는 순간 `riido-contracts`로 escalation합니다.

새로 추가되는 AI Agent client API는 workspace-aware v2 surface입니다. v1은
기존 UI 테스트와 임시 client 작업을 깨지 않기 위한 호환 경로로 유지하고, 새
client 코드는 generated facade에서 `riido.v2.aiAgent.agents.create`처럼
버전이 루트에 보이는 라이브러리 경로를 사용합니다.

## 어떤 문서를 보면 되나

| 알고 싶은 것 | 문서 |
| --- | --- |
| AI Agent client API endpoint와 mock/testnet 정책 | [`docs/20-domain/ai-agent-client-api.md`](docs/20-domain/ai-agent-client-api.md) |
| authorization resource/action/scope 규칙 | [`docs/20-domain/request-authorization.md`](docs/20-domain/request-authorization.md) |
| agent catalog RBAC 규칙 | [`docs/20-domain/agent-catalog-rbac.md`](docs/20-domain/agent-catalog-rbac.md) |
| runtime/agent binding domain | [`docs/20-domain/agent-runtime-binding.md`](docs/20-domain/agent-runtime-binding.md) |
| control-plane bounded context | [`docs/20-domain/context-map.md`](docs/20-domain/context-map.md) |
| riido-client generated React Query 전달 정책 | [`docs/30-architecture/api-client-delivery.md`](docs/30-architecture/api-client-delivery.md) |
| module/package 책임 분해 | [`docs/30-architecture/module-decomposition.md`](docs/30-architecture/module-decomposition.md) |
| runtime env 변수와 설정 책임 | [`docs/30-architecture/config-reference.md`](docs/30-architecture/config-reference.md) |
| daemon/contracts/infra/client와의 연결 | [`docs/30-architecture/integration-matrix.md`](docs/30-architecture/integration-matrix.md) |
| public runtime과 deploy boundary | [`docs/30-architecture/runtime-deployment-boundary.md`](docs/30-architecture/runtime-deployment-boundary.md) |
| 마이그레이션 히스토리 | [`docs/migration/control-plane.md`](docs/migration/control-plane.md) |

## AI Agent mock testnet

AI Agent client mock API는 다음 env로 켭니다.

```bash
RIIDO_AI_SERVER_AI_AGENT_CLIENT_MOCK=true
```

mock API도 인증 없이 열리지 않습니다. request-token scope와 owner/public/private
visibility policy를 통과해야 합니다.

현재 testnet smoke는 별도 GitHub Actions workflow가 담당합니다.

- workflow: `ai-agent-client-testnet-smoke`
- base URL: `RIIDO_AI_SERVER_TESTNET_BASE_URL`
- AI Agent token secret: `RIIDO_AI_SERVER_TESTNET_TOKEN`
- 현재 testnet URL: `http://ai-api.riido.io`

검증하는 endpoint:

- `GET /healthz`
- `GET /readyz`
- `GET /v1/client/ai-agent/bootstrap`
- `GET /v1/client/ai-agent/devices`
- `GET /v1/client/ai-agent/agents/{agent_id}/daemon`
- `POST /v1/client/ai-agent/agents/{agent_id}/daemon/start`
- `POST /v1/client/ai-agent/agents/{agent_id}/daemon/restart`
- `POST /v1/client/ai-agent/agents/{agent_id}/daemon/stop`
- `GET /v1/client/ai-agent/tasks/{task_id}/assignable-agents`
- `POST /v1/client/ai-agent/tasks/{task_id}/assignment`
- `DELETE /v1/client/ai-agent/tasks/{task_id}/assignment`
- `POST /v1/client/ai-agent/tasks/{task_id}/comments`
- `POST /v1/client/ai-agent/tasks/{task_id}/stop`
- `GET /v1/client/ai-agent/events?replay=1`
- v2 workspace-scoped duplicate:
  `/v2/client/workspaces/{workspace_id}/ai-agent/...`

## Contract / generated client 흐름

```text
riido-contracts canonical DSL/IR
  -> control-plane AI Agent client API sub-DSL
  -> control-plane API IR
  -> OpenAPI projection
  -> contracts/ai-agent-client/*.json
  -> tools/reactquerygen
  -> web/generated/aiAgentClient.ts
```

OpenAPI와 generated client는 사람이 임의로 고치는 SSOT가 아닙니다. API surface
계약이 바뀌면 control-plane API sub-DSL을 먼저 바꾸고 projection과 generated
client를 다시 생성해야 합니다. canonical vocabulary나 비즈니스 정책이 바뀌는
경우에는 `riido-contracts` 변경이 선행되어야 합니다.

`riido-client`로 React Query 코드를 전달하는 cross-repo workflow는
[`docs/30-architecture/api-client-delivery.md`](docs/30-architecture/api-client-delivery.md)의
tag-triggered delivery 정책을 따라야 합니다. 이 레포의 workflow 산출물만 신뢰하고
client repo에서 Orval을 직접 실행하지 않는 것이 원칙입니다.

## 검증

```bash
go test ./...
go list -m all
go test ./internal/riidoaiserver -run 'AIAgentClient' -count=1
go test ./cmd/riido_ai_server -run 'AIAgentClientMock|ConfigFromEnv' -count=1
go test ./tools/reactquerygen -count=1
go run ./tools/reactquerygen -openapi contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json -out web/generated/aiAgentClient.ts
go test ./tools/containercontract -count=1
go run ./tools/containercontract -contract packaging/containers/riido_ai_server_container.riido.json -out -
docker build -f packaging/containers/riido_ai_server.Dockerfile -t riido-control-plane:local .
```

CI는 public repo에서 가벼운 검증을 돌리기 위한 경계입니다. private infra
billing pool에서 테스트 비용이 커지지 않도록, 배포 wiring은 `riido-infra`에
두고 API/generator/smoke 검증은 이 레포에서 수행합니다.

## Module

```text
github.com/teamswyg/riido-control-plane
```

## License

Apache-2.0.
