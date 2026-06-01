# Figma AI Agent Control-Plane Projection

> Riido task: RIID-4810 `[Control Plane] Figma coverage projection generated-client gate`

This document is the control-plane projection of the upstream Figma coverage
SSOT owned by `riido-contracts`:

```text
riido-contracts/docs/30-architecture/figma-ai-agent-coverage.riido.json
```

`riido-control-plane` does not redefine the Figma top-level UI coverage. This
repo only checks the parts of that coverage that require HTTP/SSE, OpenAPI, and
generated-client behavior.

The executable local manifest is
[`figma-ai-agent-control-plane-projection.riido.json`](figma-ai-agent-control-plane-projection.riido.json).
The mirrored upstream coverage manifest is
[`../../contracts/ai-agent-client/figma-ai-agent-coverage.riido.json`](../../contracts/ai-agent-client/figma-ai-agent-coverage.riido.json).
It is copied from `riido-contracts` only so this repository can prove its
projection does not ask for generated paths that upstream coverage did not
name. The mirror also carries the upstream page registry and non-UI top-level
wireframe evidence so this repository can prove it is consuming the whole-file
coverage map, not only the primary `UI` page.

The mirrored page registry keeps the contracts-owned inspection method:
`figma.root.children` from the Figma Plugin API is the page registry authority,
but non-current page child counts are authoritative only after
`await figma.setCurrentPageAsync(page); page.children.length`. Passive page
objects can be lazy/unloaded, so metadata XML/read output and unloaded
`page.children.length` reads remain supporting evidence only and must not
redefine page-level counts in the control-plane mirror.

The mirror also preserves contracts-owned `client_delivery_annotations`. Figma
Dev Mode category `39:0` / `클라이언트 전달` may show frontend facade examples
such as `riido.aiAgent.events.stream` or `riido.aiAgent.tasks.stop`. This repo
does not treat the leading `riido.` variable name as a contracts generated path;
the projection gate normalizes those examples to canonical generated paths such
as `aiAgent.events.stream` and `aiAgent.tasks.stop`, then verifies that both the
canonical path and the Korean generated-client access example appear in
`web/generated/aiAgentClient.ts` and `web/generated/aiAgentClient.react.ts`.
The stale Figma handoff copy `상세내용은 작업중입니다` on `node-id=153:8545`
is therefore resolved by generated-client projection rather than by adding a
separate endpoint.

| Figma annotation node | Figma facade example | Canonical generated path |
| --- | --- | --- |
| `153:8545` | `riido.aiAgent.events.stream` | `aiAgent.events.stream` |
| `236:20768` | `riido.aiAgent.tasks.stop` | `aiAgent.tasks.stop` |

The generator test reads both manifests and verifies that every required
generated path exists in the upstream coverage mirror and in:

- `contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json`
- `web/generated/aiAgentClient.ts`
- `web/generated/aiAgentClient.react.ts`

## Projection Rule

Top-down:

```text
contracts Figma coverage
  -> control-plane API DSL/OpenAPI projection
  -> generated TypeScript / React wrapper comments and facade
  -> client handoff
```

Bottom-up:

```text
control-plane implementation finding
  -> local projection manifest if the issue is API/generated-client local
  -> contracts Figma coverage if the upstream screen/policy mapping is wrong
```

## Local Coverage

| Figma node | Section | Control-plane status |
| --- | --- | --- |
| `153:12742` | 컴포넌트 참여자 드롭다운 | generated client covered |
| `153:15931` | 댓글 소통 | generated client covered |
| `153:15932` | image 14 | non-decision asset |
| `153:15935` | 추가 기획 내용 | generated client covered |
| `162:23468` | 논의 필요 | generated client covered |
| `156:18767` | image 13 | non-decision asset |
| `156:19307` | 메뉴바 | client route, no endpoint |
| `156:19308` | Group 6 | non-decision asset |
| `162:23090` | 런타임 설정페이지 | generated client covered |
| `164:30658` | 데스크탑앱 온보딩-런타임 감지 O | generated client covered |
| `435:60050` | 데스크탑앱 온보딩-런타임 감지 X | generated client covered |
| `164:45736` | image 8 | non-decision asset |
| `164:45741` | Group 5 | non-decision asset |
| `236:29749` | 웹 온보딩 | product surface, no AI Agent endpoint |
| `275:22731` | 런타임 설정페이지 엠티 | generated client covered |
| `432:37336` | 에이전트 설정페이지 | generated client covered |

## Important Boundaries

- The menu section can mention route consumption, but it must not create an API
  endpoint.
- Web onboarding remains auth/team/product/distribution work. This repo must not
  expose AI Agent waitlist, marketing, or consent generated helpers from that
  Figma evidence alone.
- Runtime empty states are derived from `aiAgent.devices.runtimes`; provider
  install cards stay client/product presentation until a separate SSOT adds a
  server operation.
- Runtime settings contains an endpoint-looking Figma label at
  `node-id=129:17930`, but this repo must not project it as a canonical base
  URL, generated path, live host, or checked-in configuration value. Generated
  clients use caller-provided base URL configuration and typed operation paths.
  It is not a canonical base URL, generated path, or live host export.
- Task/subtask-only assignment is represented by task-scoped generated paths.
  This repo must not ship project, milestone, intake, mention, or property-filler
  helpers without a new owning SSOT.
- The former discussion node `node-id=162-23468`, with resolved evidence in
  `node-id=162-23475`, is now covered by generated paths. Fixture-created
  agents use `onboarding.fixtures.createAgent`, keep duplicate display names
  without suffixing, appear through `tasks.assignableAgents`, and use normal
  `agents.updateConfiguration`, `agents.delete`, and `agents.editability`
  lifecycle paths.
- Daemon detail uses `aiAgent.agents.daemon.details`; `aiAgent.agents.daemon` is
  a cache tag / namespace, not the operation generated path.
- Agent configuration update uses `aiAgent.agents.updateConfiguration`, matching
  the OpenAPI `x-riido-client.generated_path`.

## Verification

```bash
go test ./tools/reactquerygen -run TestFigmaAIAgentControlPlaneProjectionManifest -count=1
go test ./tools/reactquerygen -count=1
```

The test catches these drift classes:

- the local projection manifest names a generated path that OpenAPI does not
  expose;
- the local projection manifest requires a generated path that the mirrored
  contracts Figma coverage did not name for the same node;
- the mirrored contracts coverage no longer records the three inspected Figma
  pages, the four non-UI top-level coverage evidence nodes, or the loaded
  non-UI top-level inventory;
- the mirrored contracts coverage loses the Figma Plugin API inspection method
  that owns page registry and child-count evidence;
- the mirrored contracts coverage loses `클라이언트 전달` annotation
  normalization from `riido.*` facade examples to canonical generated paths;
- generated TypeScript or React wrapper comments no longer carry the required
  generated path;
- a no-endpoint Figma section accidentally gains forbidden helper names such as
  waitlist, marketing, consent, provider install, or project assignment paths.
