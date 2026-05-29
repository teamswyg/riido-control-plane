# API Client Delivery

> Riido task: RIID-4746 `[Control Plane] AI Agent client API delivery SSOT 정리`

This file owns the control-plane decision for delivering the AI Agent client API
as generated React Query code to `riido-client`.

## Purpose

`riido-control-plane` must make the AI Agent client API feel like a lightweight
frontend module while keeping API shape, lifecycle metadata, and generated code
traceable to one source of truth.

The target client repository is not edited by this slice. This document defines
the future delivery boundary and the checks that must exist before a workflow is
allowed to open a generated-client PR.

## Ownership

| Owner | Owns | Does not own |
| --- | --- | --- |
| `riido-contracts` | canonical Riido vocabulary, domain policy grammar, shared enum and sum-type semantics, lifecycle/deprecation grammar | control-plane handler behavior, client repository branches |
| `riido-control-plane` | AI Agent client API sub-DSL, HTTP/SSE implementation, OpenAPI projection handoff, generated-client delivery workflow, release manifest | canonical business authority that changes non-API domain meaning |
| `riido-client` | consuming generated React Query code and reporting API usability feedback | OpenAPI SSOT, Orval execution, generated-code edits |

The AI Agent client API sub-DSL is not a random generated artifact. It is the
control-plane-owned API surface language that imports canonical terms from
`riido-contracts`. Client-facing API change requests enter through
`riido-control-plane`. If a request changes business meaning, narrows or expands
domain policy, or exposes a contradiction in the canonical language, it must be
escalated to `riido-contracts`.

The same escalation path applies in the other direction: backend implementation,
cost, or operations findings may ask `riido-contracts` to change or remove a
canonical concept. That does not immediately delete the lower API sub-DSL. The
API surface moves through lifecycle states so existing clients can migrate.

Agent setting fields follow that dependency direction. `profile_thumbnail_url`,
`description`, `instruction`, and server-authored `updated_at` are canonical
terms from `riido-contracts`; this repository delivers them as generated client
API shape and history/manifest metadata. If client usage suggests a different
thumbnail storage model, description limit, instruction semantics, timestamp
meaning, or model catalog, the delivery workflow records the local finding but
does not alter the canonical meaning without a contracts change.

Figma menu placement (`node-id=156-19307`) is a generated-client consumption
context, not a generated-client endpoint. The delivery workflow may document
which generated calls a route uses after navigation, but client menu labels,
ordering, selected state, and route wiring stay outside the control-plane API
projection.

Figma task-thread annotations (`node-id=153-15931`) are also consumption context.
They may name generated call chains such as `riido.aiAgent.events.stream` and
`riido.aiAgent.tasks.stop`, but the delivery artifact must derive those chains
from the control-plane OpenAPI projection and generated-client manifest. It must
not hand-code annotation strings as a second source of truth. Task-thread
screens first consume the generated cold collection call for
`GET /v1/client/ai-agent/tasks/{task_id}/threads`; the returned
`active_stream` link, when present, is the generated handoff into the SSE client
event stream.

Figma participant dropdown annotations (`node-id=153-12742`) are generated-client
consumption context for `listAIAgentTaskAssignableAgents`. The delivery artifact
may document that this query feeds the AI Agent rows in the participant dropdown,
but member sorting, long-name truncation, max-height, scrollbar width, and
checkbox layout remain client implementation details.

Figma runtime settings annotations (`node-id=162-23090`) are generated-client
consumption context for `listAIAgentDeviceRuntimes` and
`device_runtime_snapshot`. The delivery artifact may document that these
generated pieces feed the runtime settings route and attached-agent avatar rows.
It must not turn the agent hover popover, daemon stop modal, or restart
animation into a control-plane endpoint. Desktop-local daemon lifecycle
composition remains outside the generated SaaS client.

Figma agent setting annotations (`node-id=164-50215`) and add-screen evidence
(`node-id=134-6542`) are generated-client consumption context for
bootstrap/create/update/editability APIs. The delivery artifact may document
that `createAIAgent` feeds the add-agent save flow and `updated_at` feeds list
update dates and absolute-time tooltips. It must not turn row click, meatball
edit entry, save-button enablement, long-description presentation, dropdown
rendering, or provider-specific model labels into generated API facts before
the contracts model-catalog question is resolved.

Figma onboarding annotations (`node-id=42-3014`) are generated-client
consumption context for bootstrap, devices, and create APIs. The delivery
artifact may document that `agent_templates` feeds the starter-agent selection
screen, but workspace selection, row selection, direct-setting expansion,
scroll, two-line ellipsis, and no-installed-AI step skipping remain client
composition over generated data. It must not make frontend hard-coded template
copy a second SSOT.

Figma web onboarding annotations (`node-id=236-29749`) are not generated-client
endpoint evidence for this API. The delivery artifact may mention that AI Agent
screens are reached after auth/onboarding, but sign-up/login, terms consent,
email validation, member invite, app download, Windows launch notification,
marketing consent, and animation references stay in auth/team/product/client
documentation until a separate owning SSOT adds a generated operation. The AI
Agent generated client must not ship placeholder waitlist or marketing helpers
from this screen alone.

## DSL Lifecycle

The sub-DSL must preserve lifecycle metadata before OpenAPI generation:

| State | Meaning |
| --- | --- |
| `active` | current default API surface |
| `deprecated` | still generated, but JSDoc and OpenAPI mark replacement guidance |
| `superseded` | replaced by a newer operation/schema, still emitted for migration |
| `frozen` | no new features; bug/security fixes only |
| `removed_from_active_projection` | omitted from default generated client after the removal horizon |
| `archived` | retained only in historical manifest and changelog artifacts |

Lifecycle metadata must be lossless from DSL to IR. OpenAPI may expose the
standard `deprecated` flag, but Riido-specific metadata must remain in extension
fields such as `x-riido-lifecycle`, `x-riido-replacement`, and
`x-riido-removal-horizon`.

## Release Trigger

Generated-client delivery is allowed only from a `riido-control-plane` Git tag
that represents an API release, for example `v1.20.2`.

Regular pushes to `main` may validate local drift, but they must not push a
branch to `riido-client`. This keeps frequent server development from creating
client churn and CI cost.

## Target Branch

The future workflow creates a branch in `teamswyg/riido-client` from its current
`main`:

```text
react-query-{tag}-{shortsha}
```

Examples:

```text
react-query-v1.20.2-a1b2c3d
```

The branch must contain generated files only under this allowlist:

```text
src/generated/react-query/riido-control-plane/**
```

After generation, the workflow must fail if `git diff --name-only` contains any
path outside that allowlist. This prevents Orval config drift, package metadata
drift, or unrelated client edits from riding along with generated API delivery.

## Generator Boundary

Only the control-plane workflow output is trusted. `riido-client` consumes the
generated files but does not run Orval as part of normal development.

Generator requirements:

- pin the Orval package version in the control-plane generator workspace
- commit the lockfile for that generator workspace
- run with frozen lockfile semantics
- avoid `npx` or floating package installs
- disable or audit install scripts where the package manager supports it
- keep the generated output deterministic enough for reviewable diffs

`tools/reactquerygen` remains a small deterministic public fixture generator for
the checked-in mock surface. It is useful for drift tests, but it is not the
cross-repository Orval delivery mechanism.

## Generated Artifacts

The client branch should receive:

- React Query hooks and DTO types generated from the OpenAPI projection
- `apiHistory.generated.ts`
- `contractManifest.generated.ts`
- `README.generated.md`

The history artifact is intentionally lightweight. It should expose release
entries, operation-level lifecycle changes, replacements, removals, and notable
breaking or deprecation notes without making frontend developers read the full
OpenAPI diff.

The release manifest is the authoritative machine-readable summary for a
control-plane API release. It links the control-plane tag, source DSL/IR digest,
OpenAPI digest, generator version, generated output path, and lifecycle summary.

## Secret And Permission Boundary

Prefer a GitHub App installed only where needed. Required secrets live in
`riido-control-plane`, not in `riido-client`:

- `RIIDO_CLIENT_DELIVERY_APP_ID`
- `RIIDO_CLIENT_DELIVERY_PRIVATE_KEY`

If a fine-grained token is used temporarily, it must be scoped to the target
repository and limited to contents and pull-request write permissions.

The workflow must not require npm publish tokens, cloud credentials, Terraform
state, customer data, or production bearer tokens.

## Acceptance Criteria

- `riido-control-plane` docs name the API sub-DSL owner and canonical contract
  escalation path.
- Release delivery is tag-triggered, not main-push-triggered.
- Target `riido-client` branch naming and generated path allowlist are defined.
- Orval is a pinned control-plane generator dependency when the workflow is
  implemented.
- Generated history and manifest artifacts carry lifecycle/deprecation
  information from DSL/IR through the client handoff.
- This slice does not edit `teamswyg/riido-client`.
