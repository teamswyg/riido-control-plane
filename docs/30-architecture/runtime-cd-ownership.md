# Runtime CD Ownership

> Riido task: RIID-4825 `[Control Plane/Infra] CD ownership remodel and public redaction SSOT`

This document explains the deploy ownership manifest in
[`runtime-cd-ownership.riido.json`](runtime-cd-ownership.riido.json). It does
not redefine AWS topology. Terraform topology and operator evidence remain in
`riido-infra`.

## Ownership

`riido-control-plane` owns runtime artifact CD for `riido_ai_server`.

That means this repository owns the workflow that:

- builds the checked-in container contract
- pushes an immutable image tag
- resolves the image digest
- promotes the runtime artifact into the configured deployment target
- waits until the deployment target is stable
- runs live AI Agent smoke checks

`riido-infra` owns the AWS topology that the workflow points at:

- ECR repository
- ECS cluster and service
- task definition bootstrap shape
- ALB, target groups, listeners, WAF, DNS, and ACM
- IAM roles and deploy permissions
- DynamoDB, EventBridge, secret references, Terraform backend, and evidence

## Current And Future Strategies

The current testnet default strategy is ECS rolling deployment by registering a
new task definition revision and updating the ECS service. That execution
remains in the `deploy-ai-agent-testnet` workflow.

The same workflow also has a topology-gated CodeDeploy mode. When
`RIIDO_AI_SERVER_CODEDEPLOY_APPLICATION` and
`RIIDO_AI_SERVER_CODEDEPLOY_DEPLOYMENT_GROUP` are both configured, the workflow
still builds and registers the task-definition revision itself, then creates
CodeDeploy AppSpec content in `$RUNNER_TEMP`, creates a CodeDeploy deployment,
waits for deployment success, and runs the same smoke checks after the
deployment succeeds.

The ownership does not flip in that mode: `riido-control-plane` owns create /
wait / smoke execution, while `riido-infra` owns the CodeDeploy application,
deployment group, blue/green target group/listener topology, rollback policy,
and IAM boundary through Terraform/operator evidence. After RIID-4822 the public
workflow consumes only the configured application and deployment-group names
from infra outputs; service role ARNs, target group/listener ARNs,
task-definition JSON, generated AppSpec JSON, deployment IDs, and smoke payloads
must not become public workflow inputs or uploaded artifacts.

## Public Redaction

Public repo docs and workflow files may contain only stable key names and
behavior. Required deploy key names should stay centralized in the workflow and
ownership/configuration docs, and public files should avoid environment-specific
examples for domains, clusters, services, applications, deployment groups, ARNs,
and URLs. The public workflow must not commit or upload:

- live URL values
- AWS account IDs
- ARNs
- task-definition revision values
- CodeDeploy deployment IDs
- image digests or image URIs
- workflow_dispatch input values carrying live URLs
- GitHub step outputs carrying live deployment values
- task definition JSON
- CodeDeploy AppSpec JSON
- smoke response payloads
- Terraform plans, state, tfvars, apply logs, or raw evidence

Those values may exist only in GitHub environment configuration, AWS, or
ignored operator evidence owned by `riido-infra`.

When one deploy step must hand a live value to the next step in the same job, it
uses `$RUNNER_TEMP` files with restrictive permissions and masks the value again
before use. That includes image URI, ECS task-definition ARN, generated
CodeDeploy AppSpec JSON, generated CodeDeploy request JSON, and CodeDeploy
deployment id. Live URL overrides are not accepted as manual workflow inputs;
the configured GitHub environment variable is the only smoke target source.
This rule applies to both the runtime deploy workflow and the companion
AI Agent client testnet smoke workflow.

RIID-4833 tightens the implementation side of that rule. Shell steps that write
live deployment handoff files, task-definition JSON, CodeDeploy JSON, or smoke
replay files set `umask 077` before writing them. Cross-step handoff files are
still explicitly `chmod 600` before reuse, and same-step files are deleted by
traps. The companion smoke workflow treats its SSE replay capture as a temporary
payload, not a public artifact or evidence file.

The deploy job removes long-lived handoff temp files in an `always()` cleanup
step, and CodeDeploy-only generated JSON/deployment-id files are removed by
same-step shell traps. That cleanup rule is part of the public redaction
contract: temp files are an implementation detail of one deploy job, not release
evidence, workflow output, or an artifact.

## Public Export Contract

RIID-4835 narrows the remodel into a public export contract. The public
`riido-control-plane` repository may export only stable, non-live information:
workflow file names, GitHub secret/variable key names, stable infra output key
names, control-plane git tag or commit identifiers, and aggregate pass/fail
status. Those values are enough for operators and `riido-infra` to understand
what must be configured without teaching the public repository live environment
details.

Image values are deliberately not in that public export set. The deploy job may
resolve an immutable image URI/digest and pass it between steps in same-job
runner temp files, but that value is runtime evidence, not a public handoff
artifact, reusable workflow output, uploaded artifact, or checked-in example.

`riido-infra` must know the contract because it owns topology and operator
evidence, but it must consume only stable output names, redaction categories,
out-of-band GitHub environment values, and private evidence summaries. It must
not receive public workflow payloads such as image URIs, image digests, ECS
task-definition ARNs, task-definition JSON, CodeDeploy AppSpec/request JSON,
deployment IDs, smoke payloads, Terraform plans, state, tfvars, apply logs, or
raw operator evidence.

The workflow must not use uploaded artifacts, `GITHUB_OUTPUT`, manual
`workflow_dispatch` URL inputs, or cross-workflow handoff for live deployment
values. Same-job `$RUNNER_TEMP` files with restrictive permissions remain the
only allowed bridge between deploy steps.

## Public Surface Scan

RIID-4836 turns the public export contract into a deterministic scan. The scan
does not try to hide that this repository has a deploy workflow. It checks that
the public surface only contains stable names and non-live behavior, and that it
does not accidentally pin live hosts, AWS account IDs, literal ARNs, ALB/API
Gateway/CloudFront URLs, or public handoff mechanisms for live deploy values.

The scan scope is intentionally small and explicit: README, the CD ownership and
deployment-boundary docs, the SaaS/client API docs, the migration log, the two
public deploy/smoke workflows, and generated-client guidance. Workflow-internal
AWS CLI field names such as `imageDigest` or `deploymentId` are allowed because
they describe the API shape, not a live value. Placeholder host examples that
use angle brackets are also allowed.

`riido-infra` must know that this scan exists because it owns topology and
operator evidence. It must not treat scan output as a deployment artifact,
release evidence, or a source of live handoff values. The only durable takeaway
for infra is the same ownership rule: CD execution and public redaction gates
stay in `riido-control-plane`; topology, IAM, drift, and private evidence stay in
`riido-infra`.

RIID-4837 closes the final guard around that remodel. The scan scope also covers
the generated-client delivery guide and generated React wrapper because those
files are public client-facing surfaces that can accidentally teach a live host
or handoff mechanism. The infra awareness path also lists every no-diff
hardening work unit from RIID-4833, RIID-4835, RIID-4836, and RIID-4837 so infra
knows the policy sequence without consuming workflow payloads.

Private/operator gates may use image digests or workflow run references only as
out-of-band evidence summaries. They are not public release hand-off artifacts,
workflow outputs, uploaded artifacts, checked-in examples, or reusable inputs for
`riido-infra`.

## Public Config Key Minimization

RIID-4839 makes the public configuration surface explicit. The public
`riido-control-plane` workflows and docs may name only the stable GitHub
configuration keys needed for runtime artifact CD and smoke:

- secrets: `RIIDO_AI_SERVER_DEPLOY_ROLE_ARN`,
  `RIIDO_AI_SERVER_TESTNET_TOKEN`
- required variables: `RIIDO_AI_SERVER_AWS_REGION`,
  `RIIDO_AI_SERVER_ECR_REPOSITORY`, `RIIDO_AI_SERVER_ECS_CLUSTER`,
  `RIIDO_AI_SERVER_ECS_SERVICE`, `RIIDO_AI_SERVER_ECS_CONTAINER_NAME`,
  `RIIDO_AI_SERVER_TESTNET_BASE_URL`
- optional variables: `RIIDO_AI_SERVER_TESTNET_WORKSPACE_ID`,
  `RIIDO_AI_SERVER_CODEDEPLOY_APPLICATION`,
  `RIIDO_AI_SERVER_CODEDEPLOY_DEPLOYMENT_GROUP`

Those names are enough for operators to wire GitHub environment configuration.
Their values, live examples, generated deploy payloads, and detailed evidence do
not belong in this public repository. Adding another
`RIIDO_AI_SERVER_*` GitHub key is a public surface change and must update
[`runtime-cd-ownership.riido.json`](runtime-cd-ownership.riido.json) before the
workflow uses it.

`riido-infra` may know the stable source names that map to those keys, such as
`ecr_repository_name`, `ecs_cluster_name`, `service_name`, `container_name`,
`codedeploy_application_name`, and `codedeploy_deployment_group_name`. It must
still not receive live workflow payloads, image values, task-definition values,
CodeDeploy generated JSON, deployment IDs, smoke payloads, Terraform plans,
state, tfvars, apply logs, or raw operator evidence through public outputs or
artifacts.

## Public Sensitive Surface Guard

RIID-4842 treats public CD configuration key names as a sensitivity budget, not
a glossary. The stable `RIIDO_AI_SERVER_*` names listed in the manifest may be
used where operators need to configure GitHub environments, but public docs and
workflows must not introduce another key name first and justify it later.

That ratchet matters because public repositories are allowed to describe the
deploy mechanism, while still minimizing anything that helps reconstruct a live
environment. New public key names, live examples, hostnames, AWS identifiers,
image/task-definition values, generated CodeDeploy payloads, deployment IDs,
smoke payloads, and raw operator evidence remain outside the public surface.
The current non-CD exceptions are `RIIDO_AI_SERVER_AI_AGENT_CLIENT_MOCK`, a
disposable testnet mock runtime flag, and `RIIDO_AI_SERVER_ADDR`, the
non-live container listen-address shape. Neither one is a deploy/smoke GitHub
configuration key.

`riido-infra` still needs to know this rule because it owns topology, IAM, drift,
and evidence. It consumes the stable key categories and source names, then wires
actual values out of band through GitHub environment configuration, AWS, or
ignored operator evidence. It must not ask public workflows to export live
deployment payloads as convenience handoffs.

## Drift Rule

Top-down changes start in this manifest or the runtime deployment boundary.
Bottom-up infra findings can ask for topology changes, but they do not move CD
execution into `riido-infra`. Bottom-up workflow findings can tighten public
redaction locally; if they require new AWS resources, they must create an infra
work unit first.

`riido-infra` must know this SSOT because it owns the target topology and
operator evidence, but it should only receive stable output names and evidence
categories from this public repo. It must not receive generated AppSpec/request
JSON, deployment IDs, image URIs/digests, task-definition JSON, or smoke
payloads through public workflow outputs or artifacts.
