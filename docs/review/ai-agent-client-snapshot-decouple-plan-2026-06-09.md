# AI-Agent-Client 스냅샷 디커플 계획서 (2026-06-09, rev3 — 확정)

> 대상: `riido-control-plane` (브랜치 `RIID-4925-최소-기능-구현`, 기준 HEAD `5d9c088`)
> 상태: **확정 (구현 시작)** — rev3는 2차 적대 리뷰 6개 조건 반영
> 이 문서는 working-tree 문서(아직 untracked)

> ## PR 범위 (좁힘 — 가장 안전)
> 이 PR은 **"스냅샷 전체 decouple"이 아니다.** 범위는 **(1) 데몬 hot-path(`runtime-snapshot`/heartbeat)를 core-only로 분리**(full reload/save 제거) **+ (2) complete in-memory `taskThreads` invariant 유지**(agent projection 안 깨짐). 웹 agent-list가 전체 threads를 reload하는 비용(R8)과 actor 직렬화(R7)는 **후속 PR**로 분리한다.

---

## 1. 의도 (왜 고치나)

데몬 `runtime-snapshot`이 **>5초 → 타임아웃 → 등록 실패 루프**.

근본 원인: AI-agent-client **전체 상태를 DynamoDB item 1개**(`AI_AGENT_CLIENT#snapshot/CURRENT`)에 담고, **모든 작업이 그 전체를 reload+save**, **단일 goroutine `loop()`(`dynamodb_ai_agent_client_snapshot.go:153-181`)로 직렬화**. events는 full `DeviceRecord`를 embed(`ai_agent_client_api.go:452`)해 무거움. 실측 **raw 732KB**(events=200, task_threads=48).

`SyncAIAgentDaemonRuntimeSnapshot`(`ai_agent_daemon_runtime.go:63`)은 devices/daemons/events만 쓰는데도 전체 blob을 decode/encode + 직렬 actor 통과 → heartbeat(≈5초 주기)와 겹쳐 5초 초과.

### ⚠️ 크기 근거 정정 (리뷰 #7)
저장은 raw JSON이 아니라 **gzip+base64 `snapshot_gzip`**(`dynamodb_...:259`). 그래서 **raw 732KB는 "decode/encode CPU 비용" 근거**이지 DynamoDB 400KB 한도 근거가 아니다. 400KB는 **압축된 item 크기**에 적용 — 구현 전 `aws dynamodb get-item ... | jq -r .Item.snapshot_gzip.S | wc -c` 로 압축 item 크기도 실측해 한도 여유를 확인한다. (현 병목의 1차 원인은 한도 초과가 아니라 **per-op 전체 decode/encode + 직렬화**.)

---

## 2. 방향성 (어떻게 — rev2)

**데몬 hot-path를 전체 blob에서 떼는 것이 목표.** 단일 스냅샷을 컬렉션별 item으로 분리하되, **agent projection invariant를 깨지 않는다.**

### 2.1 핵심 설계 결정: in-memory taskThreads = 항상 complete (리뷰 #1, #2)
`visibleAgent`/`visibleAgents`(`ai_agent_client_development.go:1584, 1618`)와 `projectAgentWorkStatusFromThreadsLocked`(`:1640`)는 **전체 `s.taskThreads`** 로 agent의 `AssignedTaskCount`/`WorkStatus`/`Editability`를 재계산해 **`s.agents`에 기록**한다. `ListAssignableAgents`(`:672`)도 `visibleAgents()`를 거친다. → **task를 일부만 로드하면 agent가 실제 running인데 idle로 계산** → "참여자 새로고침 후 빔/상태 꼬임/다른 task 동시 시작" 회귀.

**결정: lazy thread load 폐기. in-memory `s.taskThreads`는 항상 complete set으로 유지한다.**
- threads는 **별도 DynamoDB item**(`#thread/<taskID>`)으로 분리하되, 부팅 시 전부 로드하고 agent/thread 작업 시 전부 reload(다중 인스턴스 일관성).
- **이득**: **core item이 작아져** 데몬 hot-path(`runtime-snapshot`/heartbeat/device)가 **core(+이벤트 발생 시 events)만** read/write → 타임아웃 해소. threads는 데몬 path가 절대 안 건드림.
- (대안 Option B = agent work-status를 core의 active-assignment index로 승격해 threads를 진짜 lazy하게. → denormalization 일관성 위험이 커 **이번엔 채택 안 함**, 웹 agent-list 최적화가 필요해지면 후속.)

### 2.2 컬렉션 분리 + per-item I/O
- `#core/CURRENT`: devices·credentials·daemons·agents·fixtures·seq (작고 자주).
- `#events/CURRENT`: replay 로그(최신 200, gzip) + **next_event_seq를 이 item 안에** 둠(리뷰 #6 — core와 분리 유지).
- `#thread/<taskID>`: 그 task의 thread 레코드(gzip).
- `#snapshot/CURRENT`(legacy): 마이그레이션 때만 읽음.
- 다중 인스턴스 일관성: 각 작업이 **건드리는 item만 reload**(ConsistentRead) 후 **해당 컬렉션만 targeted restore**(나머지·subscribers 보존).

### 2.3 데몬 hot-path 최소화 (리뷰 #5)
`SyncAIAgentDaemonRuntimeSnapshot`은 의미 있는 변화에만 event를 append(`ai_agent_daemon_runtime.go:171`). 따라서:
- **기본: core만** reload+save(device last-seen 갱신).
- **event append가 실제 발생할 때만** events item write.
- heartbeat(변화 없음)마다 events item을 gzip decode/encode하지 않는다.

---

## 3. 상세 계획 (rev2)

### 3.1 call-site 분류 (정정 — 리뷰 #2)
- **core만**: Enroll, Authorize, Bootstrap, Connect, Sync...RuntimeSnapshot, Control(Device)Daemon, Get(Device)Daemon, ListDevices, ListBindings, Create/Update agent, GetEditability, ListFixtures, CreateFromFixture. (event append하는 것만 +events.)
- **core + 전체 threads**(agent projection 의존): **`ListAssignableAgents`(정정: core-only 아님)**, `ListWorkspaceAssignedAgentProfiles`, `DeleteAIAgent`, 기타 `visibleAgents`/`agentForMutation` 경유.
- **core + 해당 task threads(+events)**: ListTaskThreads(taskID), Assign/Unassign/Delete assignment, SubmitComment, CreateThreadMessage, StopTask(+Assignment), GetThreadStreamSubscription, Reconcile, RecordThreadProgress, RecordAssignmentEvent.
- **events만**(+core 가시성 필터): AIAgentClientEvents, SubscribeAIAgentClientEvents.

> 단, 2.1 결정상 **agent projection이 닿는 모든 경로는 complete threads가 메모리에 있어야** 한다 → 부팅 시 전체 로드 + agent-projection 작업 시 전체 reload.

### 3.2 cross-item write 원자성 (리뷰 #3) — TransactWriteItems
현재는 mutation 후 완성 상태를 **단일 PutItem**(원자적). split 후 core/thread/events를 따로 쓰면 "core 성공·thread 실패" 부분실패가 새 장애 모드. → TransactWriteItems를 쓰되 **"모든 경우"가 아니라 경로를 나눈다**(리뷰 #3):
- **소규모(transaction 경로)**: 한 task를 건드리는 mutation = core(agent) + thread(taskID) + events ≤ 수 개 item → **`TransactWriteItems`(동일 테이블, 원자적)** 로 묶는다. 예: `RecordThreadProgress`, Assign/Unassign, Submit/CreateMessage, Stop.
- **대규모(batch+marker 경로)**: `DeleteAIAgent` 처럼 **여러 task thread를 한꺼번에** 바꾸는 케이스는 item 수가 **TransactWriteItems 100개 한도를 넘을 수 있다** → 단일 transaction 금지. 대신 batch write + (필요 시 작은 publish marker)로 진행하고, 부분 적용 가능성을 문서화한다.
- **단일 컬렉션** mutation = 단순 PutItem. event-seq는 events item 내부 갱신(또는 transaction 내 conditional)으로 **core와 분리**(리뷰 #6).

### 3.3 마이그레이션 (리뷰 #4) — publish marker / 원자적 split
"core 있으면 split 우선 + 없는 task=빈 것" 정책은 **core 저장 후 task 저장 실패 시 legacy thread 유실**. → 금지. **2-state publish marker 필수**(리뷰 #4):
- **마커 `AI_AGENT_CLIENT#migration / STATE`** 가 `writing` | `complete` 두 상태.
- **쓰기 순서**: ① 마커=`writing` → ② 모든 split item(core·events·각 task thread) 기록 → ③ **마지막에** 마커=`complete` publish(가능하면 core publish와 함께 작은 transaction). 중간 실패는 `writing`에 머무름.
- **읽기 규칙**: 마커가 **`complete`일 때만 split을 truth로** 사용. 아니면(없음/`writing`) **legacy(`#snapshot`)를 truth로** 본다. → 어떤 부분실패에도 데이터 유실 없음.
- task ≤ ~97개면 split 전체를 단일 `TransactWriteItems`로(마커까지 원자) — 부분실패 불가. 더 많으면 batch write 후 마커=complete.
- legacy item은 **1릴리스 보존**(롤백 안전). 마커 complete 안정화 후에만 삭제(후속).

### 3.4 일관성 모델
- DynamoDB = SSOT, in-memory = reload 캐시(ConsistentRead).
- 데몬 hot-path: core(+이벤트 발생 시 events) reload/save만.
- agent-projection/thread 작업: core + (전체 또는 해당)threads reload — complete invariant 유지.
- 모든 targeted restore는 **subscribers/nextSubscriberID 무조건 보존(리뷰 #3 SSE)**.

### 3.5 순서 단계 (각 단계 compile + test green)
- **Step 0 — event-seq 카운터를 events item으로**: dev store에 `nextClientEventSeq` 추가(스캔 대신 카운터), **events 스냅샷에 영속**(core 아님). restore로 시드. (리뷰 #6.)
- **Step 1 — DynamoDB store per-item + TransactWrite + 새 인터페이스**: `loop()` actor 유지하되 `LoadCore/SaveCore/LoadEvents/SaveEvents/LoadTaskThread/SaveTaskThread/LoadAllTaskThreads/LoadLegacyCombined` + **`TransactWriteCollections`**(core/threads/events 원자적) 추가. `#core/#events/#thread/#migration` pk 상수. 테스트 메모리 store 새 인터페이스로 포팅(원자성 시뮬 포함).
- **Step 2 — dev store targeted restore/projection**: `restoreCore/Events/TaskThread/AllTaskThreads` + `coreSnapshot/eventsSnapshot/taskThreadSnapshot/allTaskThreadsSnapshot` 분리(`snapshot()` `:420`, `restoreSnapshotWithSubscriberMode` `:497`에서). subscribers 항상 보존. 기존 `snapshot`/`restoreSnapshot`은 합성 유지.
- **Step 3 — wrapper 재타겟(데몬 path 먼저)**: `reloadCore`/`reloadCoreAndEvents`/`reloadCoreAndAllThreads`/`reloadCoreAndTask`/`reloadEvents` + save 헬퍼(단일은 PutItem, 다중은 TransactWrite). **`SyncAIAgentDaemonRuntimeSnapshot`부터 core-only(+event append 시에만 events, 2.3)** → 타임아웃 해소. 그다음 agent-projection 경로는 core+AllThreads. 기존 cross-process 일관성 테스트가 회귀 가드 + "Sync는 thread item 미접근" spy 테스트 추가.
- **Step 4 — 마이그레이션(3.3)**: `OpenPersistentAIAgentClientStore`(`:56`)에 마커 기반 로드/원자적 split. 마이그레이션 멱등·롤백 테스트(legacy seed → open → 동일 read; 부분실패 시 legacy 유지).
- **Step 5 — per-item bound/정리**: `retainLatest*`를 per-collection save에. 결합-blob 주석 제거. `snapshot_size_fix_test.go` per-item bound로.

### 3.6 문서(SSOT)
- `docs/20-domain/ai-agent-client-api.md:273-292` — core/events/per-task 키 스킴, per-item reload 일관성, **complete-threads invariant**, TransactWrite 원자성, 마이그레이션 마커.
- `docs/20-domain/saas-control-plane.md:311-328` — runtime-snapshot core-only(+event 시 events) 지연 해소.
- `docs/30-architecture/config-reference.md:25-27` — 테이블 다중 `AI_AGENT_CLIENT#` item kind.

### 3.7 리스크/결정 (rev2)
- **R1 (해결)** agent projection 전체-threads 의존 → lazy 폐기, complete in-memory 유지. `ListAssignableAgents` 등은 core+AllThreads.
- **R2 (해결)** cross-item 원자성 → TransactWriteItems.
- **R3 (해결)** 마이그레이션 유실 → publish marker + 원자적 split, 성공 전 legacy truth.
- **R4 (완화)** runtime-snapshot hot-path → core-only, event는 발생 시만.
- **R5 (해결)** event-seq → events item(또는 atomic), core와 분리.
- **R6 (정정)** 400KB 근거 → 압축 item 크기 실측 분리; 1차 원인은 per-op decode/encode+직렬화.
- **R7 (잔여)** 단일 actor `loop()` 직렬화 — split로 payload 축소(타임아웃엔 충분). per-key 동시성은 후속.
- **R8 (잔여)** 웹 agent-list가 전체 threads reload(complete 유지 비용). 데몬 path 아님. 커지면 Option B(active-assignment index) 후속.

---

## 4. 검토 필요 결정
1. **threads 모델: Option A(complete in-memory, 권장) vs Option B(core active-assignment index, 후속)** — 본 계획은 A.
2. cross-item mutation 원자성: **TransactWriteItems** 채택(동의 필요 — WCU 2배, ≤100 item 제약).
3. 마이그레이션: **마커 + 원자적 split**(task ≤~97개면 단일 transaction).

핵심 파일: `ai_agent_client_persistence.go`, `dynamodb_ai_agent_client_snapshot.go`, `ai_agent_client_development.go`, `ai_agent_daemon_runtime.go`, `cmd/riido_ai_server/main.go`.
