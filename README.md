# riido-control-plane

`riido-control-plane`은 Riido SaaS control plane의 공개 backend boundary입니다.
web client와 desktop app webview가 호출하는 HTTP/SSE API, assignment polling,
provider status, authorization port, RBAC read model, development/testnet API를
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
- tag 기반 testnet CD에서 컨테이너 이미지를 빌드하고 ECS 서비스 안정화 후
  smoke를 검증합니다. AWS topology와 secret 값은 커밋하지 않습니다.

## 이 레포가 하지 않는 일

- Terraform state, AWS 계정 topology, Fargate/ECS resource topology, raw
  secret 값을 커밋하지 않습니다.
- raw request token, IdP secret, customer data export를 소유하지 않습니다.
- provider runtime process를 실행하거나 provider CLI binary를 번들하지 않습니다.
- production persistence와 DNS 운영 evidence를 SSOT로 삼지 않습니다.

## 왜 이 작업을 여기서 했나

AI Agent client-facing endpoint는 Riido web과 desktop webview가 직접 호출하는
SaaS API입니다. 따라서 handler, auth scope gate, development store, SSE replay,
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
| AI Agent client API endpoint와 development/testnet 정책 | [`docs/20-domain/ai-agent-client-api.md`](docs/20-domain/ai-agent-client-api.md) |
| authorization resource/action/scope 규칙 | [`docs/20-domain/request-authorization.md`](docs/20-domain/request-authorization.md) |
| agent catalog RBAC 규칙 | [`docs/20-domain/agent-catalog-rbac.md`](docs/20-domain/agent-catalog-rbac.md) |
| runtime/agent binding domain | [`docs/20-domain/agent-runtime-binding.md`](docs/20-domain/agent-runtime-binding.md) |
| control-plane bounded context | [`docs/20-domain/context-map.md`](docs/20-domain/context-map.md) |
| riido-client generated React Query 전달 정책 | [`docs/30-architecture/api-client-delivery.md`](docs/30-architecture/api-client-delivery.md) |
| Figma AI Agent coverage의 control-plane generated-client 투영 | [`docs/30-architecture/figma-ai-agent-control-plane-projection.md`](docs/30-architecture/figma-ai-agent-control-plane-projection.md) |
| module/package 책임 분해 | [`docs/30-architecture/module-decomposition.md`](docs/30-architecture/module-decomposition.md) |
| runtime env 변수와 설정 책임 | [`docs/30-architecture/config-reference.md`](docs/30-architecture/config-reference.md) |
| daemon/contracts/infra/client와의 연결 | [`docs/30-architecture/integration-matrix.md`](docs/30-architecture/integration-matrix.md) |
| public runtime과 deploy boundary | [`docs/30-architecture/runtime-deployment-boundary.md`](docs/30-architecture/runtime-deployment-boundary.md) |
| runtime artifact CD 소유권과 CodeDeploy 전환 경계 | [`docs/30-architecture/runtime-cd-ownership.md`](docs/30-architecture/runtime-cd-ownership.md) |
| 마이그레이션 히스토리 | [`docs/migration/control-plane.md`](docs/migration/control-plane.md) |

## AI Agent development testnet

AI Agent client development API는 다음 env로 켭니다. development 서버는
DynamoDB snapshot item을 영속 저장소로 사용하므로, 프로세스 재시작 이후에도
agent/device credential/task thread 상태가 유지됩니다.

```bash
RIIDO_AI_SERVER_AI_AGENT_CLIENT_DEVELOPMENT=true
RIIDO_AI_SERVER_AI_AGENT_CLIENT_DYNAMODB_TABLE=<configured by riido-infra>
```

development API도 인증 없이 열리지 않습니다. request-token scope와
owner/public/private visibility policy를 통과해야 합니다. AWS topology, table
name 값, ECS task role, credential endpoint 값은 `riido-infra`와 GitHub
environment/secret 경계에 둡니다.

현재 testnet smoke는 별도 GitHub Actions workflow가 담당합니다.

- workflow: `ai-agent-client-testnet-smoke`
- deploy workflow: `deploy-ai-agent-testnet`

`deploy-ai-agent-testnet`은 `v*` tag push 또는 수동 dispatch에서만 실행합니다.
이 workflow는 GitHub OIDC로 deploy role을 assume하고, image tag를 Git ref와
commit SHA에서 만들며, ECR image digest를 ECS task definition revision에
명시합니다. `latest` tag를 배포 기준으로 쓰지 않습니다.

필요한 GitHub 설정은 이름만 공개 문서화하고 값은 secrets/variables에 둡니다.
live URL, AWS account id, ARN, image digest, task-definition revision, workflow
run URL, smoke 결과 payload는 이 public repo에 고정하지 않습니다. 수동
dispatch에서도 deploy/smoke workflow 모두 live URL은 입력값으로 받지 않고
environment variable만 사용하며, image URI/task-definition ARN 같은 live
중간값은 GitHub step output이 아니라 job 내부 `$RUNNER_TEMP` 파일로만
전달합니다.

CodeDeploy blue/green으로 전환하더라도 CD 실행 owner는
`riido-control-plane`입니다. CodeDeploy application/deployment group, target
group/listener topology, IAM, rollback policy는 `riido-infra`가 Terraform으로
소유하고, 이 public repo는 설정 이름과 동작만 문서화합니다. RIID-4822 이후
infra output으로 검증된 application/deployment group 이름만 workflow 환경 설정으로
전달할 수 있으며, service role ARN, target group/listener ARN, AppSpec/task
definition JSON, deployment id, smoke payload는 public workflow 입력이나 artifact로
노출하지 않습니다. optional CodeDeploy 설정이 없으면 기존 ECS rolling 경로를
유지합니다.

RIID-4839 기준으로 runtime CD ownership SSOT가 public CD configuration key의
최소 공개 집합입니다.
정확한 GitHub secret/variable 이름 목록은 README에서 반복하지 않고
[`docs/30-architecture/runtime-cd-ownership.md`](docs/30-architecture/runtime-cd-ownership.md)와
[`docs/30-architecture/runtime-cd-ownership.riido.json`](docs/30-architecture/runtime-cd-ownership.riido.json)에서
관리합니다. 새 `RIIDO_AI_SERVER_*` GitHub secret/variable 이름이 필요하면 workflow보다 먼저
`docs/30-architecture/runtime-cd-ownership.riido.json`의
`public_config_key_minimization`을 갱신해야 합니다. key 값, live 예시, generated
deploy payload, image/task-definition/CodeDeploy/smoke evidence는 공개 문서나
workflow output/artifact로 남기지 않습니다.

RIID-4842 기준으로 public configuration key 이름도 민감 표면 예산으로 봅니다.
기존 stable key 이름은 운영자가 GitHub environment를 설정하는 데 필요한 범위에서만
공개하고, 새 `RIIDO_AI_SERVER_*` key 이름은 manifest가 먼저 갱신되지 않으면
README/docs/workflow에 등장할 수 없습니다. `riido-infra`는 이 key category와
stable source name만 알고, 실제 값과 live payload는 계속 out-of-band로 관리합니다.

RIID-4835 public export contract에 따라 이 레포가 외부로 남겨도 되는 CD 정보는
stable key 이름, workflow 이름, git tag/commit, aggregate pass/fail 상태뿐입니다.
live URL, AWS 식별자, image 값, task definition/AppSpec JSON, deployment id,
smoke payload, Terraform evidence는 GitHub environment나 private infra/operator
evidence에만 둡니다.

RIID-4836은 위 경계를 공개 표면 스캔으로 고정합니다. README, CD 관련 문서,
deploy/smoke workflow, generated-client 안내에서 live host, AWS account literal,
checked-in ARN, live ALB/API Gateway/CloudFront URL, public workflow handoff
mechanism이 새로 들어오면 `tools/deploypolicy` 검증에서 실패해야 합니다.

RIID-4837은 이 스캔을 generated React wrapper와 generated-client delivery 문서까지
확장합니다. Private/operator evidence가 image digest나 workflow run reference를
다룰 수는 있지만, public workflow output/artifact, checked-in example, generated
client 안내로 넘기지 않습니다.

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
Riido work `branchName` delivery 정책을 따라야 합니다. 이 레포의 workflow 산출물만
신뢰하고 client repo에서 Orval을 직접 실행하지 않는 것이 원칙입니다.

## 검증

```bash
go test ./...
go list -m all
go test ./internal/riidoaiserver -run 'AIAgentClient' -count=1
go test ./cmd/riido_ai_server -run 'AIAgentClient|ConfigFromEnv' -count=1
go test ./tools/reactquerygen -count=1
go run ./tools/reactquerygen -openapi contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json -out web/generated/aiAgentClient.ts
go test ./tools/containercontract -count=1
go run ./tools/containercontract -contract packaging/containers/riido_ai_server_container.riido.json -out -
docker build -f packaging/containers/riido_ai_server.Dockerfile -t riido-control-plane:local .
```

CI는 public repo에서 가벼운 검증을 돌리기 위한 경계입니다. 배포 비용이 생기는
동작은 pull request에서 실행하지 않습니다. runtime artifact CD는 tag/manual
workflow가 소유하고, AWS resource topology와 Terraform drift 정책은
`riido-infra`가 소유합니다.

## Module

```text
github.com/teamswyg/riido-control-plane
```

## License

Apache-2.0.
