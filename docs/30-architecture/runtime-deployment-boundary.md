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

`tools/containercontract` validates this shape.

## Testnet Runtime CD

Riido task RIID-4807 moves the **testnet runtime artifact CD execution** into
this public repository while keeping AWS topology and secret values outside the
repository.
RIID-4814 adds the executable ownership projection in
[`runtime-cd-ownership.riido.json`](runtime-cd-ownership.riido.json).

The `deploy-ai-agent-testnet` workflow is allowed to:

- run only from a `v*` tag push or explicit manual dispatch
- assume the deploy role through GitHub OIDC
- build the checked-in container contract image
- push an immutable ECR tag derived from the Git ref and commit SHA
- resolve the pushed image to an ECR image digest
- register a new ECS task-definition revision by replacing only the configured
  container image
- update the configured ECS service
- wait for ECS service stability
- smoke `healthz`, `readyz`, and the v2 workspace-scoped AI Agent bootstrap API

The workflow must not commit or print unmasked AWS account values, raw token
values, Terraform state, plan output, production secret payloads, task
definition JSON, task-definition ARNs, image digests, live workflow run URLs,
workflow_dispatch URL inputs, GitHub step outputs that carry live deployment
values, or smoke response payloads. Public repo configuration uses only GitHub
secrets/variables:

- secret: `RIIDO_AI_SERVER_DEPLOY_ROLE_ARN`
- secret: `RIIDO_AI_SERVER_TESTNET_TOKEN`
- variable: `RIIDO_AI_SERVER_AWS_REGION`
- variable: `RIIDO_AI_SERVER_ECR_REPOSITORY`
- variable: `RIIDO_AI_SERVER_ECS_CLUSTER`
- variable: `RIIDO_AI_SERVER_ECS_SERVICE`
- variable: `RIIDO_AI_SERVER_ECS_CONTAINER_NAME`
- variable: `RIIDO_AI_SERVER_TESTNET_BASE_URL`
- optional variable: `RIIDO_AI_SERVER_TESTNET_WORKSPACE_ID`
- optional variable pair: `RIIDO_AI_SERVER_CODEDEPLOY_APPLICATION`,
  `RIIDO_AI_SERVER_CODEDEPLOY_DEPLOYMENT_GROUP`

The names above are a workflow contract. Their values, the current live URL, and
any deployment evidence stay in GitHub environment configuration or
`riido-infra`/operator evidence, not in checked-in public docs. Live values
needed only inside the job are passed through `$RUNNER_TEMP` files with
restrictive permissions and re-masked before use, not through reusable GitHub
step outputs. The workflow must not upload deployment artifacts from the live
run.

`riido-infra` still owns the Terraform module that creates ECR, ECS, ALB,
security groups, IAM boundaries, DynamoDB, EventBridge, DNS/ACM/WAF, and the
policy that Terraform should not roll back the ECS service task definition after
CD promotes a new image digest.

## CodeDeploy Handoff

CodeDeploy blue/green is a topology-gated production hardening strategy, not the
default testnet mechanism. The public workflow is allowed to use CodeDeploy only
when the optional CodeDeploy application and deployment group names are
configured together. In that mode, runtime artifact CD execution still belongs
to `riido-control-plane`: this repository owns creating the deployment from an
immutable image digest and task-definition revision, waiting for the deployment
to finish, and running smoke after traffic shift.

`riido-infra` owns the CodeDeploy application/deployment group, blue/green
target group and listener topology, CodeDeploy IAM role, rollback policy,
Terraform drift handling, and operator evidence. After RIID-4822 the public
workflow may consume only the configured application and deployment-group names
from GitHub environment variables populated from infra outputs; it must not
consume service role ARNs, target group/listener ARNs, task definition JSON,
AppSpec JSON, deployment IDs, image digests, or smoke payloads as reusable
inputs or artifacts. Generated CodeDeploy AppSpec/request JSON and the
deployment id stay in same-job `$RUNNER_TEMP` files and are removed by the job.

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
- image digest produced by the tag-triggered testnet CD workflow
- ECS service stability and AI Agent testnet smoke result

PR descriptions and chat messages are not release SSOT. Any durable decision must
land in a domain, architecture, ADR, migration, or infra evidence document.
