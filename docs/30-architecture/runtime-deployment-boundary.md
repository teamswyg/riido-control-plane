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
RIID-4825 keeps that ownership in `riido-control-plane` while tightening the
public redaction and same-job handoff cleanup policy.
RIID-4835 narrows public export to stable key names, non-live behavior
descriptions, git identifiers, and aggregate status only.

The `deploy-ai-agent-testnet` workflow is allowed to:

- run only from a `v*` tag push or explicit manual dispatch
- use the testnet GitHub environment for tag pushes
- use an explicit manual-dispatch target to select only the configured testnet
  or development GitHub environment
- assume the deploy role through GitHub OIDC
- build the checked-in container contract image
- push an immutable ECR tag derived from the Git ref, commit SHA, and workflow
  run attempt
- resolve the pushed image to an ECR image digest
- register a new ECS task-definition revision by replacing only the configured
  container image
- update the configured ECS service
- wait for ECS service stability
- smoke `healthz`, `readyz`, and the v2 workspace-scoped AI Agent bootstrap API

The workflow must not accept live URLs, AWS identifiers, task-definition values,
or smoke tokens as manual-dispatch inputs. Environment selection only chooses
which preconfigured GitHub environment supplies the same stable variable/secret
names. The workflow must not commit or print unmasked AWS account values, raw token
values, Terraform state, plan output, production secret payloads, task
definition JSON, task-definition ARNs, image digests, live workflow run URLs,
workflow_dispatch URL inputs, GitHub step outputs that carry live deployment
values, or smoke response payloads. Public repo configuration uses only the
stable GitHub secret/variable categories recorded in
[`runtime-cd-ownership.riido.json`](runtime-cd-ownership.riido.json). RIID-4845
keeps the exact deploy/smoke key-name list in that machine-readable manifest and
in the workflow files that consume those keys, not repeated in human-readable
public docs.

Those keys are a workflow contract. Their values, the current live URL, and any
deployment evidence stay in GitHub environment configuration or
`riido-infra`/operator evidence, not in checked-in public docs. Live values
needed only inside the job are passed through `$RUNNER_TEMP` files with
restrictive permissions and re-masked before use, not through reusable GitHub
step outputs. The workflow must not upload deployment artifacts from the live
run. The workflow also removes image URI, task-definition ARN, and container
port temp files in an `always()` cleanup step so those values remain live deploy
implementation details, not public release evidence.
The public export contract intentionally avoids publishing live URLs, AWS
identifiers, image values, task-definition JSON/ARNs, CodeDeploy JSON or
deployment IDs, smoke payloads, Terraform plans, state, tfvars, apply logs, and
raw operator evidence.

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
immutable same-job image value and task-definition revision, waiting for the
deployment to finish, and running smoke after traffic shift.

`riido-infra` owns the CodeDeploy application/deployment group, blue/green
target group and listener topology, CodeDeploy IAM role, rollback policy,
Terraform drift handling, and operator evidence. After RIID-4822 the public
workflow may consume only the configured application and deployment-group names
from GitHub environment variables populated from infra outputs; it must not
consume service role ARNs, target group/listener ARNs, task definition JSON,
AppSpec JSON, deployment IDs, image digests, or smoke payloads as reusable
inputs or artifacts. Generated CodeDeploy AppSpec/request JSON and the
deployment id stay in same-job `$RUNNER_TEMP` files, are masked before use, and
are removed by same-step traps.

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
- aggregate deploy/smoke pass-fail status without live payload values

The image URI/digest produced by the tag-triggered testnet CD workflow, the ECS
task-definition revision/ARN, detailed ECS stability evidence, CodeDeploy
deployment id, and smoke payloads are not public hand-off artifacts. They remain
same-job temp values or private/operator evidence when an operator needs them.

RIID-4836 keeps that hand-off boundary executable by scanning the public CD
surface for live host literals, AWS account literals, checked-in ARN literals,
live ALB/API Gateway/CloudFront URL literals, and public workflow handoff
mechanisms. This scan is a public redaction gate, not a release artifact.

RIID-4837 extends the same scan to generated-client delivery docs and the
generated React wrapper. Those are not deploy workflows, but they are public
surfaces that frontend developers read and copy from. They may explain stable
configuration keys and generated call chains, but they must not pin live hosts or
teach public workflow handoff payloads.

RIID-4839 narrows the public configuration surface itself. The public deploy and
smoke workflows may reference only the stable `RIIDO_AI_SERVER_*` GitHub secret
and variable names listed in
[`runtime-cd-ownership.riido.json`](runtime-cd-ownership.riido.json). Those key
names are public; their values, live examples, generated deploy payloads, image
values, task-definition values, CodeDeploy generated JSON, deployment IDs, smoke
payloads, and detailed evidence are not public hand-off artifacts.

RIID-4842 treats even those public key names as a managed sensitivity budget.
The existing names can stay public because operators need them, but a new
`RIIDO_AI_SERVER_*` name must be added to the ownership manifest before it
appears in README, docs, or workflow files. `riido-infra` consumes the stable
categories and source names only; it does not receive live payloads through
public outputs or artifacts.

RIID-4845 narrows the human-readable public docs one more step. Exact deploy and
smoke key-name lists stay in the machine-readable manifest and workflow files;
docs like this boundary describe key categories, ownership, and the manifest
location without repeating the list.

RIID-4853 keeps the same CD ownership remodel and narrows the disclosure posture
again. Public control-plane docs and workflows may describe only the smallest
useful set of non-live CD facts needed for workflow wiring, review, and operator
setup. `riido-infra` knows that policy because it owns topology and evidence,
but infra still must not receive convenience public handoffs for image values,
task-definition values, CodeDeploy generated payloads, deployment IDs, smoke
payloads, Terraform plans/state/tfvars/apply logs, or raw evidence.

PR descriptions and chat messages are not release SSOT. Any durable decision must
land in a domain, architecture, ADR, migration, or infra evidence document.
