# Control Plane Context Map

> Riido task: RIID-4712 `[Control Plane] Architecture SSOT docs migration`

This file is the public context map for `github.com/teamswyg/riido-control-plane`
after the split from the former private monolith.

## Owned Contexts

| Context | Public owner | Responsibility |
| --- | --- | --- |
| C10 SaaS Control Plane | `internal/riidoaiserver` | assignment creation, polling, heartbeat, event append, SSE, metrics read model, health/ready routes, review seed provisioning, provider status sync/read, agent catalog RBAC, AI Agent client development API, request authorization ports, browser frontend CORS transport |
| C10 Runtime Adapter | `cmd/riido_ai_server` | environment parsing, HTTP server startup/shutdown, optional stdout metrics publisher startup |
| C10 Public AWS Adapter Facade | `awsadapters` | module-consumable aliases for stdlib-only DynamoDB/EventBridge adapter DTOs, constructors, and ports needed by private infra evidence tooling |
| C10 Container Contract | `packaging/containers` and `tools/containercontract` | public executable image shape for the control-plane binary |

`riido-control-plane` is allowed to own stdlib-only AWS request construction and
fake-endpoint verified adapter behavior. It does not own AWS resource topology,
Terraform state, live apply evidence, production secrets, or account-specific
configuration.

## Imported Contexts

| Context | Imported from | Use |
| --- | --- | --- |
| Assignment Polling Contract | `github.com/teamswyg/riido-contracts/assignment` | shared assignment state, poll action, heartbeat/event DTOs, and agent binding DTOs |
| Provider / Distribution Vocabulary | `github.com/teamswyg/riido-contracts/provider/capability` | provider status and routing vocabulary shared with the daemon |
| IR / Task Vocabulary | `github.com/teamswyg/riido-contracts/ir` and `task` packages when needed by compatibility gates | cross-repo schema compatibility checks |

Imported contracts are tagged Go module dependencies. Control-plane code must not
import daemon internals, private infra modules, or former monolith package paths.

## External Contexts

| Context | Owner | Boundary |
| --- | --- | --- |
| Customer-PC daemon runtime | `riido-daemon` | provider process execution, local task DB, Unix socket API, customer host integration |
| Shared contract tags | `riido-contracts` | cross-repo DTOs, state vocabulary, schema-versioned fixtures |
| Infrastructure / deployment | `riido-infra` | Terraform modules, remote state, AWS account wiring, ECR push, ECS/Fargate deploy, DNS/ACM/WAF, deployment evidence |
| Store/app clients | `riido-client` web and `riido-desktop` webview | user-facing UI over public HTTP contracts, generated AI Agent client artifacts, and configured browser origins. Clients own screen composition and route wiring; control-plane owns HTTP/SSE behavior, OpenAPI projection, and generated-client delivery boundaries. |

## Direction Rules

Allowed imports:

- `cmd/riido_ai_server` -> `internal/riidoaiserver`
- `awsadapters` -> `internal/riidoaiserver`
- `internal/riidoaiserver` -> `github.com/teamswyg/riido-contracts/...`
- tests/tools -> public package surfaces under this module

Forbidden imports:

- any `github.com/teamswyg/riido-daemon` internal package
- any former private monolith package path
- AWS SDK packages without a new ADR
- Terraform, credential, or deployment-evidence packages

## SSOT Links

- SaaS behavior: [`saas-control-plane.md`](saas-control-plane.md)
- AI Agent client API: [`ai-agent-client-api.md`](ai-agent-client-api.md)
- Agent catalog RBAC: [`agent-catalog-rbac.md`](agent-catalog-rbac.md)
- Request authorization: [`request-authorization.md`](request-authorization.md)
- Provider status: [`provider-status.md`](provider-status.md)
- Assignment operation journal: [`assignment-operation-journal.md`](assignment-operation-journal.md)
- Architecture decomposition: [`../30-architecture/module-decomposition.md`](../30-architecture/module-decomposition.md)
- Config catalog: [`../30-architecture/config-reference.md`](../30-architecture/config-reference.md)
- Open questions: [`../50-roadmap/open-questions.md`](../50-roadmap/open-questions.md)
