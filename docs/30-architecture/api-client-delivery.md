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
- emit client-facing comments and notes in Korean, except for stable code
  identifiers, endpoint paths, enum literal values, package names, and repository
  names
- carry OpenAPI schema descriptions and operation summaries into generated
  JSDoc so frontend developers do not have to infer API behavior from type names
- generate a config-bound API facade so frontend developers can consume the
  surface as one module instead of stitching together isolated functions
- derive the facade module and namespace from DSL/IR client metadata projected
  into OpenAPI `x-riido-client`, not from generator-local operation-id switches

`tools/reactquerygen` remains a small deterministic public fixture generator for
the checked-in mock surface. It is useful for drift tests, but it is not the
cross-repository Orval delivery mechanism.

## Generated Artifacts

The client branch should receive:

- React Query hooks and DTO types generated from the OpenAPI projection
- `apiHistory.generated.ts`
- `contractManifest.generated.ts`
- `README.generated.md`

## Generated Client Facade

The generated client must keep primitive exports and a facade together:

- primitive request functions, query keys, query options, mutation options, and
  hooks remain individually exported for tests, tree-shaking, and direct use
- `createRiidoControlPlaneClient(config)` groups the same primitives by DSL
  client metadata, for example `aiAgent.agents.editability`,
  `aiAgent.tasks.stop`, and `aiAgent.devices.runtimes`
- the facade is config-bound, so consumers do not repeatedly pass `baseUrl`,
  `token`, and optional `fetcher`
- each operation exposes concise library-style aliases: `query(...)` mirrors
  `queryOptions(...)`, and `mutation(...)` mirrors `mutationOptions(...)`
- the facade does not create or replace TanStack `QueryClient`
- the facade does not own token refresh, global error toasts, retry policy, cache
  invalidation, or app-specific optimistic updates

The intended frontend usage is:

```ts
const riido = createRiidoControlPlaneClient(config);

useQuery(riido.aiAgent.bootstrap.query());
useMutation(riido.aiAgent.tasks.stop.mutation());
queryClient.prefetchQuery(riido.aiAgent.devices.runtimes.query());
```

This keeps the generated output feeling like a lightweight module while leaving
application integration policy inside `riido-client`.

The history artifact is intentionally lightweight. It should expose release
entries, operation-level lifecycle changes, replacements, removals, and notable
breaking or deprecation notes without making frontend developers read the full
OpenAPI diff.

The release manifest is the authoritative machine-readable summary for a
control-plane API release. It links the control-plane tag, source DSL/IR digest,
OpenAPI digest, generator version, generated output path, and lifecycle summary.

Client-facing README, history notes, manifest notes, and generated JSDoc are part
of the delivered developer experience. They must be written in Korean because
`riido-client` UI and team-facing development notes are Korean-first. English is
kept only where changing it would change code identity or wire contract shape.

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
- Client-facing generated comments and notes are Korean-first and preserve
  OpenAPI descriptions in JSDoc.
- Generated clients include `queryOptions`, `mutationOptions`, and the
  config-bound `createRiidoControlPlaneClient` facade without taking ownership of
  app-level query policies.
- Generated facade paths are sourced from DSL/IR/OpenAPI `client` metadata and
  fail generation if that metadata is missing or inconsistent.
- This slice does not edit `teamswyg/riido-client`.
