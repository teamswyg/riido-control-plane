# Runtime Deployment Boundary

> Riido task: RIID-4712 `[Control Plane] Architecture SSOT docs migration`

This file separates what the public control-plane module can prove from what the
private infrastructure repository must prove.

## Public Runtime

`cmd/riido_ai_server` owns the buildable runtime process:

- parses documented environment variables
- starts `net/http` routes from `internal/riidoaiserver`
- shuts down on SIGINT/SIGTERM
- optionally provisions the public-safe review account seed using only a token
  hash
- optionally writes CloudWatch EMF-compatible JSON Lines to stdout

The process does not create AWS resources, fetch production secrets, mutate
Terraform state, push container images, or deploy itself.

## Container Artifact

The public image contract requires:

- static Go build of `./cmd/riido_ai_server`
- `CGO_ENABLED=0`
- `scratch` final image
- copied CA certificates
- non-root `65532:65532`
- `RIIDO_AI_SERVER_ADDR=:8080`
- `/riido_ai_server` entrypoint

`tools/containercontract` validates this shape. ECR upload, digest promotion,
and ECS/Fargate rollout are private infra responsibilities.

## AWS Adapter Boundary

`internal/riidoaiserver` owns stdlib-only DynamoDB/EventBridge request
construction and fake-endpoint verified behavior. `awsadapters` exposes a narrow
facade so private infra/evidence tooling can reuse that behavior without
importing an `internal` package.

This repository does not own:

- AWS account topology
- IAM policy attachment
- DynamoDB table creation
- EventBridge bus/rule creation
- ECS task/service wiring
- live traffic evidence
- SDK credential discovery policy beyond testable adapter inputs

Those facts live in `riido-infra`.

## Release Hand-off

The public hand-off artifacts are:

- Git commit / tag for `github.com/teamswyg/riido-control-plane`
- Go module version when tagged
- container image contract result
- optional image digest produced by private infra release tooling

PR descriptions and chat messages are not release SSOT. Any durable decision must
land in a domain, architecture, ADR, migration, or infra evidence document.
