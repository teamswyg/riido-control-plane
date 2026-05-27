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

```bash
go test ./...
go list -m all
go test ./tools/containercontract -count=1
go run ./tools/containercontract -contract packaging/containers/riido_ai_server_container.riido.json -out -
docker build -f packaging/containers/riido_ai_server.Dockerfile -t riido-control-plane:local .
```

The public GitHub Actions workflow runs the lightweight verification suite
outside the private infrastructure repository billing pool. CI allows only
Riido-owned Go module dependencies; any third-party dependency requires a new
documented decision before it is introduced.

## License

Apache-2.0.
