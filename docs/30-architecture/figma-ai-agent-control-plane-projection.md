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
coverage map, not only the primary `UI` page. Legacy non-UI Wireframe nodes
that carry current product/API meaning are projected as absorptions, not as new
endpoints. Each one points back to the current UI top-level node that already
owns the generated-client surface.
The projection gate decodes both the local projection manifest and this mirrored
contracts coverage manifest with unknown-field and trailing-document rejection;
control-plane may mirror the upstream SSOT, but it must not silently accept a
shape that the contracts-owned manifest schema does not expose.

The mirrored page registry keeps the contracts-owned inspection method:
`figma.root.children` from the Figma Plugin API is the page registry authority,
but non-current page child counts are authoritative only after
`await figma.setCurrentPageAsync(page); page.children.length`. Passive page
objects can be lazy/unloaded, so metadata XML/read output and unloaded
`page.children.length` reads remain supporting evidence only and must not
redefine page-level counts in the control-plane mirror.

The mirror also preserves
`figma-metadata-page-list-underreports-pages.v1`. On 2026-06-02, no-`nodeId`
Figma `get_metadata` listed only `129:5215` `UI`, while the Figma Plugin API
page registry returned `129:5215`, `42:3014`, and `0:1`. Control-plane must not
turn that supporting metadata output into a generated-client decision: it must
not remove `expected_pages`, drop `non_ui_top_level_inventory`, or delete
`legacy_non_ui_absorptions`.

The mirror also preserves `figma-headless-file-key-placeholder.v1`. A live
`use_figma` root/category inspection for file `MUOd9lctoEHASUStN3vUuK` returned
`figma.fileKey=headless`, even though the same invocation read the AI Agent pages
and annotation categories. Control-plane must keep the contracts mirror's
`figma.file_key` and local `source_contracts_manifest` identity from the
contracts manifest and must not replace either with the headless runtime
placeholder. The value `headless` is a headless runtime placeholder only.

The mirror also preserves `figma-onboarding-page-load-timeout.v1`. Current live
reads of `node-id=42:3014` (`Wireframe - 온보딩`) can time out after 120s when
using Figma `get_metadata(nodeId=42:3014)` or `use_figma` scripts that attempt
full inventory traversal after `await figma.setCurrentPageAsync(page)`. RIID-4900
live annotation traversal counted `child_count=84` for that page, while the
known `non_ui_top_level_inventory` still identifies 83 top-level nodes because
the extra top-level node identity is unresolved. Direct registered-node lookup
for `236:33845` and `236:33847` still preserves the six onboarding `riido.*` `API Generated` annotations. Control-plane must not treat that timeout as proof
that onboarding generated-client coverage disappeared: it must not remove
`expected_pages`, drop `non_ui_top_level_inventory`, remove onboarding generated
paths, or mark onboarding generated paths unresolved. Control-plane also must
not redefine the Figma top-level identity; the `child_count=84` /
`known_inventory_count=83` split remains contracts-owned evidence.

The local `source_contracts_manifest.stabilized_by` list mirrors the full
upstream coverage provenance used by this projection:
`teamswyg/riido-contracts#38`, `teamswyg/riido-contracts#39`,
`teamswyg/riido-contracts#45`, `teamswyg/riido-contracts#46`,
`teamswyg/riido-contracts#51`, `teamswyg/riido-contracts#52`,
`teamswyg/riido-contracts#54`, `teamswyg/riido-contracts#55`,
`teamswyg/riido-contracts#56`, `teamswyg/riido-contracts#57`,
`teamswyg/riido-contracts#58`, `teamswyg/riido-contracts#60`,
`teamswyg/riido-contracts#62`, `teamswyg/riido-contracts#63`,
`teamswyg/riido-contracts#64`, `teamswyg/riido-contracts#65`,
`teamswyg/riido-contracts#66`, and `teamswyg/riido-contracts#67`. The
projection gate treats the full upstream coverage provenance as part of the
mirror contract, not as PR-description trivia.

After `teamswyg/riido-contracts#53`, the mirrored contracts coverage fixture
also carries top-level `stabilized_by`. The projection gate compares that source
field with local `source_contracts_manifest.stabilized_by` so control-plane
does not preserve upstream history from local memory alone.

That is separate from limitation-local provenance. The specific
`figma-metadata-page-list-underreports-pages.v1` limitation entered the
upstream contracts coverage in `teamswyg/riido-contracts#52`, so #52 remains
the local provenance for that tooling limitation while the full list above
identifies the whole contracts coverage history consumed by control-plane.

`teamswyg/riido-contracts#54` adds Figma planning node `432:46849`
(`Ex AI - 온보딩 순서 변경 메모`). Control-plane absorbs it as onboarding
generated-client projection, not as a persisted draft API. The revised order is
agent draft/configuration, runtime selection, then workspace selection, but the
draft is client-local until final submit uses the existing fixture/direct create
operations with selected `workspace_id` and `runtime_id`.

The mirror also preserves contracts-owned annotation evidence for Figma Dev
Mode category `700:0` / `API Generated`. The mirrored manifest uses
`api_generated_annotations` and `api_generated_annotation_inventory` so the
field names match the current category authority. Those annotations may show
frontend facade examples such as `riido.v2.aiAgent.events.stream` or
`riido.v2.aiAgent.tasks.stop`. Current labels keep the workspace-scoped facade
path on the first line, then add Korean context for the operation kind and
background:

```text
riido.v2.aiAgent.tasks.stop
종류: Mutation
배경: 작업 중인 Agent에게 중지 요청을 보냅니다. daemon은 이 요청을 읽어 provider 실행을 강제 중지합니다.
```

This repo does not treat the leading `riido.` variable name as a contracts
generated path; the projection gate normalizes those examples to canonical
generated paths such as `aiAgent.events.stream` and `aiAgent.tasks.stop`. When
Figma's live label is `riido.v2.aiAgent.*`, the `v2.` segment is live handoff
context and not part of the canonical `aiAgent.*` path. The gate then verifies
that both the canonical path and the live generated-client access example appear
in `web/generated/aiAgentClient.ts` and `web/generated/aiAgentClient.react.ts`.

`teamswyg/riido-contracts#62` records this as a live Figma annotation content
policy. Control-plane mirrors it as generated-client handoff context:
`riido.v2.aiAgent.*` Figma labels must have a facade path,
`종류: Query | Mutation | SSE Stream`, and Korean `배경:` text, while the
generated TypeScript artifact still derives canonical paths from OpenAPI. The
mirrored 2026-06-03 live inspection counts are 84 `riido.*` annotations on
`129:5215` `UI`, 6 on `42:3014` `Wireframe - 온보딩`,
and 0 on `0:1` `Wireframe`, all with zero missing kind/background counts.
RIID-4898 split the former bundled daemon control label into separate
`riido.v2.aiAgent.agents.daemon.start`, `restart`, and `stop` Mutation
annotations so Figma search matches generated-client facade paths exactly.
`teamswyg/riido-contracts#66` tightens that mirror: the annotation
`operation_kind` must match generated OpenAPI transport. `text/event-stream`
responses are `SSE Stream`, non-stream `GET` operations are `Query`, and
non-`GET` operations are `Mutation`.
`teamswyg/riido-contracts#67` adds the v2 counterpart guard. Figma keeps
frontend-searchable facade labels such as `riido.v2.aiAgent.tasks.stop`; each
mirrored `api_generated_annotation_inventory` item also keeps canonical
`aiAgent.*` and matching `v2.aiAgent.*` generated paths in the same source
coverage entry and in the generated TypeScript comments. That source coverage entry
requirement is what keeps the Figma label from becoming an independent API naming
authority.

`teamswyg/riido-contracts#63` records old category `39:0` / `클라이언트 전달` as
retired. The mirrored live usage count is 0, so generated-client delivery must
not point new handoff notes at that category. The category definition may still
exist in Figma, but automation does not delete shared design categories unless a
design owner explicitly approves it.

| Figma annotation node | Figma facade example | Kind | Canonical generated path | v2 counterpart |
| --- | --- | --- | --- | --- |
| `153:8545` | `riido.v2.aiAgent.events.stream` | SSE Stream | `aiAgent.events.stream` | `v2.aiAgent.events.stream` |
| `236:20768` | `riido.v2.aiAgent.tasks.stop` | Mutation | `aiAgent.tasks.stop` | `v2.aiAgent.tasks.stop` |

The broader screen-level Figma handoff pass also labels participant assignment,
task-thread reply, runtime settings, onboarding fixture, direct create, edit,
delete, and editability nodes with `Query`, `Mutation`, or `SSE Stream`
background text. Contracts owns that full list in
`api_generated_annotation_inventory`; control-plane mirrors it to prove every
facade path below exists in OpenAPI and generated TypeScript comments.

| UI area | Facade path | v2 counterpart | Kind | Background shown in Figma |
| --- | --- | --- | --- | --- |
| Participant dropdown / task details | `riido.v2.aiAgent.tasks.assignableAgents` | `v2.aiAgent.tasks.assignableAgents` | Query | 참여자 드롭다운에서 현재 task/subtask에 배정할 수 있는 Agent 목록을 조회합니다. task 배정 mutation에 넘길 agent_id는 이 응답의 agents[]에서 얻습니다. |
| Participant dropdown / task details | `riido.v2.aiAgent.tasks.assign` | `v2.aiAgent.tasks.assign` | Mutation | 작업에 Agent를 참여자로 배정하고 daemon이 런타임으로 작업을 시작할 수 있는 서버 상태를 만듭니다. |
| Participant dropdown / task details | `riido.v2.aiAgent.tasks.unassign` | `v2.aiAgent.tasks.unassign` | Mutation | 참여자에서 Agent를 제거합니다. 진행 중이면 중지 요청/큐 해제 흐름으로 이어집니다. |
| Task thread | `riido.v2.aiAgent.tasks.threads` | `v2.aiAgent.tasks.threads` | Query | 작업의 완료/진행 중 Agent thread cold collection을 조회합니다. active_stream이 있으면 SSE로 이어집니다. |
| Task thread | `riido.v2.aiAgent.tasks.threadMessages.create` | `v2.aiAgent.tasks.threadMessages.create` | Mutation | 특정 thread_id에 다음 지시/답글을 추가하고 Agent 응답을 이어서 요청합니다. |
| Task thread | `riido.v2.aiAgent.tasks.submitComment` | `v2.aiAgent.tasks.submitComment` | Mutation | 호환용 댓글 제출 경로입니다. thread_id 없이 입력하면 서버가 적절한 thread 응답 흐름을 처리합니다. |
| Task thread | `riido.v2.aiAgent.events.stream` | `v2.aiAgent.events.stream` | SSE Stream | threads 조회 결과에 active_stream이 있을 때만 연결해 진행 상태와 thread 갱신 이벤트를 받습니다. |
| Task thread | `riido.v2.aiAgent.tasks.stop` | `v2.aiAgent.tasks.stop` | Mutation | 작업 중인 Agent에게 중지 요청을 보냅니다. daemon은 이 요청을 읽어 provider 실행을 강제 중지합니다. |
| Runtime and agent settings | `riido.v2.aiAgent.devices.runtimes` | `v2.aiAgent.devices.runtimes` | Query | 계정 소유 device에서 감지된 runtime 목록과 온라인/오프라인 상태를 조회합니다. 화면은 SaaS 값을 신뢰합니다. |
| Runtime settings | `riido.v2.aiAgent.agents.daemon.details` | `v2.aiAgent.agents.daemon.details` | Query | Agent에 연결된 daemon/runtime 상세와 제어 가능 상태를 SaaS 기준으로 조회합니다. |
| Runtime settings | `riido.v2.aiAgent.agents.daemon.start` | `v2.aiAgent.agents.daemon.start` | Mutation | stopped/offline 상태의 daemon에 시작 요청을 남깁니다. daemon 실행 트리거는 desktop이 담당하고, SaaS는 시작 요청 상태와 runtime projection을 제공합니다. |
| Runtime settings | `riido.v2.aiAgent.agents.daemon.stop` | `v2.aiAgent.agents.daemon.stop` | Mutation | SaaS에 daemon 중지 요청을 남깁니다. daemon은 요청을 읽은 뒤 스스로 종료합니다. |
| Runtime settings | `riido.v2.aiAgent.agents.daemon.restart` | `v2.aiAgent.agents.daemon.restart` | Mutation | SaaS에 daemon 재시작 요청을 남깁니다. daemon은 polling으로 요청을 읽어 실행합니다. |
| Onboarding | `riido.v2.aiAgent.onboarding.fixtures` | `v2.aiAgent.onboarding.fixtures` | Query | 리도/영실/홍도/지원처럼 제품이 제공하는 초기값 목록을 조회합니다. template entity가 아니라 fixture입니다. |
| Onboarding | `riido.v2.aiAgent.onboarding.fixtures.createAgent` | `v2.aiAgent.onboarding.fixtures.createAgent` | Mutation | 선택한 fixture 값을 기반으로 일반 Agent를 생성합니다. fixture 자체를 생성하는 기능은 아닙니다. |
| Agent settings / direct setting | `riido.v2.aiAgent.agents.create` | `v2.aiAgent.agents.create` | Mutation | 직접 설정 화면에서 워크스페이스 안에 새 Agent를 생성합니다. 신규 v2 create는 workspace_id를 포함합니다. |
| Agent settings | `riido.v2.aiAgent.bootstrap` | `v2.aiAgent.bootstrap` | Query | AI Agent 설정/온보딩 초기 화면에 필요한 bootstrap.agents[]를 조회합니다. 설정/목록 화면의 agent list와 agent_id 출처이며, task 참여자 드롭다운은 tasks.assignableAgents를 사용합니다. |
| Agent settings | `riido.v2.aiAgent.agents.updateConfiguration` | `v2.aiAgent.agents.updateConfiguration` | Mutation | 할당 작업이 없는 Agent의 이름, 썸네일, 설명, 지침, 런타임, 모델, 공개 범위를 저장합니다. |
| Agent settings | `riido.v2.aiAgent.agents.editability` | `v2.aiAgent.agents.editability` | Query | Agent를 수정할 수 있는지 먼저 조회합니다. 할당된 작업이 있으면 저장/수정 UI는 막혀야 합니다. |
| Agent settings | `riido.v2.aiAgent.agents.delete` | `v2.aiAgent.agents.delete` | Mutation | Agent 삭제를 요청합니다. 진행/예약 중 작업은 서버 정책에 따라 중지 또는 큐 해제됩니다. |

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

## Legacy Non-UI Absorptions

These nodes come from page `0:1` (`Wireframe`). The upstream contracts coverage
marks them as covered because they are meaningful, but the control-plane
projection does not create a second route family for them. They inherit the
generated paths from the current UI section named in the last column.

| Legacy Figma node | Section | Absorbed by current UI node | Control-plane projection |
| --- | --- | --- | --- |
| `13:3789` | 런타임 | `162:23090` 런타임 설정페이지 | `aiAgent.devices.runtimes`, `v2.aiAgent.devices.runtimes` |
| `86:9988` | 런타임 | `162:23090` 런타임 설정페이지 | device/runtime read plus agent-bound daemon detail |
| `17:3551` | 에이전트 | `432:37336` 에이전트 설정페이지 | bootstrap, agent lifecycle, editability, and runtime candidate paths |
| `17:4231` | 에이전트 수정 | `432:37336` 에이전트 설정페이지 | `agents.updateConfiguration`, `agents.editability`, and runtime candidate paths |
| `84:9846` | 에이전트 추가 | `432:37336` 에이전트 설정페이지 | direct `agents.create`, bootstrap, and runtime candidate paths |
| `17:2871` | 데몬 상세 | `162:23090` 런타임 설정페이지 | agent-bound daemon detail/start/restart/stop and event stream paths |
| `17:3111` | 런타임 상세 | `162:23090` 런타임 설정페이지 | device/runtime read plus agent-bound daemon detail |

## Non-UI Planning Absorptions

These nodes come from loaded non-UI Figma planning pages, not from the current
UI page. Control-plane projects only the generated-client effect that contracts
owns and does not create extra HTTP/SSE endpoints for planning notes.

| Planning Figma node | Section | Control-plane projection |
| --- | --- | --- |
| `432:46849` | Ex AI - 온보딩 순서 변경 메모 | local onboarding draft/configuration is client-owned; final create uses `aiAgent.onboarding.fixtures`, `aiAgent.onboarding.fixtures.createAgent`, `aiAgent.agents.create`, and v2 equivalents |

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
- Figma planning node `432:46849` changes onboarding order, but the first
  agent draft/configuration step stays client-local. Control-plane must not add
  a persisted draft route or workspace-less create route from this note.

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
  pages, the twelve non-UI top-level coverage evidence nodes, the seven legacy
  semantic Wireframe absorptions, the onboarding planning absorption, or the
  loaded non-UI top-level inventory;
- the mirrored contracts coverage loses the Figma Plugin API inspection method
  that owns page registry and child-count evidence;
- the mirrored contracts coverage loses `API Generated` annotation
  normalization from `riido.*` facade examples to canonical generated paths;
- generated TypeScript or React wrapper comments no longer carry the required
  generated path;
- a no-endpoint Figma section accidentally gains forbidden helper names such as
  waitlist, marketing, consent, provider install, or project assignment paths.
