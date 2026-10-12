# Agent 작업 알림 계약

RIID-5829. 배정한 사람이 결과와 확인 요청을 놓치지 않도록 실행 중인 데스크탑에 전달한다.

- Assignment의 created_by·workspace_id·team_id는 배정 시 인증 주체와 검증된 작업 맥락에서 저장한다. 과거 created_by가 비어 있으면 소유자로 추정하지 않는다.
- Assignment.state는 기존 실행 FSM을 유지한다. running과 confirmation_reason(answer_required/approval_required)이 함께 있으면 WorkStatus는 waiting_for_user다. 로그는 대기를 해제하지 않고 명시적 running 재개와 terminal 결과만 해제한다.
- TaskEvent의 event_id는 assignment ID와 영속 event sequence로 구성하고 at이 발생 시각이다. 별도 agent_task_notification에 workspace/task/agent/assignment/run ID, created_by, recipient_principal_id, notification_type, confirmation_reason, work_status, assignment_state, team_id를 제공한다.
- 알림 종류는 result_ready, answer_required, approval_required, failed다. completed는 실행 결과 준비이며 업무 완료 상태를 변경하지 않는다. 일반 로그·경고·복구 중 provider_error·사용자 취소는 알림을 만들지 않는다.
- 수신자는 배정자 한 명이다. 공유 에이전트 소유권·관리자 역할로 수신 권한을 확대하지 않는다. SSE 발송 직전 현재 bearer로 해당 workspace/task의 읽기 권한을 재확인하고 실패하면 전달하지 않는다.
- 최초 연결 및 재연결은 live-only다. SSE 연결이 끊긴 구간은 과거 알림으로 재생하지 않으며 현재 실행 상태는 작업 상세 재조회로 확인한다. 프로세스 완전 종료 원격 푸시는 범위 밖이다.
- 같은 상태/사유의 재전송과 종료 뒤 늦은 이벤트는 추가 알림을 만들지 않는다. Assignment 저장 이후 알림 투영이 실패해 재시도하더라도 이미 저장한 event ID를 재사용한다. 클라이언트는 사용자·워크스페이스별 event ID를 최대 512개 보존하고 Electron은 창 사이 중복을 막는다.
- 클릭은 발신 webview의 같은 origin에서 /workspace/teams/team/tasks/task로 이동하고 소유 창을 복원한다. 이전 생산자가 team_id를 제공하지 않으면 현재 작업 메타데이터에서 같은 workspace/task임을 확인한 후 경로를 구성한다. 이동 시 기존 인증·권한 가드가 다시 적용된다.

검증: Go의 TaskNotification 상태·수신자·권한·스냅샷 테스트, daemon provider/SaaS 변환 테스트, client 필터·재접속 테스트, desktop 창 복원·경로·중복 테스트. macOS/Windows 실제 알림 표시와 권한 정책은 OS별 수동 확인이 필요하다. 롤백 시 새 optional 필드를 무시하는 구버전과 호환되며 신규 알림 이벤트는 구버전 클라이언트에서 소비하지 않는다.

## 로컬 검증과 출시 조건

- control-plane과 daemon은 contracts 알림 커밋의 immutable pseudo-version을 go.mod/go.sum에 고정한다. PR 검증은 로컬 replace 없이 `GOWORK=off go test ./...`로 실행했다. RIID-5827 선행 PR 위에 쌓인 서버·계약 PR이므로 최신 main 통합과 정식 contracts 릴리즈 버전 정리는 병합 전에 필요하다.
- contracts와 control-plane: `go test ./...`. daemon: `go test ./internal/agentbridge/... ./internal/provider/...`.
- desktop: `pnpm exec tsc -b` 후 `node scripts/verify-task-notification.js`로 경로·소유 창 복원·사용자별 중복 방지를 검증한다.
- client: `pnpm exec vitest run src/app/domains/agent/slices/receive-task-notifications` 및 `pnpm run typecheck`. 인증된 앱 셸의 SSE→Electron IPC 연결은 `pnpm exec playwright test --config playwright.agent-notifications.config.ts`로 검증한다. 이 E2E는 외부 API를 차단하고 IPC를 모의하며 실제 OS 표시를 증명하지 않는다.
- 기존 provider의 확인 요청·명시적 재개 이벤트를 전달한다. 사용자 응답을 provider에 보내는 입력 UI/API의 신설은 이 알림 계약에 포함하지 않는다.
