# Runtime CD Ownership

> Riido task: RIID-4814 `[Control Plane/Infra] CodeDeploy 전환 CD 소유권과 public redaction SSOT`

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

The current testnet strategy is ECS rolling deployment by registering a new task
definition revision and updating the ECS service. That execution remains in the
`deploy-ai-agent-testnet` workflow.

If production moves to CodeDeploy blue/green, the ownership does not flip:
`riido-control-plane` still owns the deployment workflow and smoke checks.
`riido-infra` must first create or expose the CodeDeploy application,
deployment group, blue/green target group/listener topology, rollback policy,
and IAM boundary through Terraform/operator evidence. The public workflow may
only consume the configured names or ARNs through GitHub environment
secrets/variables.

## Public Redaction

Public repo docs and workflow files may contain key names and behavior, not live
deployment values. The public workflow must not commit or upload:

- live URL values
- AWS account IDs
- ARNs
- task-definition revision values
- CodeDeploy deployment IDs
- image digests or image URIs
- task definition JSON
- CodeDeploy AppSpec JSON
- smoke response payloads
- Terraform plans, state, tfvars, apply logs, or raw evidence

Those values may exist only in GitHub environment configuration, AWS, or
ignored operator evidence owned by `riido-infra`.

## Drift Rule

Top-down changes start in this manifest or the runtime deployment boundary.
Bottom-up infra findings can ask for topology changes, but they do not move CD
execution into `riido-infra`. Bottom-up workflow findings can tighten public
redaction locally; if they require new AWS resources, they must create an infra
work unit first.
