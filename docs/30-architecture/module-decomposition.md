# Control Plane Module Decomposition

> Riido task: RIID-4712 `[Control Plane] Architecture SSOT docs migration`

`github.com/teamswyg/riido-control-plane` is the public SaaS backend module. It
keeps domain decisions and adapter boundaries in a small number of packages so
public CI can verify the backend without AWS credentials.

## Packages

| Package/path | Role | Must not own |
| --- | --- | --- |
| `cmd/riido_ai_server` | binary entrypoint, Factor 12 env parsing, HTTP server lifecycle, optional stdout EMF publisher startup | domain decisions, Terraform, production secret values, provider process execution |
| `internal/riidoaiserver` | C10 domain and adapter implementation: assignment store actor, HTTP/SSE/metrics/health routes, AI Agent client development routes, request authorization, RBAC, provider status, review seed provisioning, file snapshot/outbox, stdlib-only DynamoDB/EventBridge request adapters | daemon runtime, external AWS SDK, live deploy evidence, cloud resource topology |
| `awsadapters` | public facade for private infra/evidence tools that need the public AWS adapter surface | new adapter behavior, duplicated DTOs, credential storage |
| `tools/containercontract` | executable verifier for `riido-container-image-contract.v1` | image publishing, ECR credentials, ECS task definition deployment |
| `tools/deploypolicy` | public workflow/docs redaction and CD ownership drift tests | live AWS deploy execution or private operator evidence |
| `tools/reactquerygen` | deterministic OpenAPI-to-React-Query fixture generator for the AI Agent client development surface | frontend app implementation, cross-repository client delivery, Orval runtime ownership |
| `internal/contractscompat` | dependency compatibility smoke tests for shared public contracts | domain redefinition |
| `internal/repoidentity` | repository identity guard | runtime behavior |

## Hexagonal Shape

The center of the module is `internal/riidoaiserver`. It exposes narrow ports for
authorization, assignment storage, metrics reading, event sinks, snapshots,
operation journals, and stream relay publishing. Adapters call these ports:

- HTTP handlers adapt `net/http` requests into domain calls.
- `cmd/riido_ai_server` adapts environment variables into runtime config.
- file snapshot/outbox adapters persist JSON locally for deterministic tests.
- DynamoDB/EventBridge adapters construct and sign AWS HTTP requests using only
  the standard library.
- `awsadapters` re-exports selected public surfaces for external private infra
  tooling.

The module deliberately keeps public behavior black-box testable with
`httptest`, fake AWS endpoints, fake credentials, and local JSON fixtures.

Generated React Query delivery to `riido-client` is an architecture boundary,
not a package responsibility. The delivery rules, tag trigger, target branch
shape, allowlisted output path, and Orval supply-chain boundary are owned by
[`api-client-delivery.md`](api-client-delivery.md).

## Dependency Rules

Allowed:

- standard library
- `github.com/teamswyg/riido-contracts` tagged releases
- package-local test helpers

Forbidden without a new ADR:

- unapproved direct Go module dependencies such as ORMs, web frameworks,
  dependency-injection frameworks, or config frameworks
- direct imports from `riido-daemon`
- private infra packages
- former monolith package paths
- generated credential, tfstate, tfvars, or release evidence files

Direct Go module dependencies must be declared in
[`../../dependency_allowlist.riido.json`](../../dependency_allowlist.riido.json)
and pass `go run ./tools/dependencyallowlist -contract
dependency_allowlist.riido.json`. Transitive dependencies are accepted through
the approved direct dependency graph and must not become direct dependencies
without a new allowlist entry.

## Runtime Boundaries

`cmd/riido_ai_server` is a thin runtime shell. It may parse environment
variables, start the HTTP server, wire review provisioning, and start optional
stdout metrics publication. It must not make deployment decisions. Container
shape is validated by `tools/containercontract`; runtime artifact CD execution
is owned by this repository's tag/manual workflow, while AWS topology and
operator evidence stay in `riido-infra`.
