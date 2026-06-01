# Runtime CD Ownership

> Riido task: RIID-4822 `[Infra/Control Plane] CodeDeploy topology ownership and public redaction`

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
behavior. They should avoid environment-specific examples for domains, clusters,
services, applications, deployment groups, ARNs, and URLs. The public workflow
must not commit or upload:

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

## Drift Rule

Top-down changes start in this manifest or the runtime deployment boundary.
Bottom-up infra findings can ask for topology changes, but they do not move CD
execution into `riido-infra`. Bottom-up workflow findings can tighten public
redaction locally; if they require new AWS resources, they must create an infra
work unit first.
