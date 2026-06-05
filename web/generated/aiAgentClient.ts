// 이 파일은 tools/reactquerygen으로 생성됩니다. 직접 수정하지 마세요.
// 원본: contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json

import type { QueryClient, UseMutationOptions, UseQueryOptions } from '@/lib/react-query';

/**
 * task thread action 요청 이후 client가 즉시 반영할 agent 작업 상태 응답입니다.
 */
export interface AIAgentTaskActionResponse {
  agent_id: string;
  assignment_id?: string;
  assignment_state: AgentAssignmentState;
  comment_kind: AgentTaskCommentKind;
  message: string;
  run_id: string;
  schema_version: string;
  task_id: string;
  thread_id: string;
  work_status: AgentWorkStatus;
}

/**
 * 활성 thread filter들과 함께 사용하는 공통 client SSE stream handoff link입니다.
 */
export interface AIAgentTaskEventStreamLink {
  event_type: "agent_thread_progress";
  href: string;
  rel: "agent_thread_progress_stream";
}

/**
 * task 화면 진입 시 먼저 읽는 AI Agent thread cold collection 응답입니다.
 */
export interface AIAgentTaskThreadCollectionResponse {
  active_stream?: AIAgentTaskThreadStreamLink;
  schema_version: string;
  task_id: string;
  threads: AIAgentTaskThreadRecord[];
}

/**
 * task 화면에 표시할 AI Agent thread의 cold record입니다.
 */
export interface AIAgentTaskThreadRecord {
  agent_id: string;
  assignment_id?: string;
  assignment_state: AgentAssignmentState;
  comment_kind: AgentTaskCommentKind;
  completed_at?: string;
  lines: AgentThreadProgressLine[];
  message: string;
  run_id: string;
  source_comment_id?: string;
  source_message_id?: string;
  started_at?: string;
  task_id: string;
  thread_id: string;
  work_status: AgentWorkStatus;
}

/**
 * 활성 task thread가 있을 때 client가 연결할 SSE stream handoff link입니다.
 */
export interface AIAgentTaskThreadStreamLink {
  event_type: "agent_thread_progress";
  href: string;
  rel: "agent_thread_progress_stream";
  run_id: string;
  task_id: string;
  thread_id: string;
}

/**
 * task 내 여러 active AI Agent thread를 하나의 SSE stream으로 구독하기 위한 handoff 응답입니다.
 */
export interface AIAgentTaskThreadStreamSubscriptionResponse {
  active_thread_filters: AIAgentTaskThreadStreamTarget[];
  schema_version: string;
  stream: AIAgentTaskEventStreamLink;
  task_id: string;
}

/**
 * 공통 SSE stream에서 client가 적용할 활성 task thread filter입니다.
 */
export interface AIAgentTaskThreadStreamTarget {
  agent_id: string;
  run_id: string;
  thread_id: string;
}

/**
 * agent 삭제와 runtime lifecycle에 의해 변경되는 task assignment 상태입니다.
 */
export type AgentAssignmentState = "queued" | "running" | "stopping" | "stopped" | "completed" | "failed" | "unassigned";

/**
 * task 참여자 드롭다운에서 표시할 agent 목록 응답입니다. bootstrap.agents[]와 달리 task_id 기준 권한/상태를 반영합니다.
 */
export interface AgentClientListResponse {
  /**
   * 현재 task/subtask 참여자 드롭다운에서 배정 가능한 agent 목록입니다. 여기의 agent_id를 tasks.assign mutation에 전달합니다.
   */
  agents: AgentClientRecord[];
  schema_version: string;
}

/**
 * v2 task 참여자 드롭다운에서 표시할 workspace-scoped agent 목록 응답입니다. v2.aiAgent.bootstrap.agents[]와 달리 workspace_id와 task_id 기준 권한/상태를 반영합니다.
 */
export interface AgentClientListResponseV2 {
  /**
   * 현재 workspace의 task/subtask 참여자 드롭다운에서 배정 가능한 agent 목록입니다. 여기의 agent_id를 v2.aiAgent.tasks.assign mutation에 전달합니다.
   */
  agents: AgentClientRecordV2[];
  schema_version: string;
}

/**
 * client 설정/목록/배정 화면에 전달되는 agent 요약 record입니다. agent_id는 bootstrap.agents[]와 tasks.assignableAgents.agents[]에서 client가 mutation에 넘기는 안정 식별자입니다.
 */
export interface AgentClientRecord {
  /**
   * client가 agent 상세/수정/삭제/배정 mutation에 전달하는 안정 식별자입니다. 설정/목록 화면에서는 bootstrap.agents[]에서 얻고, task 참여자 드롭다운에서는 tasks.assignableAgents.agents[]에서 얻습니다.
   */
  agent_id: string;
  assigned_task_count: number;
  created_at: string;
  description?: string;
  editability: AgentEditability;
  instruction?: string;
  is_owned_by_viewer: boolean;
  model_id?: string;
  model_label?: string;
  name: string;
  owner_principal_id: string;
  profile_thumbnail_url?: string;
  runtime_id?: string;
  runtime_kind?: RuntimeKind;
  /**
   * fixture-created agent에 복사되는 avatar fallback color입니다. 예: #C9A452. 직접 설정 agent는 생략될 수 있습니다.
   */
  tmp_color?: string;
  updated_at: string;
  visibility: AgentVisibility;
  work_status: AgentWorkStatus;
}

/**
 * 단일 agent 변경 후 최신 agent record를 반환하는 응답입니다.
 */
export interface AgentClientRecordResponse {
  agent: AgentClientRecord;
  schema_version: string;
}

/**
 * v2 단일 agent 변경 후 최신 workspace-scoped agent record를 반환하는 응답입니다.
 */
export interface AgentClientRecordResponseV2 {
  agent: AgentClientRecordV2;
  schema_version: string;
}

/**
 * v2 client 설정/목록/배정 화면에 전달되는 workspace-scoped agent 요약 record입니다. agent_id는 v2.aiAgent.bootstrap.agents[]와 v2.aiAgent.tasks.assignableAgents.agents[]에서 client가 mutation에 넘기는 안정 식별자입니다.
 */
export interface AgentClientRecordV2 {
  /**
   * client가 workspace-scoped agent 상세/수정/삭제/배정 mutation에 전달하는 안정 식별자입니다. 설정/목록 화면에서는 v2.aiAgent.bootstrap.agents[]에서 얻고, task 참여자 드롭다운에서는 v2.aiAgent.tasks.assignableAgents.agents[]에서 얻습니다.
   */
  agent_id: string;
  assigned_task_count: number;
  created_at: string;
  description?: string;
  editability: AgentEditability;
  instruction?: string;
  is_owned_by_viewer: boolean;
  model_id?: string;
  model_label?: string;
  name: string;
  owner_principal_id: string;
  profile_thumbnail_url?: string;
  runtime_id?: string;
  runtime_kind?: RuntimeKind;
  /**
   * fixture-created agent에 복사되는 avatar fallback color입니다. 예: #C9A452. 직접 설정 agent는 생략될 수 있습니다.
   */
  tmp_color?: string;
  updated_at: string;
  visibility: AgentVisibility;
  work_status: AgentWorkStatus;
  /**
   * agent가 소속된 workspace id입니다. v2에서는 URL path의 workspace_id로 서버가 stamp합니다.
   */
  workspace_id: string;
}

/**
 * agent를 현재 수정할 수 있는지 나타냅니다.
 */
export type AgentEditability = "editable" | "blocked_assigned_tasks";

/**
 * agent 수정 가능 여부 변경을 전달하는 SSE event입니다.
 */
export interface AgentEditabilityChangedEvent {
  agent_id: string;
  assigned_task_count?: number;
  editability: AgentEditability;
  event_type: "agent_editability_changed";
  schema_version: string;
}

/**
 * agent 수정 가능 여부와 차단 사유를 알려주는 응답입니다.
 */
export interface AgentEditabilityResponse {
  agent_id: string;
  assigned_task_count: number;
  editability: AgentEditability;
  reason?: string;
  schema_version: string;
}

/**
 * AI Agent 온보딩에서 선택할 수 있는 서버 제공 fixture입니다. 템플릿 엔티티가 아니라 agent 생성 폼에 복사할 고정 초기값입니다.
 */
export interface AgentOnboardingFixture {
  /**
   * fixture 선택 시 생성 폼에 미리 채울 공개 범위입니다. 기본값은 안전하게 private 값을 사용합니다.
   */
  default_visibility: AgentVisibility;
  /**
   * 에이전트 설명 기본값입니다. 최대 160자입니다.
   */
  description: string;
  /**
   * fixture를 안정적으로 구분하는 ID입니다. 직접 설정 행은 포함되지 않습니다.
   */
  fixture_id: string;
  /**
   * 에이전트 지침 기본값입니다. 최대 1000자입니다.
   */
  instruction: string;
  /**
   * 에이전트 생성 폼에 복사할 기본 이름입니다.
   */
  name: string;
  /**
   * 에이전트 프로필 이미지로 복사할 HTTPS 썸네일 URL입니다.
   */
  profile_thumbnail_url?: string;
  /**
   * fixture에 권장하는 런타임 종류입니다. 선택 가능 여부는 devices.runtimes로 다시 판단합니다.
   */
  recommended_runtime_kind?: RuntimeKind;
  /**
   * fixture 목록에 보조로 표시할 역할 라벨입니다.
   */
  role_label?: string;
  /**
   * 피그마 온보딩 avatar swatch에서 온 fixture fallback color입니다. 생성된 agent에는 표시 보조값으로 복사될 수 있습니다.
   */
  tmp_color?: string;
}

/**
 * 온보딩 화면에서 사용할 서버 제공 fixture 목록 응답입니다. 직접 설정은 포함하지 않습니다.
 */
export interface AgentOnboardingFixtureListResponse {
  fixtures: AgentOnboardingFixture[];
  schema_version: string;
}

/**
 * task thread에 기록되는 AI Agent 상태 update 종류입니다. 기존 comment 필드명은 호환명입니다.
 */
export type AgentTaskCommentKind = "queued_by_busy_agent" | "assignment_started" | "stopped_by_agent_deleted" | "stopped_by_user_request" | "runtime_progress" | "task_completed" | "task_failed";

/**
 * 활성 task thread에 추가된 AI Agent 진행 상태를 client SSE로 전달하는 event입니다.
 */
export interface AgentThreadProgressEvent {
  agent_id: string;
  assignment_id?: string;
  assignment_state: AgentAssignmentState;
  batch_ended_at?: string;
  batch_started_at?: string;
  comment_kind: AgentTaskCommentKind;
  event_type: "agent_thread_progress";
  lines: AgentThreadProgressLine[];
  run_id: string;
  schema_version: string;
  task_id: string;
  thread_id: string;
  work_status: AgentWorkStatus;
}

/**
 * daemon이 batch로 보고한 runtime 진행 메시지 한 줄입니다.
 */
export interface AgentThreadProgressLine {
  message: string;
  observed_at?: string;
  seq: number;
}

/**
 * agent의 workspace 공개 범위입니다.
 */
export type AgentVisibility = "public" | "private";

/**
 * runtime output과 task assignment state에서 파생된 client 표시용 agent 작업 상태입니다.
 */
export type AgentWorkStatus = "idle" | "queued" | "running" | "waiting_for_user" | "completed" | "failed" | "offline";

/**
 * agent 작업 상태와 task thread 상태 변경을 전달하는 SSE event입니다.
 */
export interface AgentWorkStatusChangedEvent {
  agent_id: string;
  assignment_id?: string;
  assignment_state?: AgentAssignmentState;
  comment_kind?: AgentTaskCommentKind;
  event_type: "agent_work_status_changed";
  run_id?: string;
  schema_version: string;
  task_id?: string;
  thread_id?: string;
  work_status: AgentWorkStatus;
}

/**
 * task participant dropdown에서 agent를 배정하기 위한 요청입니다. 요청 body는 agent_id만 받으며 team_id/teamId/OpenAPI/Open API key는 이 계약의 입력이 아닙니다.
 */
export interface AssignAIAgentTaskRequest {
  agent_id: string;
}

/**
 * workspace task/card 목록에서 현재 배정된 AI Agent를 표시하기 위한 최소 profile 값입니다.
 */
export interface AssignedAgentProfile {
  /**
   * agent profile_thumbnail_url에서 파생된 표시용 이미지 URL입니다.
   */
  avatar_url?: string;
  /**
   * fixture-created agent에 복사된 avatar fallback color입니다. 예: 리도 #C9A452.
   */
  tmp_color?: string;
}

/**
 * workspace 안에서 현재 AI Agent가 배정된 task/subtask component_id를 key로 하는 profile 표시용 해시맵 응답입니다.
 */
export interface AssignedAgentProfileMapResponse {
  /**
   * 실제 Riido component_id/task_id 문자열을 key로 사용합니다. 예: {"23958923859":{"avatar_url":"https://...","tmp_color":"#C9A452"}}
   */
  assigned_agent_profiles: Record<string, AssignedAgentProfile>;
  schema_version: string;
  workspace_id: string;
}

/**
 * AI Agent 설정/온보딩 화면 진입 시 필요한 agent와 device runtime 초기 데이터입니다. agents[]는 settings/list 화면의 agent list이자 agent_id 출처입니다.
 */
export interface ClientBootstrapResponse {
  /**
   * AI Agent 설정/목록 화면에서 출력할 visible agent 배열입니다. 이 배열의 agent_id는 agent 수정/삭제/daemon 상세 조회 같은 settings/list action의 입력으로 사용합니다. task 참여자 드롭다운에서 배정할 agent_id는 tasks.assignableAgents.agents[]를 우선 사용합니다.
   */
  agents: AgentClientRecord[];
  client_kind: ClientKind;
  devices: DeviceRecord[];
  schema_version: string;
  workspace_id: string;
}

/**
 * v2 AI Agent 설정/온보딩 화면 진입 시 선택된 workspace의 agent와 account-owned device runtime 초기 데이터입니다. agents[]는 settings/list 화면의 workspace-visible agent list이자 agent_id 출처입니다.
 */
export interface ClientBootstrapResponseV2 {
  /**
   * 선택된 workspace의 AI Agent 설정/목록 화면에서 출력할 visible agent 배열입니다. 이 배열의 agent_id는 agent 수정/삭제/daemon 상세 조회 같은 settings/list action의 입력으로 사용합니다. task 참여자 드롭다운에서 배정할 agent_id는 v2.aiAgent.tasks.assignableAgents.agents[]를 우선 사용합니다.
   */
  agents: AgentClientRecordV2[];
  client_kind: ClientKind;
  devices: DeviceRecord[];
  schema_version: string;
  workspace_id: string;
}

/**
 * AI Agent control-plane surface를 소비하는 client 종류입니다.
 */
export type ClientKind = "web" | "desktop_webview";

/**
 * web과 desktop webview client가 소비하는 SSE event union입니다.
 */
export type ClientStreamEvent = DeviceRuntimeSnapshotEvent | DeviceDaemonStatusEvent | AgentEditabilityChangedEvent | AgentWorkStatusChangedEvent | AgentThreadProgressEvent;

/**
 * agent 권한으로 접근 가능한 daemon start/restart/stop command 요청입니다. reason은 audit 표시용이며 화면 표시 정책은 client가 결정합니다.
 */
export interface ControlDeviceDaemonRequest {
  reason?: string;
}

/**
 * task thread에 사용자의 다음 작업 지시 메시지를 남기는 정식 요청입니다. thread_id가 target agent를 결정하므로 agent_id를 받지 않습니다.
 */
export interface CreateAIAgentTaskThreadMessageRequest {
  body: string;
  source_message_id?: string;
}

/**
 * Figma agent 추가/설정 화면의 저장 요청입니다. 프로필 사진은 profile_thumbnail_url 문자열로 받으며, 이름(name), 설명(description), 런타임(runtime_id), 모델(model_id), 공개 범위(visibility), 지침(instruction)을 같은 agent 설정으로 저장합니다.
 */
export interface CreateAgentConfigurationRequest {
  /**
   * agent 추가 화면의 설명 입력값입니다.
   */
  description?: string;
  /**
   * agent 추가 화면의 지침 textarea 입력값입니다.
   */
  instruction?: string;
  /**
   * 선택 runtime의 RuntimeRecord.models[]에서 고르는 모델 dropdown 선택값입니다. 생략하면 runtime 기본 모델로 저장됩니다.
   */
  model_id?: string;
  /**
   * agent 추가 화면의 이름 입력값입니다.
   */
  name: string;
  /**
   * agent 프로필 사진 URL입니다. 이미지 업로드/저장 자체는 별도 media/storage 계약이 소유합니다.
   */
  profile_thumbnail_url?: string;
  /**
   * agent가 사용할 런타임 dropdown 선택값입니다.
   */
  runtime_id: string;
  /**
   * agent 공개 범위 radio 입력값입니다.
   */
  visibility: AgentVisibility;
}

/**
 * runtime 설정 화면에서 표시하는 desktop local daemon online/offline 상태입니다.
 */
export type DaemonAvailability = "online" | "offline";

/**
 * SaaS가 desktop local daemon에 전달할 수 있는 제어 command 종류입니다.
 */
export type DaemonControlAction = "start" | "restart" | "stop";

/**
 * daemon command가 접수된 뒤 client가 버튼/상태를 표시할 때 사용하는 제어 상태입니다.
 */
export type DaemonControlState = "idle" | "starting" | "restarting" | "stopping" | "failed";

/**
 * agent 삭제로 정리된 queued/running task 수를 포함한 응답입니다.
 */
export interface DeleteAgentResponse {
  agent_id: string;
  queued_tasks_unassigned: number;
  running_tasks_force_stopped: number;
  schema_version: string;
}

/**
 * daemon command가 SaaS에 접수된 뒤 client가 즉시 버튼과 runtime offline 상태를 갱신할 수 있도록 반환하는 응답입니다.
 */
export interface DeviceDaemonCommandResponse {
  accepted_at: string;
  action: DaemonControlAction;
  availability: DaemonAvailability;
  command_id: string;
  control_state: DaemonControlState;
  device_id: string;
  message: string;
  schema_version: string;
}

/**
 * runtime 설정 화면의 device-bound 내 기기 daemon 상세와 agent-bound 선택 Agent daemon 상세가 공유하는 응답입니다.
 */
export interface DeviceDaemonDetailResponse {
  daemon: DeviceDaemonRecord;
  schema_version: string;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세 표시와 제어 버튼 상태를 구성하는 read model입니다.
 */
export interface DeviceDaemonRecord {
  availability: DaemonAvailability;
  control_state: DaemonControlState;
  daemon_id?: string;
  device_display_name?: string;
  device_id: string;
  last_command_action?: DaemonControlAction;
  last_command_id?: string;
  last_command_requested_at?: string;
  last_seen_at?: string;
  owner_principal_id: string;
  pid?: number;
  profile?: string;
  started_at?: string;
  supported_actions: DaemonControlAction[];
  uptime_seconds?: number;
}

/**
 * agent 권한으로 접근 가능한 daemon 상세/제어 상태가 변경되었음을 client SSE로 전달하는 event입니다.
 */
export interface DeviceDaemonStatusEvent {
  daemon: DeviceDaemonRecord;
  event_type: "device_daemon_status_changed";
  schema_version: string;
}

/**
 * 데몬이 연결된 device와 해당 device에서 감지된 runtime 목록입니다.
 */
export interface DeviceRecord {
  daemon_last_seen_at?: string;
  device_id: string;
  display_name?: string;
  owner_principal_id: string;
  runtimes: RuntimeRecord[];
}

/**
 * 권한이 있는 principal이 볼 수 있는 device runtime 목록 응답입니다.
 */
export interface DeviceRuntimeListResponse {
  devices: DeviceRecord[];
  schema_version: string;
}

/**
 * device runtime 감지 결과가 갱신되었음을 알리는 SSE event입니다.
 */
export interface DeviceRuntimeSnapshotEvent {
  device: DeviceRecord;
  event_type: "device_runtime_snapshot";
  schema_version: string;
}

/**
 * control-plane API 오류 응답 envelope입니다.
 */
export interface ErrorEnvelope {
  error: string;
  schema_version: string;
}

/**
 * client에 노출되는 runtime online/offline 상태입니다.
 */
export type RuntimeAvailability = "online" | "offline";

/**
 * actor state reduction 전에 데몬이 감지한 runtime 상태입니다.
 */
export type RuntimeDetectionState = "detected" | "missing" | "error";

/**
 * device에 설치되어 device owner가 제어하는 runtime입니다.
 */
export type RuntimeKind = "codex" | "claude_code" | "cursor" | "openclaw";

/**
 * runtime이 제공하는 model 후보 record입니다. model_id는 runtime 범위 안에서만 의미가 있는 opaque identifier입니다.
 */
export interface RuntimeModelRecord {
  is_default: boolean;
  label: string;
  model_id: string;
}

/**
 * device에 설치되었거나 assignment 정책상 offline으로 유지되는 runtime record입니다.
 */
export interface RuntimeRecord {
  availability: RuntimeAvailability;
  detection_state: RuntimeDetectionState;
  device_id: string;
  has_assigned_agent: boolean;
  kind: RuntimeKind;
  last_detected_at?: string;
  models: RuntimeModelRecord[];
  owner_principal_id?: string;
  requires_experimental_opt_in: boolean;
  runtime_id: string;
}

/**
 * task thread에서 agent 작업을 중단하기 위한 요청입니다.
 */
export interface StopAIAgentTaskRequest {
  agent_id?: string;
  reason?: string;
}

/**
 * 호환 task comment route로 agent에게 메시지를 전달하기 위한 요청입니다. 정식 command는 CreateAIAgentTaskThreadMessageRequest를 사용합니다.
 */
export interface SubmitAIAgentTaskCommentRequest {
  agent_id: string;
  body: string;
  source_comment_id?: string;
}

/**
 * task participant에서 agent를 제거하며 active/queued 작업을 중단하기 위한 요청입니다.
 */
export interface UnassignAIAgentTaskRequest {
  agent_id: string;
  reason?: string;
}

/**
 * Figma agent 설정 화면에서 기존 agent의 프로필 사진 URL, 이름, 설명, 런타임, 모델, 공개 범위, 지침을 수정하기 위한 요청입니다.
 */
export interface UpdateAgentConfigurationRequest {
  /**
   * agent 설정 화면의 설명 입력값입니다.
   */
  description?: string;
  /**
   * agent 설정 화면의 지침 textarea 입력값입니다.
   */
  instruction?: string;
  /**
   * 선택 runtime의 RuntimeRecord.models[]에서 고르는 모델 dropdown 선택값입니다. 생략하면 runtime 기본 모델로 저장됩니다.
   */
  model_id?: string;
  /**
   * agent 설정 화면의 이름 입력값입니다.
   */
  name?: string;
  /**
   * agent 프로필 사진 URL입니다. 이미지 업로드/저장 자체는 별도 media/storage 계약이 소유합니다.
   */
  profile_thumbnail_url?: string;
  /**
   * agent가 사용할 런타임 dropdown 선택값입니다.
   */
  runtime_id?: string;
  /**
   * agent 공개 범위 radio 입력값입니다.
   */
  visibility?: AgentVisibility;
}

/**
 * 앱에서 사용하는 fetch 구현을 주입하기 위한 타입입니다.
 */
export type RiidoFetcher = typeof fetch;

/**
 * control-plane 호출에 필요한 기본 설정입니다.
 * `baseUrl`은 예: `https://<control-plane-host>`처럼 마지막 슬래시 없이 전달해도 됩니다.
 * `aiAgentToken`은 기존 Riido 앱 로그인 토큰과 구분되는 AI Agent SaaS 전용 토큰입니다.
 * 요청에는 `X-Riido-AI-Agent-Token` 헤더로 전달됩니다.
 * `fetcher`는 테스트나 앱 공통 transport 래핑이 필요할 때만 주입합니다.
 */
export interface RiidoClientConfig {
  baseUrl: string;
  aiAgentToken: string;
  fetcher?: RiidoFetcher;
}

/**
 * 요청 단위로 전달할 수 있는 옵션입니다. 현재는 AbortSignal만 전달합니다.
 */
export interface RiidoRequestOptions {
  signal?: AbortSignal;
}

/**
 * React Query query option에 Riido 요청 옵션을 함께 전달하기 위한 타입입니다.
 */
export type RiidoQueryOptions<TData> = Omit<UseQueryOptions<TData>, 'queryKey' | 'queryFn'> & RiidoRequestOptions;

/**
 * React Query mutation option을 Riido endpoint 변수 타입과 묶은 타입입니다.
 */
export type RiidoMutationOptions<TData, TVariables> = UseMutationOptions<TData, Error, TVariables>;

async function riidoRequest<T>(config: RiidoClientConfig, path: string, init: RequestInit = {}): Promise<T> {
  const fetcher = config.fetcher ?? fetch;
  const response = await fetcher(`${config.baseUrl.replace(/\/$/, '')}${path}`, {
    ...init,
    headers: {
      Accept: 'application/json',
      'X-Riido-AI-Agent-Token': config.aiAgentToken,
      ...(init.body ? { 'Content-Type': 'application/json' } : {}),
      ...init.headers,
    },
  });
  if (!response.ok) {
    throw new Error(`Riido API ${response.status}: ${await response.text()}`);
  }
  return response.json() as Promise<T>;
}

async function riidoRawRequest(config: RiidoClientConfig, path: string, init: RequestInit = {}): Promise<Response> {
  const fetcher = config.fetcher ?? fetch;
  const response = await fetcher(`${config.baseUrl.replace(/\/$/, '')}${path}`, {
    ...init,
    headers: {
      'X-Riido-AI-Agent-Token': config.aiAgentToken,
      ...init.headers,
    },
  });
  if (!response.ok) {
    throw new Error(`Riido API ${response.status}: ${await response.text()}`);
  }
  return response;
}

/**
 * Figma agent 추가 화면에서 AI agent 설정을 생성합니다
 */
export async function createAIAgent(config: RiidoClientConfig, body: CreateAgentConfigurationRequest, options: RiidoRequestOptions = {}): Promise<AgentClientRecordResponse> {
  const path = "/v1/client/ai-agent/agents";
  return riidoRequest<AgentClientRecordResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * Figma agent 추가 화면에서 AI agent 설정을 생성합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface CreateAIAgentMutationVariables {
  body: CreateAgentConfigurationRequest;
}

/**
 * Figma agent 추가 화면에서 AI agent 설정을 생성합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function createAIAgentMutationKey(): readonly unknown[] {
  return ["createAIAgent"] as const;
}

/**
 * Figma agent 추가 화면에서 AI agent 설정을 생성합니다
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function createAIAgentMutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AgentClientRecordResponse, CreateAIAgentMutationVariables> = {}) {
  return {
    ...options,
    mutationKey: createAIAgentMutationKey(),
    mutationFn: (variables: CreateAIAgentMutationVariables) => createAIAgent(config, variables.body, {}),
  };
}

/**
 * agent를 삭제하고 queued/running assignment를 강제로 정리합니다
 * 경로 파라미터입니다.
 */
export interface DeleteAIAgentPathParams {
  agent_id: string;
}

/**
 * agent를 삭제하고 queued/running assignment를 강제로 정리합니다
 */
export async function deleteAIAgent(config: RiidoClientConfig, params: DeleteAIAgentPathParams, options: RiidoRequestOptions = {}): Promise<DeleteAgentResponse> {
  const path = `/v1/client/ai-agent/agents/${params.agent_id}`;
  return riidoRequest<DeleteAgentResponse>(config, path, { method: 'DELETE', signal: options.signal });
}

/**
 * agent를 삭제하고 queued/running assignment를 강제로 정리합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface DeleteAIAgentMutationVariables {
  params: DeleteAIAgentPathParams;
}

/**
 * agent를 삭제하고 queued/running assignment를 강제로 정리합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function deleteAIAgentMutationKey(): readonly unknown[] {
  return ["deleteAIAgent"] as const;
}

/**
 * agent를 삭제하고 queued/running assignment를 강제로 정리합니다
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function deleteAIAgentMutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<DeleteAgentResponse, DeleteAIAgentMutationVariables> = {}) {
  return {
    ...options,
    mutationKey: deleteAIAgentMutationKey(),
    mutationFn: (variables: DeleteAIAgentMutationVariables) => deleteAIAgent(config, variables.params, {}),
  };
}

/**
 * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다
 * 경로 파라미터입니다.
 */
export interface UpdateAIAgentConfigurationPathParams {
  agent_id: string;
}

/**
 * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다
 */
export async function updateAIAgentConfiguration(config: RiidoClientConfig, params: UpdateAIAgentConfigurationPathParams, body: UpdateAgentConfigurationRequest, options: RiidoRequestOptions = {}): Promise<AgentClientRecordResponse> {
  const path = `/v1/client/ai-agent/agents/${params.agent_id}`;
  return riidoRequest<AgentClientRecordResponse>(config, path, { method: 'PATCH', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface UpdateAIAgentConfigurationMutationVariables {
  params: UpdateAIAgentConfigurationPathParams;
  body: UpdateAgentConfigurationRequest;
}

/**
 * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function updateAIAgentConfigurationMutationKey(): readonly unknown[] {
  return ["updateAIAgentConfiguration"] as const;
}

/**
 * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function updateAIAgentConfigurationMutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AgentClientRecordResponse, UpdateAIAgentConfigurationMutationVariables> = {}) {
  return {
    ...options,
    mutationKey: updateAIAgentConfigurationMutationKey(),
    mutationFn: (variables: UpdateAIAgentConfigurationMutationVariables) => updateAIAgentConfiguration(config, variables.params, variables.body, {}),
  };
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다
 * 경로 파라미터입니다.
 */
export interface GetAIAgentDaemonPathParams {
  agent_id: string;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다
 */
export async function getAIAgentDaemon(config: RiidoClientConfig, params: GetAIAgentDaemonPathParams, options: RiidoRequestOptions = {}): Promise<DeviceDaemonDetailResponse> {
  const path = `/v1/client/ai-agent/agents/${params.agent_id}/daemon`;
  return riidoRequest<DeviceDaemonDetailResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다
 * cache tag: `aiAgent.agents.daemon`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function getAIAgentDaemonQueryKeyRoot(): readonly unknown[] {
  return ["aiAgent.agents.daemon"] as const;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function getAIAgentDaemonQueryKey(params: GetAIAgentDaemonPathParams): readonly unknown[] {
  return [...getAIAgentDaemonQueryKeyRoot(), params] as const;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function getAIAgentDaemonQueryOptions(config: RiidoClientConfig, params: GetAIAgentDaemonPathParams, options: RiidoQueryOptions<DeviceDaemonDetailResponse> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: getAIAgentDaemonQueryKey(params),
    queryFn: () => getAIAgentDaemon(config, params, { signal }),
  };
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다
 * 경로 파라미터입니다.
 */
export interface RestartAIAgentDaemonPathParams {
  agent_id: string;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다
 */
export async function restartAIAgentDaemon(config: RiidoClientConfig, params: RestartAIAgentDaemonPathParams, body: ControlDeviceDaemonRequest, options: RiidoRequestOptions = {}): Promise<DeviceDaemonCommandResponse> {
  const path = `/v1/client/ai-agent/agents/${params.agent_id}/daemon/restart`;
  return riidoRequest<DeviceDaemonCommandResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface RestartAIAgentDaemonMutationVariables {
  params: RestartAIAgentDaemonPathParams;
  body: ControlDeviceDaemonRequest;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function restartAIAgentDaemonMutationKey(): readonly unknown[] {
  return ["restartAIAgentDaemon"] as const;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function restartAIAgentDaemonMutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDaemonMutationVariables> = {}) {
  return {
    ...options,
    mutationKey: restartAIAgentDaemonMutationKey(),
    mutationFn: (variables: RestartAIAgentDaemonMutationVariables) => restartAIAgentDaemon(config, variables.params, variables.body, {}),
  };
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다
 * 경로 파라미터입니다.
 */
export interface StartAIAgentDaemonPathParams {
  agent_id: string;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다
 */
export async function startAIAgentDaemon(config: RiidoClientConfig, params: StartAIAgentDaemonPathParams, body: ControlDeviceDaemonRequest, options: RiidoRequestOptions = {}): Promise<DeviceDaemonCommandResponse> {
  const path = `/v1/client/ai-agent/agents/${params.agent_id}/daemon/start`;
  return riidoRequest<DeviceDaemonCommandResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface StartAIAgentDaemonMutationVariables {
  params: StartAIAgentDaemonPathParams;
  body: ControlDeviceDaemonRequest;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function startAIAgentDaemonMutationKey(): readonly unknown[] {
  return ["startAIAgentDaemon"] as const;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function startAIAgentDaemonMutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDaemonMutationVariables> = {}) {
  return {
    ...options,
    mutationKey: startAIAgentDaemonMutationKey(),
    mutationFn: (variables: StartAIAgentDaemonMutationVariables) => startAIAgentDaemon(config, variables.params, variables.body, {}),
  };
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다
 * 경로 파라미터입니다.
 */
export interface StopAIAgentDaemonPathParams {
  agent_id: string;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다
 */
export async function stopAIAgentDaemon(config: RiidoClientConfig, params: StopAIAgentDaemonPathParams, body: ControlDeviceDaemonRequest, options: RiidoRequestOptions = {}): Promise<DeviceDaemonCommandResponse> {
  const path = `/v1/client/ai-agent/agents/${params.agent_id}/daemon/stop`;
  return riidoRequest<DeviceDaemonCommandResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface StopAIAgentDaemonMutationVariables {
  params: StopAIAgentDaemonPathParams;
  body: ControlDeviceDaemonRequest;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function stopAIAgentDaemonMutationKey(): readonly unknown[] {
  return ["stopAIAgentDaemon"] as const;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function stopAIAgentDaemonMutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDaemonMutationVariables> = {}) {
  return {
    ...options,
    mutationKey: stopAIAgentDaemonMutationKey(),
    mutationFn: (variables: StopAIAgentDaemonMutationVariables) => stopAIAgentDaemon(config, variables.params, variables.body, {}),
  };
}

/**
 * client control 활성화 전에 agent 수정 가능 여부를 조회합니다
 * 경로 파라미터입니다.
 */
export interface GetAIAgentEditabilityPathParams {
  agent_id: string;
}

/**
 * client control 활성화 전에 agent 수정 가능 여부를 조회합니다
 */
export async function getAIAgentEditability(config: RiidoClientConfig, params: GetAIAgentEditabilityPathParams, options: RiidoRequestOptions = {}): Promise<AgentEditabilityResponse> {
  const path = `/v1/client/ai-agent/agents/${params.agent_id}/editability`;
  return riidoRequest<AgentEditabilityResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * client control 활성화 전에 agent 수정 가능 여부를 조회합니다
 * cache tag: `aiAgent.agents.editability`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function getAIAgentEditabilityQueryKeyRoot(): readonly unknown[] {
  return ["aiAgent.agents.editability"] as const;
}

/**
 * client control 활성화 전에 agent 수정 가능 여부를 조회합니다
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function getAIAgentEditabilityQueryKey(params: GetAIAgentEditabilityPathParams): readonly unknown[] {
  return [...getAIAgentEditabilityQueryKeyRoot(), params] as const;
}

/**
 * client control 활성화 전에 agent 수정 가능 여부를 조회합니다
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function getAIAgentEditabilityQueryOptions(config: RiidoClientConfig, params: GetAIAgentEditabilityPathParams, options: RiidoQueryOptions<AgentEditabilityResponse> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: getAIAgentEditabilityQueryKey(params),
    queryFn: () => getAIAgentEditability(config, params, { signal }),
  };
}

/**
 * web 또는 desktop webview client의 AI Agent 설정/온보딩 초기 데이터와 bootstrap.agents[] agent list를 조회합니다
 */
export async function getAIAgentClientBootstrap(config: RiidoClientConfig, options: RiidoRequestOptions = {}): Promise<ClientBootstrapResponse> {
  const path = "/v1/client/ai-agent/bootstrap";
  return riidoRequest<ClientBootstrapResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * web 또는 desktop webview client의 AI Agent 설정/온보딩 초기 데이터와 bootstrap.agents[] agent list를 조회합니다
 * cache tag: `aiAgent.bootstrap`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function getAIAgentClientBootstrapQueryKeyRoot(): readonly unknown[] {
  return ["aiAgent.bootstrap"] as const;
}

/**
 * web 또는 desktop webview client의 AI Agent 설정/온보딩 초기 데이터와 bootstrap.agents[] agent list를 조회합니다
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function getAIAgentClientBootstrapQueryKey(): readonly unknown[] {
  return [...getAIAgentClientBootstrapQueryKeyRoot()] as const;
}

/**
 * web 또는 desktop webview client의 AI Agent 설정/온보딩 초기 데이터와 bootstrap.agents[] agent list를 조회합니다
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function getAIAgentClientBootstrapQueryOptions(config: RiidoClientConfig, options: RiidoQueryOptions<ClientBootstrapResponse> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: getAIAgentClientBootstrapQueryKey(),
    queryFn: () => getAIAgentClientBootstrap(config, { signal }),
  };
}

/**
 * 권한이 확인된 principal의 device runtime 상태를 조회합니다
 */
export async function listAIAgentDeviceRuntimes(config: RiidoClientConfig, options: RiidoRequestOptions = {}): Promise<DeviceRuntimeListResponse> {
  const path = "/v1/client/ai-agent/devices";
  return riidoRequest<DeviceRuntimeListResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * 권한이 확인된 principal의 device runtime 상태를 조회합니다
 * cache tag: `aiAgent.devices.runtimes`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function listAIAgentDeviceRuntimesQueryKeyRoot(): readonly unknown[] {
  return ["aiAgent.devices.runtimes"] as const;
}

/**
 * 권한이 확인된 principal의 device runtime 상태를 조회합니다
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function listAIAgentDeviceRuntimesQueryKey(): readonly unknown[] {
  return [...listAIAgentDeviceRuntimesQueryKeyRoot()] as const;
}

/**
 * 권한이 확인된 principal의 device runtime 상태를 조회합니다
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function listAIAgentDeviceRuntimesQueryOptions(config: RiidoClientConfig, options: RiidoQueryOptions<DeviceRuntimeListResponse> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: listAIAgentDeviceRuntimesQueryKey(),
    queryFn: () => listAIAgentDeviceRuntimes(config, { signal }),
  };
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 상세를 조회합니다
 * 경로 파라미터입니다.
 */
export interface GetAIAgentDeviceDaemonPathParams {
  device_id: string;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 상세를 조회합니다
 */
export async function getAIAgentDeviceDaemon(config: RiidoClientConfig, params: GetAIAgentDeviceDaemonPathParams, options: RiidoRequestOptions = {}): Promise<DeviceDaemonDetailResponse> {
  const path = `/v1/client/ai-agent/devices/${params.device_id}/daemon`;
  return riidoRequest<DeviceDaemonDetailResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 상세를 조회합니다
 * cache tag: `aiAgent.devices.daemon`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function getAIAgentDeviceDaemonQueryKeyRoot(): readonly unknown[] {
  return ["aiAgent.devices.daemon"] as const;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 상세를 조회합니다
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function getAIAgentDeviceDaemonQueryKey(params: GetAIAgentDeviceDaemonPathParams): readonly unknown[] {
  return [...getAIAgentDeviceDaemonQueryKeyRoot(), params] as const;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 상세를 조회합니다
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function getAIAgentDeviceDaemonQueryOptions(config: RiidoClientConfig, params: GetAIAgentDeviceDaemonPathParams, options: RiidoQueryOptions<DeviceDaemonDetailResponse> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: getAIAgentDeviceDaemonQueryKey(params),
    queryFn: () => getAIAgentDeviceDaemon(config, params, { signal }),
  };
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 재시작을 요청합니다
 * 경로 파라미터입니다.
 */
export interface RestartAIAgentDeviceDaemonPathParams {
  device_id: string;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 재시작을 요청합니다
 */
export async function restartAIAgentDeviceDaemon(config: RiidoClientConfig, params: RestartAIAgentDeviceDaemonPathParams, body: ControlDeviceDaemonRequest, options: RiidoRequestOptions = {}): Promise<DeviceDaemonCommandResponse> {
  const path = `/v1/client/ai-agent/devices/${params.device_id}/daemon/restart`;
  return riidoRequest<DeviceDaemonCommandResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 재시작을 요청합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface RestartAIAgentDeviceDaemonMutationVariables {
  params: RestartAIAgentDeviceDaemonPathParams;
  body: ControlDeviceDaemonRequest;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 재시작을 요청합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function restartAIAgentDeviceDaemonMutationKey(): readonly unknown[] {
  return ["restartAIAgentDeviceDaemon"] as const;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 재시작을 요청합니다
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function restartAIAgentDeviceDaemonMutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDeviceDaemonMutationVariables> = {}) {
  return {
    ...options,
    mutationKey: restartAIAgentDeviceDaemonMutationKey(),
    mutationFn: (variables: RestartAIAgentDeviceDaemonMutationVariables) => restartAIAgentDeviceDaemon(config, variables.params, variables.body, {}),
  };
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 시작을 요청합니다
 * 경로 파라미터입니다.
 */
export interface StartAIAgentDeviceDaemonPathParams {
  device_id: string;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 시작을 요청합니다
 */
export async function startAIAgentDeviceDaemon(config: RiidoClientConfig, params: StartAIAgentDeviceDaemonPathParams, body: ControlDeviceDaemonRequest, options: RiidoRequestOptions = {}): Promise<DeviceDaemonCommandResponse> {
  const path = `/v1/client/ai-agent/devices/${params.device_id}/daemon/start`;
  return riidoRequest<DeviceDaemonCommandResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 시작을 요청합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface StartAIAgentDeviceDaemonMutationVariables {
  params: StartAIAgentDeviceDaemonPathParams;
  body: ControlDeviceDaemonRequest;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 시작을 요청합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function startAIAgentDeviceDaemonMutationKey(): readonly unknown[] {
  return ["startAIAgentDeviceDaemon"] as const;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 시작을 요청합니다
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function startAIAgentDeviceDaemonMutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDeviceDaemonMutationVariables> = {}) {
  return {
    ...options,
    mutationKey: startAIAgentDeviceDaemonMutationKey(),
    mutationFn: (variables: StartAIAgentDeviceDaemonMutationVariables) => startAIAgentDeviceDaemon(config, variables.params, variables.body, {}),
  };
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 중지를 요청합니다
 * 경로 파라미터입니다.
 */
export interface StopAIAgentDeviceDaemonPathParams {
  device_id: string;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 중지를 요청합니다
 */
export async function stopAIAgentDeviceDaemon(config: RiidoClientConfig, params: StopAIAgentDeviceDaemonPathParams, body: ControlDeviceDaemonRequest, options: RiidoRequestOptions = {}): Promise<DeviceDaemonCommandResponse> {
  const path = `/v1/client/ai-agent/devices/${params.device_id}/daemon/stop`;
  return riidoRequest<DeviceDaemonCommandResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 중지를 요청합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface StopAIAgentDeviceDaemonMutationVariables {
  params: StopAIAgentDeviceDaemonPathParams;
  body: ControlDeviceDaemonRequest;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 중지를 요청합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function stopAIAgentDeviceDaemonMutationKey(): readonly unknown[] {
  return ["stopAIAgentDeviceDaemon"] as const;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 중지를 요청합니다
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function stopAIAgentDeviceDaemonMutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDeviceDaemonMutationVariables> = {}) {
  return {
    ...options,
    mutationKey: stopAIAgentDeviceDaemonMutationKey(),
    mutationFn: (variables: StopAIAgentDeviceDaemonMutationVariables) => stopAIAgentDeviceDaemon(config, variables.params, variables.body, {}),
  };
}

/**
 * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다
 */
export async function streamAIAgentClientEvents(config: RiidoClientConfig, options: RiidoRequestOptions = {}): Promise<Response> {
  const path = "/v1/client/ai-agent/events";
  return riidoRawRequest(config, path, { method: 'GET', signal: options.signal });
}

/**
 * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다
 * cache tag: `aiAgent.events.stream`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function streamAIAgentClientEventsQueryKeyRoot(): readonly unknown[] {
  return ["aiAgent.events.stream"] as const;
}

/**
 * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function streamAIAgentClientEventsQueryKey(): readonly unknown[] {
  return [...streamAIAgentClientEventsQueryKeyRoot()] as const;
}

/**
 * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function streamAIAgentClientEventsQueryOptions(config: RiidoClientConfig, options: RiidoQueryOptions<Response> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: streamAIAgentClientEventsQueryKey(),
    queryFn: () => streamAIAgentClientEvents(config, { signal }),
  };
}

/**
 * AI Agent 온보딩에서 사용할 서버 제공 fixture 목록을 조회합니다
 */
export async function listAIAgentOnboardingFixtures(config: RiidoClientConfig, options: RiidoRequestOptions = {}): Promise<AgentOnboardingFixtureListResponse> {
  const path = "/v1/client/ai-agent/onboarding/fixtures";
  return riidoRequest<AgentOnboardingFixtureListResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * AI Agent 온보딩에서 사용할 서버 제공 fixture 목록을 조회합니다
 * cache tag: `aiAgent.onboarding.fixtures`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function listAIAgentOnboardingFixturesQueryKeyRoot(): readonly unknown[] {
  return ["aiAgent.onboarding.fixtures"] as const;
}

/**
 * AI Agent 온보딩에서 사용할 서버 제공 fixture 목록을 조회합니다
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function listAIAgentOnboardingFixturesQueryKey(): readonly unknown[] {
  return [...listAIAgentOnboardingFixturesQueryKeyRoot()] as const;
}

/**
 * AI Agent 온보딩에서 사용할 서버 제공 fixture 목록을 조회합니다
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function listAIAgentOnboardingFixturesQueryOptions(config: RiidoClientConfig, options: RiidoQueryOptions<AgentOnboardingFixtureListResponse> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: listAIAgentOnboardingFixturesQueryKey(),
    queryFn: () => listAIAgentOnboardingFixtures(config, { signal }),
  };
}

/**
 * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다
 * 경로 파라미터입니다.
 */
export interface CreateAIAgentFromOnboardingFixturePathParams {
  fixture_id: string;
}

/**
 * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다
 */
export async function createAIAgentFromOnboardingFixture(config: RiidoClientConfig, params: CreateAIAgentFromOnboardingFixturePathParams, body: CreateAgentConfigurationRequest, options: RiidoRequestOptions = {}): Promise<AgentClientRecordResponse> {
  const path = `/v1/client/ai-agent/onboarding/fixtures/${params.fixture_id}/agents`;
  return riidoRequest<AgentClientRecordResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface CreateAIAgentFromOnboardingFixtureMutationVariables {
  params: CreateAIAgentFromOnboardingFixturePathParams;
  body: CreateAgentConfigurationRequest;
}

/**
 * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function createAIAgentFromOnboardingFixtureMutationKey(): readonly unknown[] {
  return ["createAIAgentFromOnboardingFixture"] as const;
}

/**
 * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function createAIAgentFromOnboardingFixtureMutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AgentClientRecordResponse, CreateAIAgentFromOnboardingFixtureMutationVariables> = {}) {
  return {
    ...options,
    mutationKey: createAIAgentFromOnboardingFixtureMutationKey(),
    mutationFn: (variables: CreateAIAgentFromOnboardingFixtureMutationVariables) => createAIAgentFromOnboardingFixture(config, variables.params, variables.body, {}),
  };
}

/**
 * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다
 * 경로 파라미터입니다.
 */
export interface ListAIAgentTaskAssignableAgentsPathParams {
  task_id: string;
}

/**
 * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다
 */
export async function listAIAgentTaskAssignableAgents(config: RiidoClientConfig, params: ListAIAgentTaskAssignableAgentsPathParams, options: RiidoRequestOptions = {}): Promise<AgentClientListResponse> {
  const path = `/v1/client/ai-agent/tasks/${params.task_id}/assignable-agents`;
  return riidoRequest<AgentClientListResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다
 * cache tag: `aiAgent.tasks.assignableAgents`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function listAIAgentTaskAssignableAgentsQueryKeyRoot(): readonly unknown[] {
  return ["aiAgent.tasks.assignableAgents"] as const;
}

/**
 * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function listAIAgentTaskAssignableAgentsQueryKey(params: ListAIAgentTaskAssignableAgentsPathParams): readonly unknown[] {
  return [...listAIAgentTaskAssignableAgentsQueryKeyRoot(), params] as const;
}

/**
 * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function listAIAgentTaskAssignableAgentsQueryOptions(config: RiidoClientConfig, params: ListAIAgentTaskAssignableAgentsPathParams, options: RiidoQueryOptions<AgentClientListResponse> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: listAIAgentTaskAssignableAgentsQueryKey(params),
    queryFn: () => listAIAgentTaskAssignableAgents(config, params, { signal }),
  };
}

/**
 * task participant에서 AI agent를 제거하고 작업을 중단합니다
 * 경로 파라미터입니다.
 */
export interface UnassignAIAgentTaskPathParams {
  task_id: string;
}

/**
 * task participant에서 AI agent를 제거하고 작업을 중단합니다
 */
export async function unassignAIAgentTask(config: RiidoClientConfig, params: UnassignAIAgentTaskPathParams, body: UnassignAIAgentTaskRequest, options: RiidoRequestOptions = {}): Promise<AIAgentTaskActionResponse> {
  const path = `/v1/client/ai-agent/tasks/${params.task_id}/assignment`;
  return riidoRequest<AIAgentTaskActionResponse>(config, path, { method: 'DELETE', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * task participant에서 AI agent를 제거하고 작업을 중단합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface UnassignAIAgentTaskMutationVariables {
  params: UnassignAIAgentTaskPathParams;
  body: UnassignAIAgentTaskRequest;
}

/**
 * task participant에서 AI agent를 제거하고 작업을 중단합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function unassignAIAgentTaskMutationKey(): readonly unknown[] {
  return ["unassignAIAgentTask"] as const;
}

/**
 * task participant에서 AI agent를 제거하고 작업을 중단합니다
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function unassignAIAgentTaskMutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AIAgentTaskActionResponse, UnassignAIAgentTaskMutationVariables> = {}) {
  return {
    ...options,
    mutationKey: unassignAIAgentTaskMutationKey(),
    mutationFn: (variables: UnassignAIAgentTaskMutationVariables) => unassignAIAgentTask(config, variables.params, variables.body, {}),
  };
}

/**
 * task participant dropdown에서 AI agent를 배정합니다
 * 경로 파라미터입니다.
 */
export interface AssignAIAgentTaskPathParams {
  task_id: string;
}

/**
 * task participant dropdown에서 AI agent를 배정합니다
 */
export async function assignAIAgentTask(config: RiidoClientConfig, params: AssignAIAgentTaskPathParams, body: AssignAIAgentTaskRequest, options: RiidoRequestOptions = {}): Promise<AIAgentTaskActionResponse> {
  const path = `/v1/client/ai-agent/tasks/${params.task_id}/assignment`;
  return riidoRequest<AIAgentTaskActionResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * task participant dropdown에서 AI agent를 배정합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface AssignAIAgentTaskMutationVariables {
  params: AssignAIAgentTaskPathParams;
  body: AssignAIAgentTaskRequest;
}

/**
 * task participant dropdown에서 AI agent를 배정합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function assignAIAgentTaskMutationKey(): readonly unknown[] {
  return ["assignAIAgentTask"] as const;
}

/**
 * task participant dropdown에서 AI agent를 배정합니다
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function assignAIAgentTaskMutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AIAgentTaskActionResponse, AssignAIAgentTaskMutationVariables> = {}) {
  return {
    ...options,
    mutationKey: assignAIAgentTaskMutationKey(),
    mutationFn: (variables: AssignAIAgentTaskMutationVariables) => assignAIAgentTask(config, variables.params, variables.body, {}),
  };
}

/**
 * 호환 task comment route로 AI agent에게 메시지를 전달합니다
 * 경로 파라미터입니다.
 */
export interface SubmitAIAgentTaskCommentPathParams {
  task_id: string;
}

/**
 * 호환 task comment route로 AI agent에게 메시지를 전달합니다
 */
export async function submitAIAgentTaskComment(config: RiidoClientConfig, params: SubmitAIAgentTaskCommentPathParams, body: SubmitAIAgentTaskCommentRequest, options: RiidoRequestOptions = {}): Promise<AIAgentTaskActionResponse> {
  const path = `/v1/client/ai-agent/tasks/${params.task_id}/comments`;
  return riidoRequest<AIAgentTaskActionResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * 호환 task comment route로 AI agent에게 메시지를 전달합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface SubmitAIAgentTaskCommentMutationVariables {
  params: SubmitAIAgentTaskCommentPathParams;
  body: SubmitAIAgentTaskCommentRequest;
}

/**
 * 호환 task comment route로 AI agent에게 메시지를 전달합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function submitAIAgentTaskCommentMutationKey(): readonly unknown[] {
  return ["submitAIAgentTaskComment"] as const;
}

/**
 * 호환 task comment route로 AI agent에게 메시지를 전달합니다
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function submitAIAgentTaskCommentMutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AIAgentTaskActionResponse, SubmitAIAgentTaskCommentMutationVariables> = {}) {
  return {
    ...options,
    mutationKey: submitAIAgentTaskCommentMutationKey(),
    mutationFn: (variables: SubmitAIAgentTaskCommentMutationVariables) => submitAIAgentTaskComment(config, variables.params, variables.body, {}),
  };
}

/**
 * task thread의 stop action으로 AI agent 작업을 중단합니다
 * 경로 파라미터입니다.
 */
export interface StopAIAgentTaskPathParams {
  task_id: string;
}

/**
 * task thread의 stop action으로 AI agent 작업을 중단합니다
 */
export async function stopAIAgentTask(config: RiidoClientConfig, params: StopAIAgentTaskPathParams, body: StopAIAgentTaskRequest, options: RiidoRequestOptions = {}): Promise<AIAgentTaskActionResponse> {
  const path = `/v1/client/ai-agent/tasks/${params.task_id}/stop`;
  return riidoRequest<AIAgentTaskActionResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * task thread의 stop action으로 AI agent 작업을 중단합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface StopAIAgentTaskMutationVariables {
  params: StopAIAgentTaskPathParams;
  body: StopAIAgentTaskRequest;
}

/**
 * task thread의 stop action으로 AI agent 작업을 중단합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function stopAIAgentTaskMutationKey(): readonly unknown[] {
  return ["stopAIAgentTask"] as const;
}

/**
 * task thread의 stop action으로 AI agent 작업을 중단합니다
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function stopAIAgentTaskMutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AIAgentTaskActionResponse, StopAIAgentTaskMutationVariables> = {}) {
  return {
    ...options,
    mutationKey: stopAIAgentTaskMutationKey(),
    mutationFn: (variables: StopAIAgentTaskMutationVariables) => stopAIAgentTask(config, variables.params, variables.body, {}),
  };
}

/**
 * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다
 * 경로 파라미터입니다.
 */
export interface ListAIAgentTaskThreadsPathParams {
  task_id: string;
}

/**
 * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다
 */
export async function listAIAgentTaskThreads(config: RiidoClientConfig, params: ListAIAgentTaskThreadsPathParams, options: RiidoRequestOptions = {}): Promise<AIAgentTaskThreadCollectionResponse> {
  const path = `/v1/client/ai-agent/tasks/${params.task_id}/threads`;
  return riidoRequest<AIAgentTaskThreadCollectionResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다
 * cache tag: `aiAgent.tasks.threads`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function listAIAgentTaskThreadsQueryKeyRoot(): readonly unknown[] {
  return ["aiAgent.tasks.threads"] as const;
}

/**
 * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function listAIAgentTaskThreadsQueryKey(params: ListAIAgentTaskThreadsPathParams): readonly unknown[] {
  return [...listAIAgentTaskThreadsQueryKeyRoot(), params] as const;
}

/**
 * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function listAIAgentTaskThreadsQueryOptions(config: RiidoClientConfig, params: ListAIAgentTaskThreadsPathParams, options: RiidoQueryOptions<AIAgentTaskThreadCollectionResponse> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: listAIAgentTaskThreadsQueryKey(params),
    queryFn: () => listAIAgentTaskThreads(config, params, { signal }),
  };
}

/**
 * task thread message로 AI agent에게 다음 작업 지시를 전달합니다
 * 경로 파라미터입니다.
 */
export interface CreateAIAgentTaskThreadMessagePathParams {
  task_id: string;
  thread_id: string;
}

/**
 * task thread message로 AI agent에게 다음 작업 지시를 전달합니다
 */
export async function createAIAgentTaskThreadMessage(config: RiidoClientConfig, params: CreateAIAgentTaskThreadMessagePathParams, body: CreateAIAgentTaskThreadMessageRequest, options: RiidoRequestOptions = {}): Promise<AIAgentTaskActionResponse> {
  const path = `/v1/client/ai-agent/tasks/${params.task_id}/threads/${params.thread_id}/messages`;
  return riidoRequest<AIAgentTaskActionResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * task thread message로 AI agent에게 다음 작업 지시를 전달합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface CreateAIAgentTaskThreadMessageMutationVariables {
  params: CreateAIAgentTaskThreadMessagePathParams;
  body: CreateAIAgentTaskThreadMessageRequest;
}

/**
 * task thread message로 AI agent에게 다음 작업 지시를 전달합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function createAIAgentTaskThreadMessageMutationKey(): readonly unknown[] {
  return ["createAIAgentTaskThreadMessage"] as const;
}

/**
 * task thread message로 AI agent에게 다음 작업 지시를 전달합니다
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function createAIAgentTaskThreadMessageMutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AIAgentTaskActionResponse, CreateAIAgentTaskThreadMessageMutationVariables> = {}) {
  return {
    ...options,
    mutationKey: createAIAgentTaskThreadMessageMutationKey(),
    mutationFn: (variables: CreateAIAgentTaskThreadMessageMutationVariables) => createAIAgentTaskThreadMessage(config, variables.params, variables.body, {}),
  };
}

/**
 * Figma agent 추가 화면에서 AI agent 설정을 생성합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface CreateAIAgentV2PathParams {
  workspace_id: string;
}

/**
 * Figma agent 추가 화면에서 AI agent 설정을 생성합니다 (v2 workspace-scoped)
 */
export async function createAIAgentV2(config: RiidoClientConfig, params: CreateAIAgentV2PathParams, body: CreateAgentConfigurationRequest, options: RiidoRequestOptions = {}): Promise<AgentClientRecordResponseV2> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/agents`;
  return riidoRequest<AgentClientRecordResponseV2>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * Figma agent 추가 화면에서 AI agent 설정을 생성합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface CreateAIAgentV2MutationVariables {
  params: CreateAIAgentV2PathParams;
  body: CreateAgentConfigurationRequest;
}

/**
 * Figma agent 추가 화면에서 AI agent 설정을 생성합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function createAIAgentV2MutationKey(): readonly unknown[] {
  return ["createAIAgentV2"] as const;
}

/**
 * Figma agent 추가 화면에서 AI agent 설정을 생성합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function createAIAgentV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AgentClientRecordResponseV2, CreateAIAgentV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: createAIAgentV2MutationKey(),
    mutationFn: (variables: CreateAIAgentV2MutationVariables) => createAIAgentV2(config, variables.params, variables.body, {}),
  };
}

/**
 * agent를 삭제하고 queued/running assignment를 강제로 정리합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface DeleteAIAgentV2PathParams {
  workspace_id: string;
  agent_id: string;
}

/**
 * agent를 삭제하고 queued/running assignment를 강제로 정리합니다 (v2 workspace-scoped)
 */
export async function deleteAIAgentV2(config: RiidoClientConfig, params: DeleteAIAgentV2PathParams, options: RiidoRequestOptions = {}): Promise<DeleteAgentResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/agents/${params.agent_id}`;
  return riidoRequest<DeleteAgentResponse>(config, path, { method: 'DELETE', signal: options.signal });
}

/**
 * agent를 삭제하고 queued/running assignment를 강제로 정리합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface DeleteAIAgentV2MutationVariables {
  params: DeleteAIAgentV2PathParams;
}

/**
 * agent를 삭제하고 queued/running assignment를 강제로 정리합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function deleteAIAgentV2MutationKey(): readonly unknown[] {
  return ["deleteAIAgentV2"] as const;
}

/**
 * agent를 삭제하고 queued/running assignment를 강제로 정리합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function deleteAIAgentV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<DeleteAgentResponse, DeleteAIAgentV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: deleteAIAgentV2MutationKey(),
    mutationFn: (variables: DeleteAIAgentV2MutationVariables) => deleteAIAgentV2(config, variables.params, {}),
  };
}

/**
 * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface UpdateAIAgentConfigurationV2PathParams {
  workspace_id: string;
  agent_id: string;
}

/**
 * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다 (v2 workspace-scoped)
 */
export async function updateAIAgentConfigurationV2(config: RiidoClientConfig, params: UpdateAIAgentConfigurationV2PathParams, body: UpdateAgentConfigurationRequest, options: RiidoRequestOptions = {}): Promise<AgentClientRecordResponseV2> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/agents/${params.agent_id}`;
  return riidoRequest<AgentClientRecordResponseV2>(config, path, { method: 'PATCH', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface UpdateAIAgentConfigurationV2MutationVariables {
  params: UpdateAIAgentConfigurationV2PathParams;
  body: UpdateAgentConfigurationRequest;
}

/**
 * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function updateAIAgentConfigurationV2MutationKey(): readonly unknown[] {
  return ["updateAIAgentConfigurationV2"] as const;
}

/**
 * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function updateAIAgentConfigurationV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AgentClientRecordResponseV2, UpdateAIAgentConfigurationV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: updateAIAgentConfigurationV2MutationKey(),
    mutationFn: (variables: UpdateAIAgentConfigurationV2MutationVariables) => updateAIAgentConfigurationV2(config, variables.params, variables.body, {}),
  };
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface GetAIAgentDaemonV2PathParams {
  workspace_id: string;
  agent_id: string;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다 (v2 workspace-scoped)
 */
export async function getAIAgentDaemonV2(config: RiidoClientConfig, params: GetAIAgentDaemonV2PathParams, options: RiidoRequestOptions = {}): Promise<DeviceDaemonDetailResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/agents/${params.agent_id}/daemon`;
  return riidoRequest<DeviceDaemonDetailResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다 (v2 workspace-scoped)
 * cache tag: `v2.aiAgent.agents.daemon`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function getAIAgentDaemonV2QueryKeyRoot(): readonly unknown[] {
  return ["v2.aiAgent.agents.daemon"] as const;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다 (v2 workspace-scoped)
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function getAIAgentDaemonV2QueryKey(params: GetAIAgentDaemonV2PathParams): readonly unknown[] {
  return [...getAIAgentDaemonV2QueryKeyRoot(), params] as const;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다 (v2 workspace-scoped)
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function getAIAgentDaemonV2QueryOptions(config: RiidoClientConfig, params: GetAIAgentDaemonV2PathParams, options: RiidoQueryOptions<DeviceDaemonDetailResponse> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: getAIAgentDaemonV2QueryKey(params),
    queryFn: () => getAIAgentDaemonV2(config, params, { signal }),
  };
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface RestartAIAgentDaemonV2PathParams {
  workspace_id: string;
  agent_id: string;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다 (v2 workspace-scoped)
 */
export async function restartAIAgentDaemonV2(config: RiidoClientConfig, params: RestartAIAgentDaemonV2PathParams, body: ControlDeviceDaemonRequest, options: RiidoRequestOptions = {}): Promise<DeviceDaemonCommandResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/agents/${params.agent_id}/daemon/restart`;
  return riidoRequest<DeviceDaemonCommandResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface RestartAIAgentDaemonV2MutationVariables {
  params: RestartAIAgentDaemonV2PathParams;
  body: ControlDeviceDaemonRequest;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function restartAIAgentDaemonV2MutationKey(): readonly unknown[] {
  return ["restartAIAgentDaemonV2"] as const;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function restartAIAgentDaemonV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDaemonV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: restartAIAgentDaemonV2MutationKey(),
    mutationFn: (variables: RestartAIAgentDaemonV2MutationVariables) => restartAIAgentDaemonV2(config, variables.params, variables.body, {}),
  };
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface StartAIAgentDaemonV2PathParams {
  workspace_id: string;
  agent_id: string;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다 (v2 workspace-scoped)
 */
export async function startAIAgentDaemonV2(config: RiidoClientConfig, params: StartAIAgentDaemonV2PathParams, body: ControlDeviceDaemonRequest, options: RiidoRequestOptions = {}): Promise<DeviceDaemonCommandResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/agents/${params.agent_id}/daemon/start`;
  return riidoRequest<DeviceDaemonCommandResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface StartAIAgentDaemonV2MutationVariables {
  params: StartAIAgentDaemonV2PathParams;
  body: ControlDeviceDaemonRequest;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function startAIAgentDaemonV2MutationKey(): readonly unknown[] {
  return ["startAIAgentDaemonV2"] as const;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function startAIAgentDaemonV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDaemonV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: startAIAgentDaemonV2MutationKey(),
    mutationFn: (variables: StartAIAgentDaemonV2MutationVariables) => startAIAgentDaemonV2(config, variables.params, variables.body, {}),
  };
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface StopAIAgentDaemonV2PathParams {
  workspace_id: string;
  agent_id: string;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다 (v2 workspace-scoped)
 */
export async function stopAIAgentDaemonV2(config: RiidoClientConfig, params: StopAIAgentDaemonV2PathParams, body: ControlDeviceDaemonRequest, options: RiidoRequestOptions = {}): Promise<DeviceDaemonCommandResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/agents/${params.agent_id}/daemon/stop`;
  return riidoRequest<DeviceDaemonCommandResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface StopAIAgentDaemonV2MutationVariables {
  params: StopAIAgentDaemonV2PathParams;
  body: ControlDeviceDaemonRequest;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function stopAIAgentDaemonV2MutationKey(): readonly unknown[] {
  return ["stopAIAgentDaemonV2"] as const;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function stopAIAgentDaemonV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDaemonV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: stopAIAgentDaemonV2MutationKey(),
    mutationFn: (variables: StopAIAgentDaemonV2MutationVariables) => stopAIAgentDaemonV2(config, variables.params, variables.body, {}),
  };
}

/**
 * client control 활성화 전에 agent 수정 가능 여부를 조회합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface GetAIAgentEditabilityV2PathParams {
  workspace_id: string;
  agent_id: string;
}

/**
 * client control 활성화 전에 agent 수정 가능 여부를 조회합니다 (v2 workspace-scoped)
 */
export async function getAIAgentEditabilityV2(config: RiidoClientConfig, params: GetAIAgentEditabilityV2PathParams, options: RiidoRequestOptions = {}): Promise<AgentEditabilityResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/agents/${params.agent_id}/editability`;
  return riidoRequest<AgentEditabilityResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * client control 활성화 전에 agent 수정 가능 여부를 조회합니다 (v2 workspace-scoped)
 * cache tag: `v2.aiAgent.agents.editability`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function getAIAgentEditabilityV2QueryKeyRoot(): readonly unknown[] {
  return ["v2.aiAgent.agents.editability"] as const;
}

/**
 * client control 활성화 전에 agent 수정 가능 여부를 조회합니다 (v2 workspace-scoped)
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function getAIAgentEditabilityV2QueryKey(params: GetAIAgentEditabilityV2PathParams): readonly unknown[] {
  return [...getAIAgentEditabilityV2QueryKeyRoot(), params] as const;
}

/**
 * client control 활성화 전에 agent 수정 가능 여부를 조회합니다 (v2 workspace-scoped)
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function getAIAgentEditabilityV2QueryOptions(config: RiidoClientConfig, params: GetAIAgentEditabilityV2PathParams, options: RiidoQueryOptions<AgentEditabilityResponse> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: getAIAgentEditabilityV2QueryKey(params),
    queryFn: () => getAIAgentEditabilityV2(config, params, { signal }),
  };
}

/**
 * web 또는 desktop webview client의 AI Agent 설정/온보딩 초기 데이터와 v2.aiAgent.bootstrap.agents[] agent list를 조회합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface GetAIAgentClientBootstrapV2PathParams {
  workspace_id: string;
}

/**
 * web 또는 desktop webview client의 AI Agent 설정/온보딩 초기 데이터와 v2.aiAgent.bootstrap.agents[] agent list를 조회합니다 (v2 workspace-scoped)
 */
export async function getAIAgentClientBootstrapV2(config: RiidoClientConfig, params: GetAIAgentClientBootstrapV2PathParams, options: RiidoRequestOptions = {}): Promise<ClientBootstrapResponseV2> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/bootstrap`;
  return riidoRequest<ClientBootstrapResponseV2>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * web 또는 desktop webview client의 AI Agent 설정/온보딩 초기 데이터와 v2.aiAgent.bootstrap.agents[] agent list를 조회합니다 (v2 workspace-scoped)
 * cache tag: `v2.aiAgent.bootstrap`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function getAIAgentClientBootstrapV2QueryKeyRoot(): readonly unknown[] {
  return ["v2.aiAgent.bootstrap"] as const;
}

/**
 * web 또는 desktop webview client의 AI Agent 설정/온보딩 초기 데이터와 v2.aiAgent.bootstrap.agents[] agent list를 조회합니다 (v2 workspace-scoped)
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function getAIAgentClientBootstrapV2QueryKey(params: GetAIAgentClientBootstrapV2PathParams): readonly unknown[] {
  return [...getAIAgentClientBootstrapV2QueryKeyRoot(), params] as const;
}

/**
 * web 또는 desktop webview client의 AI Agent 설정/온보딩 초기 데이터와 v2.aiAgent.bootstrap.agents[] agent list를 조회합니다 (v2 workspace-scoped)
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function getAIAgentClientBootstrapV2QueryOptions(config: RiidoClientConfig, params: GetAIAgentClientBootstrapV2PathParams, options: RiidoQueryOptions<ClientBootstrapResponseV2> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: getAIAgentClientBootstrapV2QueryKey(params),
    queryFn: () => getAIAgentClientBootstrapV2(config, params, { signal }),
  };
}

/**
 * 권한이 확인된 principal의 device runtime 상태를 조회합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface ListAIAgentDeviceRuntimesV2PathParams {
  workspace_id: string;
}

/**
 * 권한이 확인된 principal의 device runtime 상태를 조회합니다 (v2 workspace-scoped)
 */
export async function listAIAgentDeviceRuntimesV2(config: RiidoClientConfig, params: ListAIAgentDeviceRuntimesV2PathParams, options: RiidoRequestOptions = {}): Promise<DeviceRuntimeListResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/devices`;
  return riidoRequest<DeviceRuntimeListResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * 권한이 확인된 principal의 device runtime 상태를 조회합니다 (v2 workspace-scoped)
 * cache tag: `v2.aiAgent.devices.runtimes`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function listAIAgentDeviceRuntimesV2QueryKeyRoot(): readonly unknown[] {
  return ["v2.aiAgent.devices.runtimes"] as const;
}

/**
 * 권한이 확인된 principal의 device runtime 상태를 조회합니다 (v2 workspace-scoped)
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function listAIAgentDeviceRuntimesV2QueryKey(params: ListAIAgentDeviceRuntimesV2PathParams): readonly unknown[] {
  return [...listAIAgentDeviceRuntimesV2QueryKeyRoot(), params] as const;
}

/**
 * 권한이 확인된 principal의 device runtime 상태를 조회합니다 (v2 workspace-scoped)
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function listAIAgentDeviceRuntimesV2QueryOptions(config: RiidoClientConfig, params: ListAIAgentDeviceRuntimesV2PathParams, options: RiidoQueryOptions<DeviceRuntimeListResponse> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: listAIAgentDeviceRuntimesV2QueryKey(params),
    queryFn: () => listAIAgentDeviceRuntimesV2(config, params, { signal }),
  };
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 상세를 조회합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface GetAIAgentDeviceDaemonV2PathParams {
  workspace_id: string;
  device_id: string;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 상세를 조회합니다 (v2 workspace-scoped)
 */
export async function getAIAgentDeviceDaemonV2(config: RiidoClientConfig, params: GetAIAgentDeviceDaemonV2PathParams, options: RiidoRequestOptions = {}): Promise<DeviceDaemonDetailResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/devices/${params.device_id}/daemon`;
  return riidoRequest<DeviceDaemonDetailResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 상세를 조회합니다 (v2 workspace-scoped)
 * cache tag: `v2.aiAgent.devices.daemon`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function getAIAgentDeviceDaemonV2QueryKeyRoot(): readonly unknown[] {
  return ["v2.aiAgent.devices.daemon"] as const;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 상세를 조회합니다 (v2 workspace-scoped)
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function getAIAgentDeviceDaemonV2QueryKey(params: GetAIAgentDeviceDaemonV2PathParams): readonly unknown[] {
  return [...getAIAgentDeviceDaemonV2QueryKeyRoot(), params] as const;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 상세를 조회합니다 (v2 workspace-scoped)
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function getAIAgentDeviceDaemonV2QueryOptions(config: RiidoClientConfig, params: GetAIAgentDeviceDaemonV2PathParams, options: RiidoQueryOptions<DeviceDaemonDetailResponse> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: getAIAgentDeviceDaemonV2QueryKey(params),
    queryFn: () => getAIAgentDeviceDaemonV2(config, params, { signal }),
  };
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 재시작을 요청합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface RestartAIAgentDeviceDaemonV2PathParams {
  workspace_id: string;
  device_id: string;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 재시작을 요청합니다 (v2 workspace-scoped)
 */
export async function restartAIAgentDeviceDaemonV2(config: RiidoClientConfig, params: RestartAIAgentDeviceDaemonV2PathParams, body: ControlDeviceDaemonRequest, options: RiidoRequestOptions = {}): Promise<DeviceDaemonCommandResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/devices/${params.device_id}/daemon/restart`;
  return riidoRequest<DeviceDaemonCommandResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 재시작을 요청합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface RestartAIAgentDeviceDaemonV2MutationVariables {
  params: RestartAIAgentDeviceDaemonV2PathParams;
  body: ControlDeviceDaemonRequest;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 재시작을 요청합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function restartAIAgentDeviceDaemonV2MutationKey(): readonly unknown[] {
  return ["restartAIAgentDeviceDaemonV2"] as const;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 재시작을 요청합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function restartAIAgentDeviceDaemonV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDeviceDaemonV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: restartAIAgentDeviceDaemonV2MutationKey(),
    mutationFn: (variables: RestartAIAgentDeviceDaemonV2MutationVariables) => restartAIAgentDeviceDaemonV2(config, variables.params, variables.body, {}),
  };
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 시작을 요청합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface StartAIAgentDeviceDaemonV2PathParams {
  workspace_id: string;
  device_id: string;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 시작을 요청합니다 (v2 workspace-scoped)
 */
export async function startAIAgentDeviceDaemonV2(config: RiidoClientConfig, params: StartAIAgentDeviceDaemonV2PathParams, body: ControlDeviceDaemonRequest, options: RiidoRequestOptions = {}): Promise<DeviceDaemonCommandResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/devices/${params.device_id}/daemon/start`;
  return riidoRequest<DeviceDaemonCommandResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 시작을 요청합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface StartAIAgentDeviceDaemonV2MutationVariables {
  params: StartAIAgentDeviceDaemonV2PathParams;
  body: ControlDeviceDaemonRequest;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 시작을 요청합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function startAIAgentDeviceDaemonV2MutationKey(): readonly unknown[] {
  return ["startAIAgentDeviceDaemonV2"] as const;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 시작을 요청합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function startAIAgentDeviceDaemonV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDeviceDaemonV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: startAIAgentDeviceDaemonV2MutationKey(),
    mutationFn: (variables: StartAIAgentDeviceDaemonV2MutationVariables) => startAIAgentDeviceDaemonV2(config, variables.params, variables.body, {}),
  };
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 중지를 요청합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface StopAIAgentDeviceDaemonV2PathParams {
  workspace_id: string;
  device_id: string;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 중지를 요청합니다 (v2 workspace-scoped)
 */
export async function stopAIAgentDeviceDaemonV2(config: RiidoClientConfig, params: StopAIAgentDeviceDaemonV2PathParams, body: ControlDeviceDaemonRequest, options: RiidoRequestOptions = {}): Promise<DeviceDaemonCommandResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/devices/${params.device_id}/daemon/stop`;
  return riidoRequest<DeviceDaemonCommandResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 중지를 요청합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface StopAIAgentDeviceDaemonV2MutationVariables {
  params: StopAIAgentDeviceDaemonV2PathParams;
  body: ControlDeviceDaemonRequest;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 중지를 요청합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function stopAIAgentDeviceDaemonV2MutationKey(): readonly unknown[] {
  return ["stopAIAgentDeviceDaemonV2"] as const;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 중지를 요청합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function stopAIAgentDeviceDaemonV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDeviceDaemonV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: stopAIAgentDeviceDaemonV2MutationKey(),
    mutationFn: (variables: StopAIAgentDeviceDaemonV2MutationVariables) => stopAIAgentDeviceDaemonV2(config, variables.params, variables.body, {}),
  };
}

/**
 * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface StreamAIAgentClientEventsV2PathParams {
  workspace_id: string;
}

/**
 * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다 (v2 workspace-scoped)
 */
export async function streamAIAgentClientEventsV2(config: RiidoClientConfig, params: StreamAIAgentClientEventsV2PathParams, options: RiidoRequestOptions = {}): Promise<Response> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/events`;
  return riidoRawRequest(config, path, { method: 'GET', signal: options.signal });
}

/**
 * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다 (v2 workspace-scoped)
 * cache tag: `v2.aiAgent.events.stream`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function streamAIAgentClientEventsV2QueryKeyRoot(): readonly unknown[] {
  return ["v2.aiAgent.events.stream"] as const;
}

/**
 * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다 (v2 workspace-scoped)
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function streamAIAgentClientEventsV2QueryKey(params: StreamAIAgentClientEventsV2PathParams): readonly unknown[] {
  return [...streamAIAgentClientEventsV2QueryKeyRoot(), params] as const;
}

/**
 * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다 (v2 workspace-scoped)
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function streamAIAgentClientEventsV2QueryOptions(config: RiidoClientConfig, params: StreamAIAgentClientEventsV2PathParams, options: RiidoQueryOptions<Response> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: streamAIAgentClientEventsV2QueryKey(params),
    queryFn: () => streamAIAgentClientEventsV2(config, params, { signal }),
  };
}

/**
 * AI Agent 온보딩에서 사용할 서버 제공 fixture 목록을 조회합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface ListAIAgentOnboardingFixturesV2PathParams {
  workspace_id: string;
}

/**
 * AI Agent 온보딩에서 사용할 서버 제공 fixture 목록을 조회합니다 (v2 workspace-scoped)
 */
export async function listAIAgentOnboardingFixturesV2(config: RiidoClientConfig, params: ListAIAgentOnboardingFixturesV2PathParams, options: RiidoRequestOptions = {}): Promise<AgentOnboardingFixtureListResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/onboarding/fixtures`;
  return riidoRequest<AgentOnboardingFixtureListResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * AI Agent 온보딩에서 사용할 서버 제공 fixture 목록을 조회합니다 (v2 workspace-scoped)
 * cache tag: `v2.aiAgent.onboarding.fixtures`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function listAIAgentOnboardingFixturesV2QueryKeyRoot(): readonly unknown[] {
  return ["v2.aiAgent.onboarding.fixtures"] as const;
}

/**
 * AI Agent 온보딩에서 사용할 서버 제공 fixture 목록을 조회합니다 (v2 workspace-scoped)
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function listAIAgentOnboardingFixturesV2QueryKey(params: ListAIAgentOnboardingFixturesV2PathParams): readonly unknown[] {
  return [...listAIAgentOnboardingFixturesV2QueryKeyRoot(), params] as const;
}

/**
 * AI Agent 온보딩에서 사용할 서버 제공 fixture 목록을 조회합니다 (v2 workspace-scoped)
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function listAIAgentOnboardingFixturesV2QueryOptions(config: RiidoClientConfig, params: ListAIAgentOnboardingFixturesV2PathParams, options: RiidoQueryOptions<AgentOnboardingFixtureListResponse> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: listAIAgentOnboardingFixturesV2QueryKey(params),
    queryFn: () => listAIAgentOnboardingFixturesV2(config, params, { signal }),
  };
}

/**
 * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface CreateAIAgentFromOnboardingFixtureV2PathParams {
  workspace_id: string;
  fixture_id: string;
}

/**
 * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다 (v2 workspace-scoped)
 */
export async function createAIAgentFromOnboardingFixtureV2(config: RiidoClientConfig, params: CreateAIAgentFromOnboardingFixtureV2PathParams, body: CreateAgentConfigurationRequest, options: RiidoRequestOptions = {}): Promise<AgentClientRecordResponseV2> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/onboarding/fixtures/${params.fixture_id}/agents`;
  return riidoRequest<AgentClientRecordResponseV2>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface CreateAIAgentFromOnboardingFixtureV2MutationVariables {
  params: CreateAIAgentFromOnboardingFixtureV2PathParams;
  body: CreateAgentConfigurationRequest;
}

/**
 * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function createAIAgentFromOnboardingFixtureV2MutationKey(): readonly unknown[] {
  return ["createAIAgentFromOnboardingFixtureV2"] as const;
}

/**
 * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function createAIAgentFromOnboardingFixtureV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AgentClientRecordResponseV2, CreateAIAgentFromOnboardingFixtureV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: createAIAgentFromOnboardingFixtureV2MutationKey(),
    mutationFn: (variables: CreateAIAgentFromOnboardingFixtureV2MutationVariables) => createAIAgentFromOnboardingFixtureV2(config, variables.params, variables.body, {}),
  };
}

/**
 * workspace task/card 목록에서 현재 배정된 AI Agent profile 표시값을 component_id key 해시맵으로 조회합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface ListWorkspaceAssignedAgentProfilesV2PathParams {
  workspace_id: string;
}

/**
 * workspace task/card 목록에서 현재 배정된 AI Agent profile 표시값을 component_id key 해시맵으로 조회합니다 (v2 workspace-scoped)
 */
export async function listWorkspaceAssignedAgentProfilesV2(config: RiidoClientConfig, params: ListWorkspaceAssignedAgentProfilesV2PathParams, options: RiidoRequestOptions = {}): Promise<AssignedAgentProfileMapResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/tasks/assigned-agent-profiles`;
  return riidoRequest<AssignedAgentProfileMapResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * workspace task/card 목록에서 현재 배정된 AI Agent profile 표시값을 component_id key 해시맵으로 조회합니다 (v2 workspace-scoped)
 * cache tag: `v2.aiAgent.tasks.assignedAgentProfiles`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function listWorkspaceAssignedAgentProfilesV2QueryKeyRoot(): readonly unknown[] {
  return ["v2.aiAgent.tasks.assignedAgentProfiles"] as const;
}

/**
 * workspace task/card 목록에서 현재 배정된 AI Agent profile 표시값을 component_id key 해시맵으로 조회합니다 (v2 workspace-scoped)
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function listWorkspaceAssignedAgentProfilesV2QueryKey(params: ListWorkspaceAssignedAgentProfilesV2PathParams): readonly unknown[] {
  return [...listWorkspaceAssignedAgentProfilesV2QueryKeyRoot(), params] as const;
}

/**
 * workspace task/card 목록에서 현재 배정된 AI Agent profile 표시값을 component_id key 해시맵으로 조회합니다 (v2 workspace-scoped)
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function listWorkspaceAssignedAgentProfilesV2QueryOptions(config: RiidoClientConfig, params: ListWorkspaceAssignedAgentProfilesV2PathParams, options: RiidoQueryOptions<AssignedAgentProfileMapResponse> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: listWorkspaceAssignedAgentProfilesV2QueryKey(params),
    queryFn: () => listWorkspaceAssignedAgentProfilesV2(config, params, { signal }),
  };
}

/**
 * task에 AI agent를 additive 방식으로 배정합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface CreateAIAgentTaskAgentAssignmentV2PathParams {
  workspace_id: string;
  task_id: string;
}

/**
 * task에 AI agent를 additive 방식으로 배정합니다 (v2 workspace-scoped)
 */
export async function createAIAgentTaskAgentAssignmentV2(config: RiidoClientConfig, params: CreateAIAgentTaskAgentAssignmentV2PathParams, body: AssignAIAgentTaskRequest, options: RiidoRequestOptions = {}): Promise<AIAgentTaskActionResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/tasks/${params.task_id}/agent-assignments`;
  return riidoRequest<AIAgentTaskActionResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * task에 AI agent를 additive 방식으로 배정합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface CreateAIAgentTaskAgentAssignmentV2MutationVariables {
  params: CreateAIAgentTaskAgentAssignmentV2PathParams;
  body: AssignAIAgentTaskRequest;
}

/**
 * task에 AI agent를 additive 방식으로 배정합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function createAIAgentTaskAgentAssignmentV2MutationKey(): readonly unknown[] {
  return ["createAIAgentTaskAgentAssignmentV2"] as const;
}

/**
 * task에 AI agent를 additive 방식으로 배정합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function createAIAgentTaskAgentAssignmentV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AIAgentTaskActionResponse, CreateAIAgentTaskAgentAssignmentV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: createAIAgentTaskAgentAssignmentV2MutationKey(),
    mutationFn: (variables: CreateAIAgentTaskAgentAssignmentV2MutationVariables) => createAIAgentTaskAgentAssignmentV2(config, variables.params, variables.body, {}),
  };
}

/**
 * task의 특정 AI agent assignment를 제거하고 해당 agent 작업만 중단합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface DeleteAIAgentTaskAgentAssignmentV2PathParams {
  workspace_id: string;
  task_id: string;
  agent_id: string;
}

/**
 * task의 특정 AI agent assignment를 제거하고 해당 agent 작업만 중단합니다 (v2 workspace-scoped)
 */
export async function deleteAIAgentTaskAgentAssignmentV2(config: RiidoClientConfig, params: DeleteAIAgentTaskAgentAssignmentV2PathParams, options: RiidoRequestOptions = {}): Promise<AIAgentTaskActionResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/tasks/${params.task_id}/agent-assignments/${params.agent_id}`;
  return riidoRequest<AIAgentTaskActionResponse>(config, path, { method: 'DELETE', signal: options.signal });
}

/**
 * task의 특정 AI agent assignment를 제거하고 해당 agent 작업만 중단합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface DeleteAIAgentTaskAgentAssignmentV2MutationVariables {
  params: DeleteAIAgentTaskAgentAssignmentV2PathParams;
}

/**
 * task의 특정 AI agent assignment를 제거하고 해당 agent 작업만 중단합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function deleteAIAgentTaskAgentAssignmentV2MutationKey(): readonly unknown[] {
  return ["deleteAIAgentTaskAgentAssignmentV2"] as const;
}

/**
 * task의 특정 AI agent assignment를 제거하고 해당 agent 작업만 중단합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function deleteAIAgentTaskAgentAssignmentV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AIAgentTaskActionResponse, DeleteAIAgentTaskAgentAssignmentV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: deleteAIAgentTaskAgentAssignmentV2MutationKey(),
    mutationFn: (variables: DeleteAIAgentTaskAgentAssignmentV2MutationVariables) => deleteAIAgentTaskAgentAssignmentV2(config, variables.params, {}),
  };
}

/**
 * task의 특정 AI agent 작업을 agent_id 기준으로 중단합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface StopAIAgentTaskAgentAssignmentV2PathParams {
  workspace_id: string;
  task_id: string;
  agent_id: string;
}

/**
 * task의 특정 AI agent 작업을 agent_id 기준으로 중단합니다 (v2 workspace-scoped)
 */
export async function stopAIAgentTaskAgentAssignmentV2(config: RiidoClientConfig, params: StopAIAgentTaskAgentAssignmentV2PathParams, options: RiidoRequestOptions = {}): Promise<AIAgentTaskActionResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/tasks/${params.task_id}/agent-assignments/${params.agent_id}/stop`;
  return riidoRequest<AIAgentTaskActionResponse>(config, path, { method: 'POST', signal: options.signal });
}

/**
 * task의 특정 AI agent 작업을 agent_id 기준으로 중단합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface StopAIAgentTaskAgentAssignmentV2MutationVariables {
  params: StopAIAgentTaskAgentAssignmentV2PathParams;
}

/**
 * task의 특정 AI agent 작업을 agent_id 기준으로 중단합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function stopAIAgentTaskAgentAssignmentV2MutationKey(): readonly unknown[] {
  return ["stopAIAgentTaskAgentAssignmentV2"] as const;
}

/**
 * task의 특정 AI agent 작업을 agent_id 기준으로 중단합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function stopAIAgentTaskAgentAssignmentV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AIAgentTaskActionResponse, StopAIAgentTaskAgentAssignmentV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: stopAIAgentTaskAgentAssignmentV2MutationKey(),
    mutationFn: (variables: StopAIAgentTaskAgentAssignmentV2MutationVariables) => stopAIAgentTaskAgentAssignmentV2(config, variables.params, {}),
  };
}

/**
 * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface ListAIAgentTaskAssignableAgentsV2PathParams {
  workspace_id: string;
  task_id: string;
}

/**
 * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다 (v2 workspace-scoped)
 */
export async function listAIAgentTaskAssignableAgentsV2(config: RiidoClientConfig, params: ListAIAgentTaskAssignableAgentsV2PathParams, options: RiidoRequestOptions = {}): Promise<AgentClientListResponseV2> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/tasks/${params.task_id}/assignable-agents`;
  return riidoRequest<AgentClientListResponseV2>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다 (v2 workspace-scoped)
 * cache tag: `v2.aiAgent.tasks.assignableAgents`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function listAIAgentTaskAssignableAgentsV2QueryKeyRoot(): readonly unknown[] {
  return ["v2.aiAgent.tasks.assignableAgents"] as const;
}

/**
 * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다 (v2 workspace-scoped)
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function listAIAgentTaskAssignableAgentsV2QueryKey(params: ListAIAgentTaskAssignableAgentsV2PathParams): readonly unknown[] {
  return [...listAIAgentTaskAssignableAgentsV2QueryKeyRoot(), params] as const;
}

/**
 * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다 (v2 workspace-scoped)
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function listAIAgentTaskAssignableAgentsV2QueryOptions(config: RiidoClientConfig, params: ListAIAgentTaskAssignableAgentsV2PathParams, options: RiidoQueryOptions<AgentClientListResponseV2> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: listAIAgentTaskAssignableAgentsV2QueryKey(params),
    queryFn: () => listAIAgentTaskAssignableAgentsV2(config, params, { signal }),
  };
}

/**
 * task participant에서 AI agent를 제거하고 작업을 중단합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface UnassignAIAgentTaskV2PathParams {
  workspace_id: string;
  task_id: string;
}

/**
 * task participant에서 AI agent를 제거하고 작업을 중단합니다 (v2 workspace-scoped)
 */
export async function unassignAIAgentTaskV2(config: RiidoClientConfig, params: UnassignAIAgentTaskV2PathParams, body: UnassignAIAgentTaskRequest, options: RiidoRequestOptions = {}): Promise<AIAgentTaskActionResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/tasks/${params.task_id}/assignment`;
  return riidoRequest<AIAgentTaskActionResponse>(config, path, { method: 'DELETE', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * task participant에서 AI agent를 제거하고 작업을 중단합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface UnassignAIAgentTaskV2MutationVariables {
  params: UnassignAIAgentTaskV2PathParams;
  body: UnassignAIAgentTaskRequest;
}

/**
 * task participant에서 AI agent를 제거하고 작업을 중단합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function unassignAIAgentTaskV2MutationKey(): readonly unknown[] {
  return ["unassignAIAgentTaskV2"] as const;
}

/**
 * task participant에서 AI agent를 제거하고 작업을 중단합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function unassignAIAgentTaskV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AIAgentTaskActionResponse, UnassignAIAgentTaskV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: unassignAIAgentTaskV2MutationKey(),
    mutationFn: (variables: UnassignAIAgentTaskV2MutationVariables) => unassignAIAgentTaskV2(config, variables.params, variables.body, {}),
  };
}

/**
 * task participant dropdown에서 AI agent를 배정합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface AssignAIAgentTaskV2PathParams {
  workspace_id: string;
  task_id: string;
}

/**
 * task participant dropdown에서 AI agent를 배정합니다 (v2 workspace-scoped)
 */
export async function assignAIAgentTaskV2(config: RiidoClientConfig, params: AssignAIAgentTaskV2PathParams, body: AssignAIAgentTaskRequest, options: RiidoRequestOptions = {}): Promise<AIAgentTaskActionResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/tasks/${params.task_id}/assignment`;
  return riidoRequest<AIAgentTaskActionResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * task participant dropdown에서 AI agent를 배정합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface AssignAIAgentTaskV2MutationVariables {
  params: AssignAIAgentTaskV2PathParams;
  body: AssignAIAgentTaskRequest;
}

/**
 * task participant dropdown에서 AI agent를 배정합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function assignAIAgentTaskV2MutationKey(): readonly unknown[] {
  return ["assignAIAgentTaskV2"] as const;
}

/**
 * task participant dropdown에서 AI agent를 배정합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function assignAIAgentTaskV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AIAgentTaskActionResponse, AssignAIAgentTaskV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: assignAIAgentTaskV2MutationKey(),
    mutationFn: (variables: AssignAIAgentTaskV2MutationVariables) => assignAIAgentTaskV2(config, variables.params, variables.body, {}),
  };
}

/**
 * 호환 task comment route로 AI agent에게 메시지를 전달합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface SubmitAIAgentTaskCommentV2PathParams {
  workspace_id: string;
  task_id: string;
}

/**
 * 호환 task comment route로 AI agent에게 메시지를 전달합니다 (v2 workspace-scoped)
 */
export async function submitAIAgentTaskCommentV2(config: RiidoClientConfig, params: SubmitAIAgentTaskCommentV2PathParams, body: SubmitAIAgentTaskCommentRequest, options: RiidoRequestOptions = {}): Promise<AIAgentTaskActionResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/tasks/${params.task_id}/comments`;
  return riidoRequest<AIAgentTaskActionResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * 호환 task comment route로 AI agent에게 메시지를 전달합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface SubmitAIAgentTaskCommentV2MutationVariables {
  params: SubmitAIAgentTaskCommentV2PathParams;
  body: SubmitAIAgentTaskCommentRequest;
}

/**
 * 호환 task comment route로 AI agent에게 메시지를 전달합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function submitAIAgentTaskCommentV2MutationKey(): readonly unknown[] {
  return ["submitAIAgentTaskCommentV2"] as const;
}

/**
 * 호환 task comment route로 AI agent에게 메시지를 전달합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function submitAIAgentTaskCommentV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AIAgentTaskActionResponse, SubmitAIAgentTaskCommentV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: submitAIAgentTaskCommentV2MutationKey(),
    mutationFn: (variables: SubmitAIAgentTaskCommentV2MutationVariables) => submitAIAgentTaskCommentV2(config, variables.params, variables.body, {}),
  };
}

/**
 * task thread의 stop action으로 AI agent 작업을 중단합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface StopAIAgentTaskV2PathParams {
  workspace_id: string;
  task_id: string;
}

/**
 * task thread의 stop action으로 AI agent 작업을 중단합니다 (v2 workspace-scoped)
 */
export async function stopAIAgentTaskV2(config: RiidoClientConfig, params: StopAIAgentTaskV2PathParams, body: StopAIAgentTaskRequest, options: RiidoRequestOptions = {}): Promise<AIAgentTaskActionResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/tasks/${params.task_id}/stop`;
  return riidoRequest<AIAgentTaskActionResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * task thread의 stop action으로 AI agent 작업을 중단합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface StopAIAgentTaskV2MutationVariables {
  params: StopAIAgentTaskV2PathParams;
  body: StopAIAgentTaskRequest;
}

/**
 * task thread의 stop action으로 AI agent 작업을 중단합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function stopAIAgentTaskV2MutationKey(): readonly unknown[] {
  return ["stopAIAgentTaskV2"] as const;
}

/**
 * task thread의 stop action으로 AI agent 작업을 중단합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function stopAIAgentTaskV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AIAgentTaskActionResponse, StopAIAgentTaskV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: stopAIAgentTaskV2MutationKey(),
    mutationFn: (variables: StopAIAgentTaskV2MutationVariables) => stopAIAgentTaskV2(config, variables.params, variables.body, {}),
  };
}

/**
 * task의 여러 active AI Agent thread를 하나의 SSE stream으로 구독하기 위한 filter handoff를 조회합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface GetAIAgentTaskThreadStreamSubscriptionV2PathParams {
  workspace_id: string;
  task_id: string;
}

/**
 * task의 여러 active AI Agent thread를 하나의 SSE stream으로 구독하기 위한 filter handoff를 조회합니다 (v2 workspace-scoped)
 */
export async function getAIAgentTaskThreadStreamSubscriptionV2(config: RiidoClientConfig, params: GetAIAgentTaskThreadStreamSubscriptionV2PathParams, options: RiidoRequestOptions = {}): Promise<AIAgentTaskThreadStreamSubscriptionResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/tasks/${params.task_id}/thread-stream-subscription`;
  return riidoRequest<AIAgentTaskThreadStreamSubscriptionResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * task의 여러 active AI Agent thread를 하나의 SSE stream으로 구독하기 위한 filter handoff를 조회합니다 (v2 workspace-scoped)
 * cache tag: `v2.aiAgent.tasks.threadStreamSubscription`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function getAIAgentTaskThreadStreamSubscriptionV2QueryKeyRoot(): readonly unknown[] {
  return ["v2.aiAgent.tasks.threadStreamSubscription"] as const;
}

/**
 * task의 여러 active AI Agent thread를 하나의 SSE stream으로 구독하기 위한 filter handoff를 조회합니다 (v2 workspace-scoped)
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function getAIAgentTaskThreadStreamSubscriptionV2QueryKey(params: GetAIAgentTaskThreadStreamSubscriptionV2PathParams): readonly unknown[] {
  return [...getAIAgentTaskThreadStreamSubscriptionV2QueryKeyRoot(), params] as const;
}

/**
 * task의 여러 active AI Agent thread를 하나의 SSE stream으로 구독하기 위한 filter handoff를 조회합니다 (v2 workspace-scoped)
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function getAIAgentTaskThreadStreamSubscriptionV2QueryOptions(config: RiidoClientConfig, params: GetAIAgentTaskThreadStreamSubscriptionV2PathParams, options: RiidoQueryOptions<AIAgentTaskThreadStreamSubscriptionResponse> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: getAIAgentTaskThreadStreamSubscriptionV2QueryKey(params),
    queryFn: () => getAIAgentTaskThreadStreamSubscriptionV2(config, params, { signal }),
  };
}

/**
 * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface ListAIAgentTaskThreadsV2PathParams {
  workspace_id: string;
  task_id: string;
}

/**
 * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다 (v2 workspace-scoped)
 */
export async function listAIAgentTaskThreadsV2(config: RiidoClientConfig, params: ListAIAgentTaskThreadsV2PathParams, options: RiidoRequestOptions = {}): Promise<AIAgentTaskThreadCollectionResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/tasks/${params.task_id}/threads`;
  return riidoRequest<AIAgentTaskThreadCollectionResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다 (v2 workspace-scoped)
 * cache tag: `v2.aiAgent.tasks.threads`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function listAIAgentTaskThreadsV2QueryKeyRoot(): readonly unknown[] {
  return ["v2.aiAgent.tasks.threads"] as const;
}

/**
 * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다 (v2 workspace-scoped)
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function listAIAgentTaskThreadsV2QueryKey(params: ListAIAgentTaskThreadsV2PathParams): readonly unknown[] {
  return [...listAIAgentTaskThreadsV2QueryKeyRoot(), params] as const;
}

/**
 * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다 (v2 workspace-scoped)
 * useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.
 */
export function listAIAgentTaskThreadsV2QueryOptions(config: RiidoClientConfig, params: ListAIAgentTaskThreadsV2PathParams, options: RiidoQueryOptions<AIAgentTaskThreadCollectionResponse> = {}) {
  const { signal, ...queryOptions } = options;
  return {
    ...queryOptions,
    queryKey: listAIAgentTaskThreadsV2QueryKey(params),
    queryFn: () => listAIAgentTaskThreadsV2(config, params, { signal }),
  };
}

/**
 * task thread message로 AI agent에게 다음 작업 지시를 전달합니다 (v2 workspace-scoped)
 * 경로 파라미터입니다.
 */
export interface CreateAIAgentTaskThreadMessageV2PathParams {
  workspace_id: string;
  task_id: string;
  thread_id: string;
}

/**
 * task thread message로 AI agent에게 다음 작업 지시를 전달합니다 (v2 workspace-scoped)
 */
export async function createAIAgentTaskThreadMessageV2(config: RiidoClientConfig, params: CreateAIAgentTaskThreadMessageV2PathParams, body: CreateAIAgentTaskThreadMessageRequest, options: RiidoRequestOptions = {}): Promise<AIAgentTaskActionResponse> {
  const path = `/v2/client/workspaces/${params.workspace_id}/ai-agent/tasks/${params.task_id}/threads/${params.thread_id}/messages`;
  return riidoRequest<AIAgentTaskActionResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * task thread message로 AI agent에게 다음 작업 지시를 전달합니다 (v2 workspace-scoped)
 * mutation 함수에 전달하는 변수입니다.
 */
export interface CreateAIAgentTaskThreadMessageV2MutationVariables {
  params: CreateAIAgentTaskThreadMessageV2PathParams;
  body: CreateAIAgentTaskThreadMessageRequest;
}

/**
 * task thread message로 AI agent에게 다음 작업 지시를 전달합니다 (v2 workspace-scoped)
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function createAIAgentTaskThreadMessageV2MutationKey(): readonly unknown[] {
  return ["createAIAgentTaskThreadMessageV2"] as const;
}

/**
 * task thread message로 AI agent에게 다음 작업 지시를 전달합니다 (v2 workspace-scoped)
 * useMutation에 전달할 수 있는 옵션입니다.
 */
export function createAIAgentTaskThreadMessageV2MutationOptions(config: RiidoClientConfig, options: RiidoMutationOptions<AIAgentTaskActionResponse, CreateAIAgentTaskThreadMessageV2MutationVariables> = {}) {
  return {
    ...options,
    mutationKey: createAIAgentTaskThreadMessageV2MutationKey(),
    mutationFn: (variables: CreateAIAgentTaskThreadMessageV2MutationVariables) => createAIAgentTaskThreadMessageV2(config, variables.params, variables.body, {}),
  };
}

/**
 * Figma agent 추가 화면에서 AI agent 설정을 생성합니다
 * 계약 generated path: `aiAgent.agents.create`
 * 검색용 generated 경로: `agents.create`
 * 접근 예시: `riido.aiAgent.agents.create`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface CreateAIAgentEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (body: CreateAgentConfigurationRequest, options?: RiidoRequestOptions) => Promise<AgentClientRecordResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AgentClientRecordResponse, CreateAIAgentMutationVariables>) => ReturnType<typeof createAIAgentMutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AgentClientRecordResponse, CreateAIAgentMutationVariables>) => ReturnType<typeof createAIAgentMutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly bootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly devicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly tasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * agent를 삭제하고 queued/running assignment를 강제로 정리합니다
 * 계약 generated path: `aiAgent.agents.delete`
 * 검색용 generated 경로: `agents.delete`
 * 접근 예시: `riido.aiAgent.agents.delete`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface DeleteAIAgentEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: DeleteAIAgentPathParams, options?: RiidoRequestOptions) => Promise<DeleteAgentResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<DeleteAgentResponse, DeleteAIAgentMutationVariables>) => ReturnType<typeof deleteAIAgentMutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<DeleteAgentResponse, DeleteAIAgentMutationVariables>) => ReturnType<typeof deleteAIAgentMutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly bootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly devicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.agents.editability` cache tag를 무효화합니다.
     */
    readonly agentsEditability: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly tasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.tasks.threads` cache tag를 무효화합니다.
     */
    readonly tasksThreads: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다
 * 계약 generated path: `aiAgent.agents.updateConfiguration`
 * 검색용 generated 경로: `agents.updateConfiguration`
 * 접근 예시: `riido.aiAgent.agents.updateConfiguration`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface UpdateAIAgentConfigurationEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: UpdateAIAgentConfigurationPathParams, body: UpdateAgentConfigurationRequest, options?: RiidoRequestOptions) => Promise<AgentClientRecordResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AgentClientRecordResponse, UpdateAIAgentConfigurationMutationVariables>) => ReturnType<typeof updateAIAgentConfigurationMutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AgentClientRecordResponse, UpdateAIAgentConfigurationMutationVariables>) => ReturnType<typeof updateAIAgentConfigurationMutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly bootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.agents.editability` cache tag를 무효화합니다.
     */
    readonly agentsEditability: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly tasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다
 * 계약 generated path: `aiAgent.agents.daemon.details`
 * 검색용 generated 경로: `agents.daemon.details`
 * 접근 예시: `riido.aiAgent.agents.daemon.details`
 * cache tag: `aiAgent.agents.daemon`
 */
export interface GetAIAgentDaemonEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: GetAIAgentDaemonPathParams, options?: RiidoRequestOptions) => Promise<DeviceDaemonDetailResponse>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: (params: GetAIAgentDaemonPathParams) => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (params: GetAIAgentDaemonPathParams, options?: RiidoQueryOptions<DeviceDaemonDetailResponse>) => ReturnType<typeof getAIAgentDaemonQueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (params: GetAIAgentDaemonPathParams, options?: RiidoQueryOptions<DeviceDaemonDetailResponse>) => ReturnType<typeof getAIAgentDaemonQueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient, params: GetAIAgentDaemonPathParams) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, params: GetAIAgentDaemonPathParams, options?: RiidoQueryOptions<DeviceDaemonDetailResponse>) => Promise<void>;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다
 * 계약 generated path: `aiAgent.agents.daemon.restart`
 * 검색용 generated 경로: `agents.daemon.restart`
 * 접근 예시: `riido.aiAgent.agents.daemon.restart`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface RestartAIAgentDaemonEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: RestartAIAgentDaemonPathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => Promise<DeviceDaemonCommandResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDaemonMutationVariables>) => ReturnType<typeof restartAIAgentDaemonMutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDaemonMutationVariables>) => ReturnType<typeof restartAIAgentDaemonMutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `aiAgent.agents.daemon` cache tag를 무효화합니다.
     */
    readonly agentsDaemon: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly devicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다
 * 계약 generated path: `aiAgent.agents.daemon.start`
 * 검색용 generated 경로: `agents.daemon.start`
 * 접근 예시: `riido.aiAgent.agents.daemon.start`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface StartAIAgentDaemonEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: StartAIAgentDaemonPathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => Promise<DeviceDaemonCommandResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDaemonMutationVariables>) => ReturnType<typeof startAIAgentDaemonMutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDaemonMutationVariables>) => ReturnType<typeof startAIAgentDaemonMutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `aiAgent.agents.daemon` cache tag를 무효화합니다.
     */
    readonly agentsDaemon: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly devicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다
 * 계약 generated path: `aiAgent.agents.daemon.stop`
 * 검색용 generated 경로: `agents.daemon.stop`
 * 접근 예시: `riido.aiAgent.agents.daemon.stop`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface StopAIAgentDaemonEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: StopAIAgentDaemonPathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => Promise<DeviceDaemonCommandResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDaemonMutationVariables>) => ReturnType<typeof stopAIAgentDaemonMutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDaemonMutationVariables>) => ReturnType<typeof stopAIAgentDaemonMutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `aiAgent.agents.daemon` cache tag를 무효화합니다.
     */
    readonly agentsDaemon: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly devicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * client control 활성화 전에 agent 수정 가능 여부를 조회합니다
 * 계약 generated path: `aiAgent.agents.editability`
 * 검색용 generated 경로: `agents.editability`
 * 접근 예시: `riido.aiAgent.agents.editability`
 * cache tag: `aiAgent.agents.editability`
 */
export interface GetAIAgentEditabilityEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: GetAIAgentEditabilityPathParams, options?: RiidoRequestOptions) => Promise<AgentEditabilityResponse>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: (params: GetAIAgentEditabilityPathParams) => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (params: GetAIAgentEditabilityPathParams, options?: RiidoQueryOptions<AgentEditabilityResponse>) => ReturnType<typeof getAIAgentEditabilityQueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (params: GetAIAgentEditabilityPathParams, options?: RiidoQueryOptions<AgentEditabilityResponse>) => ReturnType<typeof getAIAgentEditabilityQueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient, params: GetAIAgentEditabilityPathParams) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, params: GetAIAgentEditabilityPathParams, options?: RiidoQueryOptions<AgentEditabilityResponse>) => Promise<void>;
}

/**
 * web 또는 desktop webview client의 AI Agent 설정/온보딩 초기 데이터와 bootstrap.agents[] agent list를 조회합니다
 * 계약 generated path: `aiAgent.bootstrap`
 * 검색용 generated 경로: `bootstrap`
 * 접근 예시: `riido.aiAgent.bootstrap`
 * cache tag: `aiAgent.bootstrap`
 */
export interface GetAIAgentClientBootstrapEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (options?: RiidoRequestOptions) => Promise<ClientBootstrapResponse>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: () => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (options?: RiidoQueryOptions<ClientBootstrapResponse>) => ReturnType<typeof getAIAgentClientBootstrapQueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (options?: RiidoQueryOptions<ClientBootstrapResponse>) => ReturnType<typeof getAIAgentClientBootstrapQueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, options?: RiidoQueryOptions<ClientBootstrapResponse>) => Promise<void>;
}

/**
 * 권한이 확인된 principal의 device runtime 상태를 조회합니다
 * 계약 generated path: `aiAgent.devices.runtimes`
 * 검색용 generated 경로: `devices.runtimes`
 * 접근 예시: `riido.aiAgent.devices.runtimes`
 * cache tag: `aiAgent.devices.runtimes`
 */
export interface ListAIAgentDeviceRuntimesEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (options?: RiidoRequestOptions) => Promise<DeviceRuntimeListResponse>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: () => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (options?: RiidoQueryOptions<DeviceRuntimeListResponse>) => ReturnType<typeof listAIAgentDeviceRuntimesQueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (options?: RiidoQueryOptions<DeviceRuntimeListResponse>) => ReturnType<typeof listAIAgentDeviceRuntimesQueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, options?: RiidoQueryOptions<DeviceRuntimeListResponse>) => Promise<void>;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 상세를 조회합니다
 * 계약 generated path: `aiAgent.devices.daemon.details`
 * 검색용 generated 경로: `devices.daemon.details`
 * 접근 예시: `riido.aiAgent.devices.daemon.details`
 * cache tag: `aiAgent.devices.daemon`
 */
export interface GetAIAgentDeviceDaemonEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: GetAIAgentDeviceDaemonPathParams, options?: RiidoRequestOptions) => Promise<DeviceDaemonDetailResponse>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: (params: GetAIAgentDeviceDaemonPathParams) => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (params: GetAIAgentDeviceDaemonPathParams, options?: RiidoQueryOptions<DeviceDaemonDetailResponse>) => ReturnType<typeof getAIAgentDeviceDaemonQueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (params: GetAIAgentDeviceDaemonPathParams, options?: RiidoQueryOptions<DeviceDaemonDetailResponse>) => ReturnType<typeof getAIAgentDeviceDaemonQueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient, params: GetAIAgentDeviceDaemonPathParams) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, params: GetAIAgentDeviceDaemonPathParams, options?: RiidoQueryOptions<DeviceDaemonDetailResponse>) => Promise<void>;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 재시작을 요청합니다
 * 계약 generated path: `aiAgent.devices.daemon.restart`
 * 검색용 generated 경로: `devices.daemon.restart`
 * 접근 예시: `riido.aiAgent.devices.daemon.restart`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface RestartAIAgentDeviceDaemonEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: RestartAIAgentDeviceDaemonPathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => Promise<DeviceDaemonCommandResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDeviceDaemonMutationVariables>) => ReturnType<typeof restartAIAgentDeviceDaemonMutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDeviceDaemonMutationVariables>) => ReturnType<typeof restartAIAgentDeviceDaemonMutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `aiAgent.devices.daemon` cache tag를 무효화합니다.
     */
    readonly devicesDaemon: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly devicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 시작을 요청합니다
 * 계약 generated path: `aiAgent.devices.daemon.start`
 * 검색용 generated 경로: `devices.daemon.start`
 * 접근 예시: `riido.aiAgent.devices.daemon.start`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface StartAIAgentDeviceDaemonEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: StartAIAgentDeviceDaemonPathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => Promise<DeviceDaemonCommandResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDeviceDaemonMutationVariables>) => ReturnType<typeof startAIAgentDeviceDaemonMutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDeviceDaemonMutationVariables>) => ReturnType<typeof startAIAgentDeviceDaemonMutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `aiAgent.devices.daemon` cache tag를 무효화합니다.
     */
    readonly devicesDaemon: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly devicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 중지를 요청합니다
 * 계약 generated path: `aiAgent.devices.daemon.stop`
 * 검색용 generated 경로: `devices.daemon.stop`
 * 접근 예시: `riido.aiAgent.devices.daemon.stop`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface StopAIAgentDeviceDaemonEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: StopAIAgentDeviceDaemonPathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => Promise<DeviceDaemonCommandResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDeviceDaemonMutationVariables>) => ReturnType<typeof stopAIAgentDeviceDaemonMutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDeviceDaemonMutationVariables>) => ReturnType<typeof stopAIAgentDeviceDaemonMutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `aiAgent.devices.daemon` cache tag를 무효화합니다.
     */
    readonly devicesDaemon: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly devicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다
 * 계약 generated path: `aiAgent.events.stream`
 * 검색용 generated 경로: `events.stream`
 * 접근 예시: `riido.aiAgent.events.stream`
 * cache tag: `aiAgent.events.stream`
 */
export interface StreamAIAgentClientEventsEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (options?: RiidoRequestOptions) => Promise<Response>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: () => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (options?: RiidoQueryOptions<Response>) => ReturnType<typeof streamAIAgentClientEventsQueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (options?: RiidoQueryOptions<Response>) => ReturnType<typeof streamAIAgentClientEventsQueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, options?: RiidoQueryOptions<Response>) => Promise<void>;
}

/**
 * AI Agent 온보딩에서 사용할 서버 제공 fixture 목록을 조회합니다
 * 계약 generated path: `aiAgent.onboarding.fixtures`
 * 검색용 generated 경로: `onboarding.fixtures`
 * 접근 예시: `riido.aiAgent.onboarding.fixtures`
 * cache tag: `aiAgent.onboarding.fixtures`
 */
export interface ListAIAgentOnboardingFixturesEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (options?: RiidoRequestOptions) => Promise<AgentOnboardingFixtureListResponse>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: () => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (options?: RiidoQueryOptions<AgentOnboardingFixtureListResponse>) => ReturnType<typeof listAIAgentOnboardingFixturesQueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (options?: RiidoQueryOptions<AgentOnboardingFixtureListResponse>) => ReturnType<typeof listAIAgentOnboardingFixturesQueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, options?: RiidoQueryOptions<AgentOnboardingFixtureListResponse>) => Promise<void>;
}

/**
 * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다
 * 계약 generated path: `aiAgent.onboarding.fixtures.createAgent`
 * 검색용 generated 경로: `onboarding.fixtures.createAgent`
 * 접근 예시: `riido.aiAgent.onboarding.fixtures.createAgent`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface CreateAIAgentFromOnboardingFixtureEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: CreateAIAgentFromOnboardingFixturePathParams, body: CreateAgentConfigurationRequest, options?: RiidoRequestOptions) => Promise<AgentClientRecordResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AgentClientRecordResponse, CreateAIAgentFromOnboardingFixtureMutationVariables>) => ReturnType<typeof createAIAgentFromOnboardingFixtureMutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AgentClientRecordResponse, CreateAIAgentFromOnboardingFixtureMutationVariables>) => ReturnType<typeof createAIAgentFromOnboardingFixtureMutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly bootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly devicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly tasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다
 * 계약 generated path: `aiAgent.tasks.assignableAgents`
 * 검색용 generated 경로: `tasks.assignableAgents`
 * 접근 예시: `riido.aiAgent.tasks.assignableAgents`
 * cache tag: `aiAgent.tasks.assignableAgents`
 */
export interface ListAIAgentTaskAssignableAgentsEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: ListAIAgentTaskAssignableAgentsPathParams, options?: RiidoRequestOptions) => Promise<AgentClientListResponse>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: (params: ListAIAgentTaskAssignableAgentsPathParams) => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (params: ListAIAgentTaskAssignableAgentsPathParams, options?: RiidoQueryOptions<AgentClientListResponse>) => ReturnType<typeof listAIAgentTaskAssignableAgentsQueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (params: ListAIAgentTaskAssignableAgentsPathParams, options?: RiidoQueryOptions<AgentClientListResponse>) => ReturnType<typeof listAIAgentTaskAssignableAgentsQueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient, params: ListAIAgentTaskAssignableAgentsPathParams) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, params: ListAIAgentTaskAssignableAgentsPathParams, options?: RiidoQueryOptions<AgentClientListResponse>) => Promise<void>;
}

/**
 * task participant에서 AI agent를 제거하고 작업을 중단합니다
 * 계약 generated path: `aiAgent.tasks.unassign`
 * 검색용 generated 경로: `tasks.unassign`
 * 접근 예시: `riido.aiAgent.tasks.unassign`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface UnassignAIAgentTaskEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: UnassignAIAgentTaskPathParams, body: UnassignAIAgentTaskRequest, options?: RiidoRequestOptions) => Promise<AIAgentTaskActionResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, UnassignAIAgentTaskMutationVariables>) => ReturnType<typeof unassignAIAgentTaskMutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, UnassignAIAgentTaskMutationVariables>) => ReturnType<typeof unassignAIAgentTaskMutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly bootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly tasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.tasks.threads` cache tag를 무효화합니다.
     */
    readonly tasksThreads: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * task participant dropdown에서 AI agent를 배정합니다
 * 계약 generated path: `aiAgent.tasks.assign`
 * 검색용 generated 경로: `tasks.assign`
 * 접근 예시: `riido.aiAgent.tasks.assign`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface AssignAIAgentTaskEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: AssignAIAgentTaskPathParams, body: AssignAIAgentTaskRequest, options?: RiidoRequestOptions) => Promise<AIAgentTaskActionResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, AssignAIAgentTaskMutationVariables>) => ReturnType<typeof assignAIAgentTaskMutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, AssignAIAgentTaskMutationVariables>) => ReturnType<typeof assignAIAgentTaskMutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly bootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly tasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.tasks.threads` cache tag를 무효화합니다.
     */
    readonly tasksThreads: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * 호환 task comment route로 AI agent에게 메시지를 전달합니다
 * 계약 generated path: `aiAgent.tasks.submitComment`
 * 검색용 generated 경로: `tasks.submitComment`
 * 접근 예시: `riido.aiAgent.tasks.submitComment`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface SubmitAIAgentTaskCommentEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: SubmitAIAgentTaskCommentPathParams, body: SubmitAIAgentTaskCommentRequest, options?: RiidoRequestOptions) => Promise<AIAgentTaskActionResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, SubmitAIAgentTaskCommentMutationVariables>) => ReturnType<typeof submitAIAgentTaskCommentMutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, SubmitAIAgentTaskCommentMutationVariables>) => ReturnType<typeof submitAIAgentTaskCommentMutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly bootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly tasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.tasks.threads` cache tag를 무효화합니다.
     */
    readonly tasksThreads: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * task thread의 stop action으로 AI agent 작업을 중단합니다
 * 계약 generated path: `aiAgent.tasks.stop`
 * 검색용 generated 경로: `tasks.stop`
 * 접근 예시: `riido.aiAgent.tasks.stop`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface StopAIAgentTaskEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: StopAIAgentTaskPathParams, body: StopAIAgentTaskRequest, options?: RiidoRequestOptions) => Promise<AIAgentTaskActionResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, StopAIAgentTaskMutationVariables>) => ReturnType<typeof stopAIAgentTaskMutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, StopAIAgentTaskMutationVariables>) => ReturnType<typeof stopAIAgentTaskMutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly bootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly tasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.tasks.threads` cache tag를 무효화합니다.
     */
    readonly tasksThreads: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다
 * 계약 generated path: `aiAgent.tasks.threads`
 * 검색용 generated 경로: `tasks.threads`
 * 접근 예시: `riido.aiAgent.tasks.threads`
 * cache tag: `aiAgent.tasks.threads`
 */
export interface ListAIAgentTaskThreadsEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: ListAIAgentTaskThreadsPathParams, options?: RiidoRequestOptions) => Promise<AIAgentTaskThreadCollectionResponse>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: (params: ListAIAgentTaskThreadsPathParams) => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (params: ListAIAgentTaskThreadsPathParams, options?: RiidoQueryOptions<AIAgentTaskThreadCollectionResponse>) => ReturnType<typeof listAIAgentTaskThreadsQueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (params: ListAIAgentTaskThreadsPathParams, options?: RiidoQueryOptions<AIAgentTaskThreadCollectionResponse>) => ReturnType<typeof listAIAgentTaskThreadsQueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient, params: ListAIAgentTaskThreadsPathParams) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, params: ListAIAgentTaskThreadsPathParams, options?: RiidoQueryOptions<AIAgentTaskThreadCollectionResponse>) => Promise<void>;
}

/**
 * task thread message로 AI agent에게 다음 작업 지시를 전달합니다
 * 계약 generated path: `aiAgent.tasks.threadMessages.create`
 * 검색용 generated 경로: `tasks.threadMessages.create`
 * 접근 예시: `riido.aiAgent.tasks.threadMessages.create`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface CreateAIAgentTaskThreadMessageEndpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: CreateAIAgentTaskThreadMessagePathParams, body: CreateAIAgentTaskThreadMessageRequest, options?: RiidoRequestOptions) => Promise<AIAgentTaskActionResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, CreateAIAgentTaskThreadMessageMutationVariables>) => ReturnType<typeof createAIAgentTaskThreadMessageMutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, CreateAIAgentTaskThreadMessageMutationVariables>) => ReturnType<typeof createAIAgentTaskThreadMessageMutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly bootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly tasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * `aiAgent.tasks.threads` cache tag를 무효화합니다.
     */
    readonly tasksThreads: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * Figma agent 추가 화면에서 AI agent 설정을 생성합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.agents.create`
 * 검색용 generated 경로: `aiAgent.agents.create`
 * 접근 예시: `riido.v2.aiAgent.agents.create`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface CreateAIAgentV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: CreateAIAgentV2PathParams, body: CreateAgentConfigurationRequest, options?: RiidoRequestOptions) => Promise<AgentClientRecordResponseV2>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AgentClientRecordResponseV2, CreateAIAgentV2MutationVariables>) => ReturnType<typeof createAIAgentV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AgentClientRecordResponseV2, CreateAIAgentV2MutationVariables>) => ReturnType<typeof createAIAgentV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly aiAgentBootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly aiAgentDevicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * agent를 삭제하고 queued/running assignment를 강제로 정리합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.agents.delete`
 * 검색용 generated 경로: `aiAgent.agents.delete`
 * 접근 예시: `riido.v2.aiAgent.agents.delete`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface DeleteAIAgentV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: DeleteAIAgentV2PathParams, options?: RiidoRequestOptions) => Promise<DeleteAgentResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<DeleteAgentResponse, DeleteAIAgentV2MutationVariables>) => ReturnType<typeof deleteAIAgentV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<DeleteAgentResponse, DeleteAIAgentV2MutationVariables>) => ReturnType<typeof deleteAIAgentV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly aiAgentBootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly aiAgentDevicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.agents.editability` cache tag를 무효화합니다.
     */
    readonly aiAgentAgentsEditability: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.threads` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksThreads: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.assignedAgentProfiles` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksAssignedAgentProfiles: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.agents.updateConfiguration`
 * 검색용 generated 경로: `aiAgent.agents.updateConfiguration`
 * 접근 예시: `riido.v2.aiAgent.agents.updateConfiguration`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface UpdateAIAgentConfigurationV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: UpdateAIAgentConfigurationV2PathParams, body: UpdateAgentConfigurationRequest, options?: RiidoRequestOptions) => Promise<AgentClientRecordResponseV2>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AgentClientRecordResponseV2, UpdateAIAgentConfigurationV2MutationVariables>) => ReturnType<typeof updateAIAgentConfigurationV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AgentClientRecordResponseV2, UpdateAIAgentConfigurationV2MutationVariables>) => ReturnType<typeof updateAIAgentConfigurationV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly aiAgentBootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.agents.editability` cache tag를 무효화합니다.
     */
    readonly aiAgentAgentsEditability: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.agents.daemon.details`
 * 검색용 generated 경로: `aiAgent.agents.daemon.details`
 * 접근 예시: `riido.v2.aiAgent.agents.daemon.details`
 * cache tag: `v2.aiAgent.agents.daemon`
 */
export interface GetAIAgentDaemonV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: GetAIAgentDaemonV2PathParams, options?: RiidoRequestOptions) => Promise<DeviceDaemonDetailResponse>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: (params: GetAIAgentDaemonV2PathParams) => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (params: GetAIAgentDaemonV2PathParams, options?: RiidoQueryOptions<DeviceDaemonDetailResponse>) => ReturnType<typeof getAIAgentDaemonV2QueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (params: GetAIAgentDaemonV2PathParams, options?: RiidoQueryOptions<DeviceDaemonDetailResponse>) => ReturnType<typeof getAIAgentDaemonV2QueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient, params: GetAIAgentDaemonV2PathParams) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, params: GetAIAgentDaemonV2PathParams, options?: RiidoQueryOptions<DeviceDaemonDetailResponse>) => Promise<void>;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.agents.daemon.restart`
 * 검색용 generated 경로: `aiAgent.agents.daemon.restart`
 * 접근 예시: `riido.v2.aiAgent.agents.daemon.restart`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface RestartAIAgentDaemonV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: RestartAIAgentDaemonV2PathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => Promise<DeviceDaemonCommandResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDaemonV2MutationVariables>) => ReturnType<typeof restartAIAgentDaemonV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDaemonV2MutationVariables>) => ReturnType<typeof restartAIAgentDaemonV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.agents.daemon` cache tag를 무효화합니다.
     */
    readonly aiAgentAgentsDaemon: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly aiAgentDevicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.agents.daemon.start`
 * 검색용 generated 경로: `aiAgent.agents.daemon.start`
 * 접근 예시: `riido.v2.aiAgent.agents.daemon.start`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface StartAIAgentDaemonV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: StartAIAgentDaemonV2PathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => Promise<DeviceDaemonCommandResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDaemonV2MutationVariables>) => ReturnType<typeof startAIAgentDaemonV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDaemonV2MutationVariables>) => ReturnType<typeof startAIAgentDaemonV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.agents.daemon` cache tag를 무효화합니다.
     */
    readonly aiAgentAgentsDaemon: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly aiAgentDevicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.agents.daemon.stop`
 * 검색용 generated 경로: `aiAgent.agents.daemon.stop`
 * 접근 예시: `riido.v2.aiAgent.agents.daemon.stop`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface StopAIAgentDaemonV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: StopAIAgentDaemonV2PathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => Promise<DeviceDaemonCommandResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDaemonV2MutationVariables>) => ReturnType<typeof stopAIAgentDaemonV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDaemonV2MutationVariables>) => ReturnType<typeof stopAIAgentDaemonV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.agents.daemon` cache tag를 무효화합니다.
     */
    readonly aiAgentAgentsDaemon: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly aiAgentDevicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * client control 활성화 전에 agent 수정 가능 여부를 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.agents.editability`
 * 검색용 generated 경로: `aiAgent.agents.editability`
 * 접근 예시: `riido.v2.aiAgent.agents.editability`
 * cache tag: `v2.aiAgent.agents.editability`
 */
export interface GetAIAgentEditabilityV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: GetAIAgentEditabilityV2PathParams, options?: RiidoRequestOptions) => Promise<AgentEditabilityResponse>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: (params: GetAIAgentEditabilityV2PathParams) => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (params: GetAIAgentEditabilityV2PathParams, options?: RiidoQueryOptions<AgentEditabilityResponse>) => ReturnType<typeof getAIAgentEditabilityV2QueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (params: GetAIAgentEditabilityV2PathParams, options?: RiidoQueryOptions<AgentEditabilityResponse>) => ReturnType<typeof getAIAgentEditabilityV2QueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient, params: GetAIAgentEditabilityV2PathParams) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, params: GetAIAgentEditabilityV2PathParams, options?: RiidoQueryOptions<AgentEditabilityResponse>) => Promise<void>;
}

/**
 * web 또는 desktop webview client의 AI Agent 설정/온보딩 초기 데이터와 v2.aiAgent.bootstrap.agents[] agent list를 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.bootstrap`
 * 검색용 generated 경로: `aiAgent.bootstrap`
 * 접근 예시: `riido.v2.aiAgent.bootstrap`
 * cache tag: `v2.aiAgent.bootstrap`
 */
export interface GetAIAgentClientBootstrapV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: GetAIAgentClientBootstrapV2PathParams, options?: RiidoRequestOptions) => Promise<ClientBootstrapResponseV2>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: (params: GetAIAgentClientBootstrapV2PathParams) => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (params: GetAIAgentClientBootstrapV2PathParams, options?: RiidoQueryOptions<ClientBootstrapResponseV2>) => ReturnType<typeof getAIAgentClientBootstrapV2QueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (params: GetAIAgentClientBootstrapV2PathParams, options?: RiidoQueryOptions<ClientBootstrapResponseV2>) => ReturnType<typeof getAIAgentClientBootstrapV2QueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient, params: GetAIAgentClientBootstrapV2PathParams) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, params: GetAIAgentClientBootstrapV2PathParams, options?: RiidoQueryOptions<ClientBootstrapResponseV2>) => Promise<void>;
}

/**
 * 권한이 확인된 principal의 device runtime 상태를 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.devices.runtimes`
 * 검색용 generated 경로: `aiAgent.devices.runtimes`
 * 접근 예시: `riido.v2.aiAgent.devices.runtimes`
 * cache tag: `v2.aiAgent.devices.runtimes`
 */
export interface ListAIAgentDeviceRuntimesV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: ListAIAgentDeviceRuntimesV2PathParams, options?: RiidoRequestOptions) => Promise<DeviceRuntimeListResponse>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: (params: ListAIAgentDeviceRuntimesV2PathParams) => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (params: ListAIAgentDeviceRuntimesV2PathParams, options?: RiidoQueryOptions<DeviceRuntimeListResponse>) => ReturnType<typeof listAIAgentDeviceRuntimesV2QueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (params: ListAIAgentDeviceRuntimesV2PathParams, options?: RiidoQueryOptions<DeviceRuntimeListResponse>) => ReturnType<typeof listAIAgentDeviceRuntimesV2QueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient, params: ListAIAgentDeviceRuntimesV2PathParams) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, params: ListAIAgentDeviceRuntimesV2PathParams, options?: RiidoQueryOptions<DeviceRuntimeListResponse>) => Promise<void>;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 상세를 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.devices.daemon.details`
 * 검색용 generated 경로: `aiAgent.devices.daemon.details`
 * 접근 예시: `riido.v2.aiAgent.devices.daemon.details`
 * cache tag: `v2.aiAgent.devices.daemon`
 */
export interface GetAIAgentDeviceDaemonV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: GetAIAgentDeviceDaemonV2PathParams, options?: RiidoRequestOptions) => Promise<DeviceDaemonDetailResponse>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: (params: GetAIAgentDeviceDaemonV2PathParams) => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (params: GetAIAgentDeviceDaemonV2PathParams, options?: RiidoQueryOptions<DeviceDaemonDetailResponse>) => ReturnType<typeof getAIAgentDeviceDaemonV2QueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (params: GetAIAgentDeviceDaemonV2PathParams, options?: RiidoQueryOptions<DeviceDaemonDetailResponse>) => ReturnType<typeof getAIAgentDeviceDaemonV2QueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient, params: GetAIAgentDeviceDaemonV2PathParams) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, params: GetAIAgentDeviceDaemonV2PathParams, options?: RiidoQueryOptions<DeviceDaemonDetailResponse>) => Promise<void>;
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 재시작을 요청합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.devices.daemon.restart`
 * 검색용 generated 경로: `aiAgent.devices.daemon.restart`
 * 접근 예시: `riido.v2.aiAgent.devices.daemon.restart`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface RestartAIAgentDeviceDaemonV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: RestartAIAgentDeviceDaemonV2PathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => Promise<DeviceDaemonCommandResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDeviceDaemonV2MutationVariables>) => ReturnType<typeof restartAIAgentDeviceDaemonV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDeviceDaemonV2MutationVariables>) => ReturnType<typeof restartAIAgentDeviceDaemonV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.devices.daemon` cache tag를 무효화합니다.
     */
    readonly aiAgentDevicesDaemon: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly aiAgentDevicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 시작을 요청합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.devices.daemon.start`
 * 검색용 generated 경로: `aiAgent.devices.daemon.start`
 * 접근 예시: `riido.v2.aiAgent.devices.daemon.start`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface StartAIAgentDeviceDaemonV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: StartAIAgentDeviceDaemonV2PathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => Promise<DeviceDaemonCommandResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDeviceDaemonV2MutationVariables>) => ReturnType<typeof startAIAgentDeviceDaemonV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDeviceDaemonV2MutationVariables>) => ReturnType<typeof startAIAgentDeviceDaemonV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.devices.daemon` cache tag를 무효화합니다.
     */
    readonly aiAgentDevicesDaemon: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly aiAgentDevicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 중지를 요청합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.devices.daemon.stop`
 * 검색용 generated 경로: `aiAgent.devices.daemon.stop`
 * 접근 예시: `riido.v2.aiAgent.devices.daemon.stop`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface StopAIAgentDeviceDaemonV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: StopAIAgentDeviceDaemonV2PathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => Promise<DeviceDaemonCommandResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDeviceDaemonV2MutationVariables>) => ReturnType<typeof stopAIAgentDeviceDaemonV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDeviceDaemonV2MutationVariables>) => ReturnType<typeof stopAIAgentDeviceDaemonV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.devices.daemon` cache tag를 무효화합니다.
     */
    readonly aiAgentDevicesDaemon: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly aiAgentDevicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.events.stream`
 * 검색용 generated 경로: `aiAgent.events.stream`
 * 접근 예시: `riido.v2.aiAgent.events.stream`
 * cache tag: `v2.aiAgent.events.stream`
 */
export interface StreamAIAgentClientEventsV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: StreamAIAgentClientEventsV2PathParams, options?: RiidoRequestOptions) => Promise<Response>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: (params: StreamAIAgentClientEventsV2PathParams) => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (params: StreamAIAgentClientEventsV2PathParams, options?: RiidoQueryOptions<Response>) => ReturnType<typeof streamAIAgentClientEventsV2QueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (params: StreamAIAgentClientEventsV2PathParams, options?: RiidoQueryOptions<Response>) => ReturnType<typeof streamAIAgentClientEventsV2QueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient, params: StreamAIAgentClientEventsV2PathParams) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, params: StreamAIAgentClientEventsV2PathParams, options?: RiidoQueryOptions<Response>) => Promise<void>;
}

/**
 * AI Agent 온보딩에서 사용할 서버 제공 fixture 목록을 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.onboarding.fixtures`
 * 검색용 generated 경로: `aiAgent.onboarding.fixtures`
 * 접근 예시: `riido.v2.aiAgent.onboarding.fixtures`
 * cache tag: `v2.aiAgent.onboarding.fixtures`
 */
export interface ListAIAgentOnboardingFixturesV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: ListAIAgentOnboardingFixturesV2PathParams, options?: RiidoRequestOptions) => Promise<AgentOnboardingFixtureListResponse>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: (params: ListAIAgentOnboardingFixturesV2PathParams) => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (params: ListAIAgentOnboardingFixturesV2PathParams, options?: RiidoQueryOptions<AgentOnboardingFixtureListResponse>) => ReturnType<typeof listAIAgentOnboardingFixturesV2QueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (params: ListAIAgentOnboardingFixturesV2PathParams, options?: RiidoQueryOptions<AgentOnboardingFixtureListResponse>) => ReturnType<typeof listAIAgentOnboardingFixturesV2QueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient, params: ListAIAgentOnboardingFixturesV2PathParams) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, params: ListAIAgentOnboardingFixturesV2PathParams, options?: RiidoQueryOptions<AgentOnboardingFixtureListResponse>) => Promise<void>;
}

/**
 * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.onboarding.fixtures.createAgent`
 * 검색용 generated 경로: `aiAgent.onboarding.fixtures.createAgent`
 * 접근 예시: `riido.v2.aiAgent.onboarding.fixtures.createAgent`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface CreateAIAgentFromOnboardingFixtureV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: CreateAIAgentFromOnboardingFixtureV2PathParams, body: CreateAgentConfigurationRequest, options?: RiidoRequestOptions) => Promise<AgentClientRecordResponseV2>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AgentClientRecordResponseV2, CreateAIAgentFromOnboardingFixtureV2MutationVariables>) => ReturnType<typeof createAIAgentFromOnboardingFixtureV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AgentClientRecordResponseV2, CreateAIAgentFromOnboardingFixtureV2MutationVariables>) => ReturnType<typeof createAIAgentFromOnboardingFixtureV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly aiAgentBootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.devices.runtimes` cache tag를 무효화합니다.
     */
    readonly aiAgentDevicesRuntimes: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * workspace task/card 목록에서 현재 배정된 AI Agent profile 표시값을 component_id key 해시맵으로 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.assignedAgentProfiles`
 * 검색용 generated 경로: `aiAgent.tasks.assignedAgentProfiles`
 * 접근 예시: `riido.v2.aiAgent.tasks.assignedAgentProfiles`
 * cache tag: `v2.aiAgent.tasks.assignedAgentProfiles`
 */
export interface ListWorkspaceAssignedAgentProfilesV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: ListWorkspaceAssignedAgentProfilesV2PathParams, options?: RiidoRequestOptions) => Promise<AssignedAgentProfileMapResponse>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: (params: ListWorkspaceAssignedAgentProfilesV2PathParams) => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (params: ListWorkspaceAssignedAgentProfilesV2PathParams, options?: RiidoQueryOptions<AssignedAgentProfileMapResponse>) => ReturnType<typeof listWorkspaceAssignedAgentProfilesV2QueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (params: ListWorkspaceAssignedAgentProfilesV2PathParams, options?: RiidoQueryOptions<AssignedAgentProfileMapResponse>) => ReturnType<typeof listWorkspaceAssignedAgentProfilesV2QueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient, params: ListWorkspaceAssignedAgentProfilesV2PathParams) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, params: ListWorkspaceAssignedAgentProfilesV2PathParams, options?: RiidoQueryOptions<AssignedAgentProfileMapResponse>) => Promise<void>;
}

/**
 * task에 AI agent를 additive 방식으로 배정합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.agentAssignments.create`
 * 검색용 generated 경로: `aiAgent.tasks.agentAssignments.create`
 * 접근 예시: `riido.v2.aiAgent.tasks.agentAssignments.create`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface CreateAIAgentTaskAgentAssignmentV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: CreateAIAgentTaskAgentAssignmentV2PathParams, body: AssignAIAgentTaskRequest, options?: RiidoRequestOptions) => Promise<AIAgentTaskActionResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, CreateAIAgentTaskAgentAssignmentV2MutationVariables>) => ReturnType<typeof createAIAgentTaskAgentAssignmentV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, CreateAIAgentTaskAgentAssignmentV2MutationVariables>) => ReturnType<typeof createAIAgentTaskAgentAssignmentV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly aiAgentBootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.threads` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksThreads: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.threadStreamSubscription` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksThreadStreamSubscription: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * task의 특정 AI agent assignment를 제거하고 해당 agent 작업만 중단합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.agentAssignments.delete`
 * 검색용 generated 경로: `aiAgent.tasks.agentAssignments.delete`
 * 접근 예시: `riido.v2.aiAgent.tasks.agentAssignments.delete`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface DeleteAIAgentTaskAgentAssignmentV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: DeleteAIAgentTaskAgentAssignmentV2PathParams, options?: RiidoRequestOptions) => Promise<AIAgentTaskActionResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, DeleteAIAgentTaskAgentAssignmentV2MutationVariables>) => ReturnType<typeof deleteAIAgentTaskAgentAssignmentV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, DeleteAIAgentTaskAgentAssignmentV2MutationVariables>) => ReturnType<typeof deleteAIAgentTaskAgentAssignmentV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly aiAgentBootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.threads` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksThreads: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.threadStreamSubscription` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksThreadStreamSubscription: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * task의 특정 AI agent 작업을 agent_id 기준으로 중단합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.agentAssignments.stop`
 * 검색용 generated 경로: `aiAgent.tasks.agentAssignments.stop`
 * 접근 예시: `riido.v2.aiAgent.tasks.agentAssignments.stop`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface StopAIAgentTaskAgentAssignmentV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: StopAIAgentTaskAgentAssignmentV2PathParams, options?: RiidoRequestOptions) => Promise<AIAgentTaskActionResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, StopAIAgentTaskAgentAssignmentV2MutationVariables>) => ReturnType<typeof stopAIAgentTaskAgentAssignmentV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, StopAIAgentTaskAgentAssignmentV2MutationVariables>) => ReturnType<typeof stopAIAgentTaskAgentAssignmentV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly aiAgentBootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.threads` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksThreads: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.threadStreamSubscription` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksThreadStreamSubscription: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.assignableAgents`
 * 검색용 generated 경로: `aiAgent.tasks.assignableAgents`
 * 접근 예시: `riido.v2.aiAgent.tasks.assignableAgents`
 * cache tag: `v2.aiAgent.tasks.assignableAgents`
 */
export interface ListAIAgentTaskAssignableAgentsV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: ListAIAgentTaskAssignableAgentsV2PathParams, options?: RiidoRequestOptions) => Promise<AgentClientListResponseV2>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: (params: ListAIAgentTaskAssignableAgentsV2PathParams) => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (params: ListAIAgentTaskAssignableAgentsV2PathParams, options?: RiidoQueryOptions<AgentClientListResponseV2>) => ReturnType<typeof listAIAgentTaskAssignableAgentsV2QueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (params: ListAIAgentTaskAssignableAgentsV2PathParams, options?: RiidoQueryOptions<AgentClientListResponseV2>) => ReturnType<typeof listAIAgentTaskAssignableAgentsV2QueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient, params: ListAIAgentTaskAssignableAgentsV2PathParams) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, params: ListAIAgentTaskAssignableAgentsV2PathParams, options?: RiidoQueryOptions<AgentClientListResponseV2>) => Promise<void>;
}

/**
 * task participant에서 AI agent를 제거하고 작업을 중단합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.unassign`
 * 검색용 generated 경로: `aiAgent.tasks.unassign`
 * 접근 예시: `riido.v2.aiAgent.tasks.unassign`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface UnassignAIAgentTaskV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: UnassignAIAgentTaskV2PathParams, body: UnassignAIAgentTaskRequest, options?: RiidoRequestOptions) => Promise<AIAgentTaskActionResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, UnassignAIAgentTaskV2MutationVariables>) => ReturnType<typeof unassignAIAgentTaskV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, UnassignAIAgentTaskV2MutationVariables>) => ReturnType<typeof unassignAIAgentTaskV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly aiAgentBootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.threads` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksThreads: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.assignedAgentProfiles` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksAssignedAgentProfiles: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * task participant dropdown에서 AI agent를 배정합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.assign`
 * 검색용 generated 경로: `aiAgent.tasks.assign`
 * 접근 예시: `riido.v2.aiAgent.tasks.assign`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface AssignAIAgentTaskV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: AssignAIAgentTaskV2PathParams, body: AssignAIAgentTaskRequest, options?: RiidoRequestOptions) => Promise<AIAgentTaskActionResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, AssignAIAgentTaskV2MutationVariables>) => ReturnType<typeof assignAIAgentTaskV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, AssignAIAgentTaskV2MutationVariables>) => ReturnType<typeof assignAIAgentTaskV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly aiAgentBootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.threads` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksThreads: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.assignedAgentProfiles` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksAssignedAgentProfiles: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * 호환 task comment route로 AI agent에게 메시지를 전달합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.submitComment`
 * 검색용 generated 경로: `aiAgent.tasks.submitComment`
 * 접근 예시: `riido.v2.aiAgent.tasks.submitComment`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface SubmitAIAgentTaskCommentV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: SubmitAIAgentTaskCommentV2PathParams, body: SubmitAIAgentTaskCommentRequest, options?: RiidoRequestOptions) => Promise<AIAgentTaskActionResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, SubmitAIAgentTaskCommentV2MutationVariables>) => ReturnType<typeof submitAIAgentTaskCommentV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, SubmitAIAgentTaskCommentV2MutationVariables>) => ReturnType<typeof submitAIAgentTaskCommentV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly aiAgentBootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.threads` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksThreads: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.assignedAgentProfiles` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksAssignedAgentProfiles: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * task thread의 stop action으로 AI agent 작업을 중단합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.stop`
 * 검색용 generated 경로: `aiAgent.tasks.stop`
 * 접근 예시: `riido.v2.aiAgent.tasks.stop`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface StopAIAgentTaskV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: StopAIAgentTaskV2PathParams, body: StopAIAgentTaskRequest, options?: RiidoRequestOptions) => Promise<AIAgentTaskActionResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, StopAIAgentTaskV2MutationVariables>) => ReturnType<typeof stopAIAgentTaskV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, StopAIAgentTaskV2MutationVariables>) => ReturnType<typeof stopAIAgentTaskV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly aiAgentBootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.threads` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksThreads: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * task의 여러 active AI Agent thread를 하나의 SSE stream으로 구독하기 위한 filter handoff를 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.threadStreamSubscription`
 * 검색용 generated 경로: `aiAgent.tasks.threadStreamSubscription`
 * 접근 예시: `riido.v2.aiAgent.tasks.threadStreamSubscription`
 * cache tag: `v2.aiAgent.tasks.threadStreamSubscription`
 */
export interface GetAIAgentTaskThreadStreamSubscriptionV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: GetAIAgentTaskThreadStreamSubscriptionV2PathParams, options?: RiidoRequestOptions) => Promise<AIAgentTaskThreadStreamSubscriptionResponse>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: (params: GetAIAgentTaskThreadStreamSubscriptionV2PathParams) => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (params: GetAIAgentTaskThreadStreamSubscriptionV2PathParams, options?: RiidoQueryOptions<AIAgentTaskThreadStreamSubscriptionResponse>) => ReturnType<typeof getAIAgentTaskThreadStreamSubscriptionV2QueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (params: GetAIAgentTaskThreadStreamSubscriptionV2PathParams, options?: RiidoQueryOptions<AIAgentTaskThreadStreamSubscriptionResponse>) => ReturnType<typeof getAIAgentTaskThreadStreamSubscriptionV2QueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient, params: GetAIAgentTaskThreadStreamSubscriptionV2PathParams) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, params: GetAIAgentTaskThreadStreamSubscriptionV2PathParams, options?: RiidoQueryOptions<AIAgentTaskThreadStreamSubscriptionResponse>) => Promise<void>;
}

/**
 * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.threads`
 * 검색용 generated 경로: `aiAgent.tasks.threads`
 * 접근 예시: `riido.v2.aiAgent.tasks.threads`
 * cache tag: `v2.aiAgent.tasks.threads`
 */
export interface ListAIAgentTaskThreadsV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: ListAIAgentTaskThreadsV2PathParams, options?: RiidoRequestOptions) => Promise<AIAgentTaskThreadCollectionResponse>;
  /**
   * 이 endpoint cache 전체를 가리키는 root query key입니다.
   */
  readonly queryKeyRoot: () => readonly unknown[];
  /**
   * 특정 호출을 가리키는 query key입니다.
   */
  readonly queryKey: (params: ListAIAgentTaskThreadsV2PathParams) => readonly unknown[];
  /**
   * useQuery에 전달할 수 있는 query option입니다.
   */
  readonly query: (params: ListAIAgentTaskThreadsV2PathParams, options?: RiidoQueryOptions<AIAgentTaskThreadCollectionResponse>) => ReturnType<typeof listAIAgentTaskThreadsV2QueryOptions>;
  /**
   * query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.
   */
  readonly queryOptions: (params: ListAIAgentTaskThreadsV2PathParams, options?: RiidoQueryOptions<AIAgentTaskThreadCollectionResponse>) => ReturnType<typeof listAIAgentTaskThreadsV2QueryOptions>;
  /**
   * 특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.
   */
  readonly invalidate: (queryClient: QueryClient, params: ListAIAgentTaskThreadsV2PathParams) => Promise<void>;
  /**
   * 이 endpoint의 root cache tag 전체를 무효화합니다.
   */
  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;
  /**
   * 현재 endpoint를 prefetch합니다.
   */
  readonly prefetch: (queryClient: QueryClient, params: ListAIAgentTaskThreadsV2PathParams, options?: RiidoQueryOptions<AIAgentTaskThreadCollectionResponse>) => Promise<void>;
}

/**
 * task thread message로 AI agent에게 다음 작업 지시를 전달합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.threadMessages.create`
 * 검색용 generated 경로: `aiAgent.tasks.threadMessages.create`
 * 접근 예시: `riido.v2.aiAgent.tasks.threadMessages.create`
 * 자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.
 */
export interface CreateAIAgentTaskThreadMessageV2Endpoint {
  /**
   * HTTP 요청을 직접 실행합니다.
   */
  readonly request: (params: CreateAIAgentTaskThreadMessageV2PathParams, body: CreateAIAgentTaskThreadMessageRequest, options?: RiidoRequestOptions) => Promise<AIAgentTaskActionResponse>;
  /**
   * 이 mutation을 구분하는 key입니다.
   */
  readonly mutationKey: () => readonly unknown[];
  /**
   * useMutation에 전달할 수 있는 mutation option입니다.
   */
  readonly mutation: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, CreateAIAgentTaskThreadMessageV2MutationVariables>) => ReturnType<typeof createAIAgentTaskThreadMessageV2MutationOptions>;
  /**
   * mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.
   */
  readonly mutationOptions: (options?: RiidoMutationOptions<AIAgentTaskActionResponse, CreateAIAgentTaskThreadMessageV2MutationVariables>) => ReturnType<typeof createAIAgentTaskThreadMessageV2MutationOptions>;
  /**
   * 이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.
   */
  readonly invalidates: {
    /**
     * `v2.aiAgent.bootstrap` cache tag를 무효화합니다.
     */
    readonly aiAgentBootstrap: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.assignableAgents` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksAssignableAgents: (queryClient: QueryClient) => Promise<void>;
    /**
     * `v2.aiAgent.tasks.threads` cache tag를 무효화합니다.
     */
    readonly aiAgentTasksThreads: (queryClient: QueryClient) => Promise<void>;
    /**
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * agent visibility/access 권한을 통해 해당 agent에 연결된 desktop local daemon 상세와 제어 command를 다루는 namespace입니다.
 */
export interface RiidoAIAgentAgentsDaemonNamespace {
  /**
   * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다
   * 계약 generated path: `aiAgent.agents.daemon.details`
   * 검색용 generated 경로: `agents.daemon.details`
   * 접근 예시: `riido.aiAgent.agents.daemon.details`
   * cache tag: `aiAgent.agents.daemon`
   */
  readonly details: GetAIAgentDaemonEndpoint;
  /**
   * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다
   * 계약 generated path: `aiAgent.agents.daemon.restart`
   * 검색용 generated 경로: `agents.daemon.restart`
   * 접근 예시: `riido.aiAgent.agents.daemon.restart`
   * invalidates: `aiAgent.agents.daemon`, `aiAgent.devices.runtimes`
   */
  readonly restart: RestartAIAgentDaemonEndpoint;
  /**
   * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다
   * 계약 generated path: `aiAgent.agents.daemon.start`
   * 검색용 generated 경로: `agents.daemon.start`
   * 접근 예시: `riido.aiAgent.agents.daemon.start`
   * invalidates: `aiAgent.agents.daemon`, `aiAgent.devices.runtimes`
   */
  readonly start: StartAIAgentDaemonEndpoint;
  /**
   * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다
   * 계약 generated path: `aiAgent.agents.daemon.stop`
   * 검색용 generated 경로: `agents.daemon.stop`
   * 접근 예시: `riido.aiAgent.agents.daemon.stop`
   * invalidates: `aiAgent.agents.daemon`, `aiAgent.devices.runtimes`
   */
  readonly stop: StopAIAgentDaemonEndpoint;
}

/**
 * agent 설정과 mutation을 다루는 namespace입니다. agent 이름은 중복될 수 있고 assigned task가 있으면 수정할 수 없습니다.
 */
export interface RiidoAIAgentAgentsNamespace {
  /**
   * Figma agent 추가 화면에서 AI agent 설정을 생성합니다
   * 계약 generated path: `aiAgent.agents.create`
   * 검색용 generated 경로: `agents.create`
   * 접근 예시: `riido.aiAgent.agents.create`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.devices.runtimes`, `aiAgent.tasks.assignableAgents`
   */
  readonly create: CreateAIAgentEndpoint;
  /**
   * agent visibility/access 권한을 통해 해당 agent에 연결된 desktop local daemon 상세와 제어 command를 다루는 namespace입니다.
   */
  readonly daemon: RiidoAIAgentAgentsDaemonNamespace;
  /**
   * agent를 삭제하고 queued/running assignment를 강제로 정리합니다
   * 계약 generated path: `aiAgent.agents.delete`
   * 검색용 generated 경로: `agents.delete`
   * 접근 예시: `riido.aiAgent.agents.delete`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.devices.runtimes`, `aiAgent.agents.editability`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly delete: DeleteAIAgentEndpoint;
  /**
   * client control 활성화 전에 agent 수정 가능 여부를 조회합니다
   * 계약 generated path: `aiAgent.agents.editability`
   * 검색용 generated 경로: `agents.editability`
   * 접근 예시: `riido.aiAgent.agents.editability`
   * cache tag: `aiAgent.agents.editability`
   */
  readonly editability: GetAIAgentEditabilityEndpoint;
  /**
   * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다
   * 계약 generated path: `aiAgent.agents.updateConfiguration`
   * 검색용 generated 경로: `agents.updateConfiguration`
   * 접근 예시: `riido.aiAgent.agents.updateConfiguration`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.agents.editability`, `aiAgent.tasks.assignableAgents`
   */
  readonly updateConfiguration: UpdateAIAgentConfigurationEndpoint;
}

/**
 * aiAgent.devices.daemon namespace입니다.
 */
export interface RiidoAIAgentDevicesDaemonNamespace {
  /**
   * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 상세를 조회합니다
   * 계약 generated path: `aiAgent.devices.daemon.details`
   * 검색용 generated 경로: `devices.daemon.details`
   * 접근 예시: `riido.aiAgent.devices.daemon.details`
   * cache tag: `aiAgent.devices.daemon`
   */
  readonly details: GetAIAgentDeviceDaemonEndpoint;
  /**
   * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 재시작을 요청합니다
   * 계약 generated path: `aiAgent.devices.daemon.restart`
   * 검색용 generated 경로: `devices.daemon.restart`
   * 접근 예시: `riido.aiAgent.devices.daemon.restart`
   * invalidates: `aiAgent.devices.daemon`, `aiAgent.devices.runtimes`
   */
  readonly restart: RestartAIAgentDeviceDaemonEndpoint;
  /**
   * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 시작을 요청합니다
   * 계약 generated path: `aiAgent.devices.daemon.start`
   * 검색용 generated 경로: `devices.daemon.start`
   * 접근 예시: `riido.aiAgent.devices.daemon.start`
   * invalidates: `aiAgent.devices.daemon`, `aiAgent.devices.runtimes`
   */
  readonly start: StartAIAgentDeviceDaemonEndpoint;
  /**
   * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 중지를 요청합니다
   * 계약 generated path: `aiAgent.devices.daemon.stop`
   * 검색용 generated 경로: `devices.daemon.stop`
   * 접근 예시: `riido.aiAgent.devices.daemon.stop`
   * invalidates: `aiAgent.devices.daemon`, `aiAgent.devices.runtimes`
   */
  readonly stop: StopAIAgentDeviceDaemonEndpoint;
}

/**
 * device와 runtime 상태를 다루는 namespace입니다.
 */
export interface RiidoAIAgentDevicesNamespace {
  /**
   * aiAgent.devices.daemon namespace입니다.
   */
  readonly daemon: RiidoAIAgentDevicesDaemonNamespace;
  /**
   * 권한이 확인된 principal의 device runtime 상태를 조회합니다
   * 계약 generated path: `aiAgent.devices.runtimes`
   * 검색용 generated 경로: `devices.runtimes`
   * 접근 예시: `riido.aiAgent.devices.runtimes`
   * cache tag: `aiAgent.devices.runtimes`
   */
  readonly runtimes: ListAIAgentDeviceRuntimesEndpoint;
}

/**
 * client가 SSE로 수신하는 상태 변경 stream namespace입니다.
 */
export interface RiidoAIAgentEventsNamespace {
  /**
   * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다
   * 계약 generated path: `aiAgent.events.stream`
   * 검색용 generated 경로: `events.stream`
   * 접근 예시: `riido.aiAgent.events.stream`
   * cache tag: `aiAgent.events.stream`
   */
  readonly stream: StreamAIAgentClientEventsEndpoint;
}

/**
 * 리도, 영실, 홍도, 지원처럼 제품이 제공하는 고정 onboarding fixture 목록과 fixture 기반 agent 생성 진입점입니다.
 */
export interface RiidoAIAgentOnboardingFixturesNamespace extends ListAIAgentOnboardingFixturesEndpoint {
  /**
   * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다
   * 계약 generated path: `aiAgent.onboarding.fixtures.createAgent`
   * 검색용 generated 경로: `onboarding.fixtures.createAgent`
   * 접근 예시: `riido.aiAgent.onboarding.fixtures.createAgent`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.devices.runtimes`, `aiAgent.tasks.assignableAgents`
   */
  readonly createAgent: CreateAIAgentFromOnboardingFixtureEndpoint;
}

/**
 * AI Agent 온보딩에서 필요한 서버 제공 초기값을 다루는 namespace입니다. 템플릿 엔티티를 만들거나 관리하지 않습니다.
 */
export interface RiidoAIAgentOnboardingNamespace {
  /**
   * 리도, 영실, 홍도, 지원처럼 제품이 제공하는 고정 onboarding fixture 목록과 fixture 기반 agent 생성 진입점입니다.
   */
  readonly fixtures: RiidoAIAgentOnboardingFixturesNamespace;
}

/**
 * task thread에 사용자가 다음 작업 지시를 남기는 정식 message command namespace입니다. Figma의 댓글 표현은 이 thread message로 투영됩니다.
 */
export interface RiidoAIAgentTasksThreadMessagesNamespace {
  /**
   * task thread message로 AI agent에게 다음 작업 지시를 전달합니다
   * 계약 generated path: `aiAgent.tasks.threadMessages.create`
   * 검색용 generated 경로: `tasks.threadMessages.create`
   * 접근 예시: `riido.aiAgent.tasks.threadMessages.create`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly create: CreateAIAgentTaskThreadMessageEndpoint;
}

/**
 * task thread에서 AI Agent assignment, thread message, compatibility comment action을 다루는 namespace입니다.
 */
export interface RiidoAIAgentTasksNamespace {
  /**
   * task participant dropdown에서 AI agent를 배정합니다
   * 계약 generated path: `aiAgent.tasks.assign`
   * 검색용 generated 경로: `tasks.assign`
   * 접근 예시: `riido.aiAgent.tasks.assign`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly assign: AssignAIAgentTaskEndpoint;
  /**
   * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다
   * 계약 generated path: `aiAgent.tasks.assignableAgents`
   * 검색용 generated 경로: `tasks.assignableAgents`
   * 접근 예시: `riido.aiAgent.tasks.assignableAgents`
   * cache tag: `aiAgent.tasks.assignableAgents`
   */
  readonly assignableAgents: ListAIAgentTaskAssignableAgentsEndpoint;
  /**
   * task thread의 stop action으로 AI agent 작업을 중단합니다
   * 계약 generated path: `aiAgent.tasks.stop`
   * 검색용 generated 경로: `tasks.stop`
   * 접근 예시: `riido.aiAgent.tasks.stop`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly stop: StopAIAgentTaskEndpoint;
  /**
   * 호환 task comment route로 AI agent에게 메시지를 전달합니다
   * 계약 generated path: `aiAgent.tasks.submitComment`
   * 검색용 generated 경로: `tasks.submitComment`
   * 접근 예시: `riido.aiAgent.tasks.submitComment`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly submitComment: SubmitAIAgentTaskCommentEndpoint;
  /**
   * task thread에 사용자가 다음 작업 지시를 남기는 정식 message command namespace입니다. Figma의 댓글 표현은 이 thread message로 투영됩니다.
   */
  readonly threadMessages: RiidoAIAgentTasksThreadMessagesNamespace;
  /**
   * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다
   * 계약 generated path: `aiAgent.tasks.threads`
   * 검색용 generated 경로: `tasks.threads`
   * 접근 예시: `riido.aiAgent.tasks.threads`
   * cache tag: `aiAgent.tasks.threads`
   */
  readonly threads: ListAIAgentTaskThreadsEndpoint;
  /**
   * task participant에서 AI agent를 제거하고 작업을 중단합니다
   * 계약 generated path: `aiAgent.tasks.unassign`
   * 검색용 generated 경로: `tasks.unassign`
   * 접근 예시: `riido.aiAgent.tasks.unassign`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly unassign: UnassignAIAgentTaskEndpoint;
}

/**
 * AI Agent client API module입니다. 화면 구성을 소유하지 않고, client가 화면을 구성할 때 필요한 API 의미, 정책, lifecycle, cache 관계를 제공합니다.
 */
export interface RiidoAIAgentModule {
  /**
   * agent 설정과 mutation을 다루는 namespace입니다. agent 이름은 중복될 수 있고 assigned task가 있으면 수정할 수 없습니다.
   */
  readonly agents: RiidoAIAgentAgentsNamespace;
  /**
   * web 또는 desktop webview client의 AI Agent 설정/온보딩 초기 데이터와 bootstrap.agents[] agent list를 조회합니다
   * 계약 generated path: `aiAgent.bootstrap`
   * 검색용 generated 경로: `bootstrap`
   * 접근 예시: `riido.aiAgent.bootstrap`
   * cache tag: `aiAgent.bootstrap`
   */
  readonly bootstrap: GetAIAgentClientBootstrapEndpoint;
  /**
   * device와 runtime 상태를 다루는 namespace입니다.
   */
  readonly devices: RiidoAIAgentDevicesNamespace;
  /**
   * client가 SSE로 수신하는 상태 변경 stream namespace입니다.
   */
  readonly events: RiidoAIAgentEventsNamespace;
  /**
   * AI Agent 온보딩에서 필요한 서버 제공 초기값을 다루는 namespace입니다. 템플릿 엔티티를 만들거나 관리하지 않습니다.
   */
  readonly onboarding: RiidoAIAgentOnboardingNamespace;
  /**
   * task thread에서 AI Agent assignment, thread message, compatibility comment action을 다루는 namespace입니다.
   */
  readonly tasks: RiidoAIAgentTasksNamespace;
}

/**
 * workspace-scoped agent 권한을 통해 daemon 상세와 제어 command를 다루는 v2 namespace입니다.
 */
export interface RiidoV2AIAgentAgentsDaemonNamespace {
  /**
   * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.agents.daemon.details`
   * 검색용 generated 경로: `aiAgent.agents.daemon.details`
   * 접근 예시: `riido.v2.aiAgent.agents.daemon.details`
   * cache tag: `v2.aiAgent.agents.daemon`
   */
  readonly details: GetAIAgentDaemonV2Endpoint;
  /**
   * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.agents.daemon.restart`
   * 검색용 generated 경로: `aiAgent.agents.daemon.restart`
   * 접근 예시: `riido.v2.aiAgent.agents.daemon.restart`
   * invalidates: `v2.aiAgent.agents.daemon`, `v2.aiAgent.devices.runtimes`
   */
  readonly restart: RestartAIAgentDaemonV2Endpoint;
  /**
   * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.agents.daemon.start`
   * 검색용 generated 경로: `aiAgent.agents.daemon.start`
   * 접근 예시: `riido.v2.aiAgent.agents.daemon.start`
   * invalidates: `v2.aiAgent.agents.daemon`, `v2.aiAgent.devices.runtimes`
   */
  readonly start: StartAIAgentDaemonV2Endpoint;
  /**
   * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.agents.daemon.stop`
   * 검색용 generated 경로: `aiAgent.agents.daemon.stop`
   * 접근 예시: `riido.v2.aiAgent.agents.daemon.stop`
   * invalidates: `v2.aiAgent.agents.daemon`, `v2.aiAgent.devices.runtimes`
   */
  readonly stop: StopAIAgentDaemonV2Endpoint;
}

/**
 * workspace 안에 생성되는 agent 설정과 mutation을 다루는 v2 namespace입니다.
 */
export interface RiidoV2AIAgentAgentsNamespace {
  /**
   * Figma agent 추가 화면에서 AI agent 설정을 생성합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.agents.create`
   * 검색용 generated 경로: `aiAgent.agents.create`
   * 접근 예시: `riido.v2.aiAgent.agents.create`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.devices.runtimes`, `v2.aiAgent.tasks.assignableAgents`
   */
  readonly create: CreateAIAgentV2Endpoint;
  /**
   * workspace-scoped agent 권한을 통해 daemon 상세와 제어 command를 다루는 v2 namespace입니다.
   */
  readonly daemon: RiidoV2AIAgentAgentsDaemonNamespace;
  /**
   * agent를 삭제하고 queued/running assignment를 강제로 정리합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.agents.delete`
   * 검색용 generated 경로: `aiAgent.agents.delete`
   * 접근 예시: `riido.v2.aiAgent.agents.delete`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.devices.runtimes`, `v2.aiAgent.agents.editability`, `v2.aiAgent.tasks.assignableAgents`, `v2.aiAgent.tasks.threads`, `v2.aiAgent.tasks.assignedAgentProfiles`
   */
  readonly delete: DeleteAIAgentV2Endpoint;
  /**
   * client control 활성화 전에 agent 수정 가능 여부를 조회합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.agents.editability`
   * 검색용 generated 경로: `aiAgent.agents.editability`
   * 접근 예시: `riido.v2.aiAgent.agents.editability`
   * cache tag: `v2.aiAgent.agents.editability`
   */
  readonly editability: GetAIAgentEditabilityV2Endpoint;
  /**
   * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.agents.updateConfiguration`
   * 검색용 generated 경로: `aiAgent.agents.updateConfiguration`
   * 접근 예시: `riido.v2.aiAgent.agents.updateConfiguration`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.agents.editability`, `v2.aiAgent.tasks.assignableAgents`
   */
  readonly updateConfiguration: UpdateAIAgentConfigurationV2Endpoint;
}

/**
 * v2.aiAgent.devices.daemon namespace입니다.
 */
export interface RiidoV2AIAgentDevicesDaemonNamespace {
  /**
   * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 상세를 조회합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.devices.daemon.details`
   * 검색용 generated 경로: `aiAgent.devices.daemon.details`
   * 접근 예시: `riido.v2.aiAgent.devices.daemon.details`
   * cache tag: `v2.aiAgent.devices.daemon`
   */
  readonly details: GetAIAgentDeviceDaemonV2Endpoint;
  /**
   * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 재시작을 요청합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.devices.daemon.restart`
   * 검색용 generated 경로: `aiAgent.devices.daemon.restart`
   * 접근 예시: `riido.v2.aiAgent.devices.daemon.restart`
   * invalidates: `v2.aiAgent.devices.daemon`, `v2.aiAgent.devices.runtimes`
   */
  readonly restart: RestartAIAgentDeviceDaemonV2Endpoint;
  /**
   * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 시작을 요청합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.devices.daemon.start`
   * 검색용 generated 경로: `aiAgent.devices.daemon.start`
   * 접근 예시: `riido.v2.aiAgent.devices.daemon.start`
   * invalidates: `v2.aiAgent.devices.daemon`, `v2.aiAgent.devices.runtimes`
   */
  readonly start: StartAIAgentDeviceDaemonV2Endpoint;
  /**
   * runtime 설정 화면의 내 기기 영역에서 device_id 기준 daemon 중지를 요청합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.devices.daemon.stop`
   * 검색용 generated 경로: `aiAgent.devices.daemon.stop`
   * 접근 예시: `riido.v2.aiAgent.devices.daemon.stop`
   * invalidates: `v2.aiAgent.devices.daemon`, `v2.aiAgent.devices.runtimes`
   */
  readonly stop: StopAIAgentDeviceDaemonV2Endpoint;
}

/**
 * account-owned device/runtime을 선택된 workspace agent 권한에 맞춰 읽는 v2 namespace입니다.
 */
export interface RiidoV2AIAgentDevicesNamespace {
  /**
   * v2.aiAgent.devices.daemon namespace입니다.
   */
  readonly daemon: RiidoV2AIAgentDevicesDaemonNamespace;
  /**
   * 권한이 확인된 principal의 device runtime 상태를 조회합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.devices.runtimes`
   * 검색용 generated 경로: `aiAgent.devices.runtimes`
   * 접근 예시: `riido.v2.aiAgent.devices.runtimes`
   * cache tag: `v2.aiAgent.devices.runtimes`
   */
  readonly runtimes: ListAIAgentDeviceRuntimesV2Endpoint;
}

/**
 * 선택된 workspace 범위로 client가 수신하는 SSE stream namespace입니다.
 */
export interface RiidoV2AIAgentEventsNamespace {
  /**
   * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.events.stream`
   * 검색용 generated 경로: `aiAgent.events.stream`
   * 접근 예시: `riido.v2.aiAgent.events.stream`
   * cache tag: `v2.aiAgent.events.stream`
   */
  readonly stream: StreamAIAgentClientEventsV2Endpoint;
}

/**
 * 서버 제공 fixture 목록과 fixture 기반 agent 생성 진입점입니다. 템플릿 엔티티를 만들거나 관리하지 않습니다.
 */
export interface RiidoV2AIAgentOnboardingFixturesNamespace extends ListAIAgentOnboardingFixturesV2Endpoint {
  /**
   * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.onboarding.fixtures.createAgent`
   * 검색용 generated 경로: `aiAgent.onboarding.fixtures.createAgent`
   * 접근 예시: `riido.v2.aiAgent.onboarding.fixtures.createAgent`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.devices.runtimes`, `v2.aiAgent.tasks.assignableAgents`
   */
  readonly createAgent: CreateAIAgentFromOnboardingFixtureV2Endpoint;
}

/**
 * 선택된 workspace에서 빠른 agent 생성을 돕는 onboarding fixture namespace입니다.
 */
export interface RiidoV2AIAgentOnboardingNamespace {
  /**
   * 서버 제공 fixture 목록과 fixture 기반 agent 생성 진입점입니다. 템플릿 엔티티를 만들거나 관리하지 않습니다.
   */
  readonly fixtures: RiidoV2AIAgentOnboardingFixturesNamespace;
}

/**
 * 선택된 workspace의 task에 여러 AI Agent를 병렬로 배정/해제/중지하는 additive assignment namespace입니다. v1/v2 tasks.assignment 호환 경로는 기존 단일 active 시연 흐름을 유지합니다.
 */
export interface RiidoV2AIAgentTasksAgentAssignmentsNamespace {
  /**
   * task에 AI agent를 additive 방식으로 배정합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.agentAssignments.create`
   * 검색용 generated 경로: `aiAgent.tasks.agentAssignments.create`
   * 접근 예시: `riido.v2.aiAgent.tasks.agentAssignments.create`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.tasks.assignableAgents`, `v2.aiAgent.tasks.threads`, `v2.aiAgent.tasks.threadStreamSubscription`
   */
  readonly create: CreateAIAgentTaskAgentAssignmentV2Endpoint;
  /**
   * task의 특정 AI agent assignment를 제거하고 해당 agent 작업만 중단합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.agentAssignments.delete`
   * 검색용 generated 경로: `aiAgent.tasks.agentAssignments.delete`
   * 접근 예시: `riido.v2.aiAgent.tasks.agentAssignments.delete`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.tasks.assignableAgents`, `v2.aiAgent.tasks.threads`, `v2.aiAgent.tasks.threadStreamSubscription`
   */
  readonly delete: DeleteAIAgentTaskAgentAssignmentV2Endpoint;
  /**
   * task의 특정 AI agent 작업을 agent_id 기준으로 중단합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.agentAssignments.stop`
   * 검색용 generated 경로: `aiAgent.tasks.agentAssignments.stop`
   * 접근 예시: `riido.v2.aiAgent.tasks.agentAssignments.stop`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.tasks.assignableAgents`, `v2.aiAgent.tasks.threads`, `v2.aiAgent.tasks.threadStreamSubscription`
   */
  readonly stop: StopAIAgentTaskAgentAssignmentV2Endpoint;
}

/**
 * 선택된 workspace의 task thread에 사용자가 다음 작업 지시를 남기는 정식 message command namespace입니다.
 */
export interface RiidoV2AIAgentTasksThreadMessagesNamespace {
  /**
   * task thread message로 AI agent에게 다음 작업 지시를 전달합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.threadMessages.create`
   * 검색용 generated 경로: `aiAgent.tasks.threadMessages.create`
   * 접근 예시: `riido.v2.aiAgent.tasks.threadMessages.create`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.tasks.assignableAgents`, `v2.aiAgent.tasks.threads`
   */
  readonly create: CreateAIAgentTaskThreadMessageV2Endpoint;
}

/**
 * 선택된 workspace의 task thread에서 AI Agent assignment, thread message, compatibility comment action을 다루는 namespace입니다.
 */
export interface RiidoV2AIAgentTasksNamespace {
  /**
   * 선택된 workspace의 task에 여러 AI Agent를 병렬로 배정/해제/중지하는 additive assignment namespace입니다. v1/v2 tasks.assignment 호환 경로는 기존 단일 active 시연 흐름을 유지합니다.
   */
  readonly agentAssignments: RiidoV2AIAgentTasksAgentAssignmentsNamespace;
  /**
   * task participant dropdown에서 AI agent를 배정합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.assign`
   * 검색용 generated 경로: `aiAgent.tasks.assign`
   * 접근 예시: `riido.v2.aiAgent.tasks.assign`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.tasks.assignableAgents`, `v2.aiAgent.tasks.threads`, `v2.aiAgent.tasks.assignedAgentProfiles`
   */
  readonly assign: AssignAIAgentTaskV2Endpoint;
  /**
   * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.assignableAgents`
   * 검색용 generated 경로: `aiAgent.tasks.assignableAgents`
   * 접근 예시: `riido.v2.aiAgent.tasks.assignableAgents`
   * cache tag: `v2.aiAgent.tasks.assignableAgents`
   */
  readonly assignableAgents: ListAIAgentTaskAssignableAgentsV2Endpoint;
  /**
   * workspace task/card 목록에서 현재 배정된 AI Agent profile 표시값을 component_id key 해시맵으로 조회합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.assignedAgentProfiles`
   * 검색용 generated 경로: `aiAgent.tasks.assignedAgentProfiles`
   * 접근 예시: `riido.v2.aiAgent.tasks.assignedAgentProfiles`
   * cache tag: `v2.aiAgent.tasks.assignedAgentProfiles`
   */
  readonly assignedAgentProfiles: ListWorkspaceAssignedAgentProfilesV2Endpoint;
  /**
   * task thread의 stop action으로 AI agent 작업을 중단합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.stop`
   * 검색용 generated 경로: `aiAgent.tasks.stop`
   * 접근 예시: `riido.v2.aiAgent.tasks.stop`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.tasks.assignableAgents`, `v2.aiAgent.tasks.threads`
   */
  readonly stop: StopAIAgentTaskV2Endpoint;
  /**
   * 호환 task comment route로 AI agent에게 메시지를 전달합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.submitComment`
   * 검색용 generated 경로: `aiAgent.tasks.submitComment`
   * 접근 예시: `riido.v2.aiAgent.tasks.submitComment`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.tasks.assignableAgents`, `v2.aiAgent.tasks.threads`, `v2.aiAgent.tasks.assignedAgentProfiles`
   */
  readonly submitComment: SubmitAIAgentTaskCommentV2Endpoint;
  /**
   * 선택된 workspace의 task thread에 사용자가 다음 작업 지시를 남기는 정식 message command namespace입니다.
   */
  readonly threadMessages: RiidoV2AIAgentTasksThreadMessagesNamespace;
  /**
   * task의 여러 active AI Agent thread를 하나의 SSE stream으로 구독하기 위한 filter handoff를 조회합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.threadStreamSubscription`
   * 검색용 generated 경로: `aiAgent.tasks.threadStreamSubscription`
   * 접근 예시: `riido.v2.aiAgent.tasks.threadStreamSubscription`
   * cache tag: `v2.aiAgent.tasks.threadStreamSubscription`
   */
  readonly threadStreamSubscription: GetAIAgentTaskThreadStreamSubscriptionV2Endpoint;
  /**
   * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.threads`
   * 검색용 generated 경로: `aiAgent.tasks.threads`
   * 접근 예시: `riido.v2.aiAgent.tasks.threads`
   * cache tag: `v2.aiAgent.tasks.threads`
   */
  readonly threads: ListAIAgentTaskThreadsV2Endpoint;
  /**
   * task participant에서 AI agent를 제거하고 작업을 중단합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.unassign`
   * 검색용 generated 경로: `aiAgent.tasks.unassign`
   * 접근 예시: `riido.v2.aiAgent.tasks.unassign`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.tasks.assignableAgents`, `v2.aiAgent.tasks.threads`, `v2.aiAgent.tasks.assignedAgentProfiles`
   */
  readonly unassign: UnassignAIAgentTaskV2Endpoint;
}

/**
 * workspace_id path parameter로 범위가 정해지는 AI Agent v2 namespace입니다.
 */
export interface RiidoV2AIAgentNamespace {
  /**
   * workspace 안에 생성되는 agent 설정과 mutation을 다루는 v2 namespace입니다.
   */
  readonly agents: RiidoV2AIAgentAgentsNamespace;
  /**
   * web 또는 desktop webview client의 AI Agent 설정/온보딩 초기 데이터와 v2.aiAgent.bootstrap.agents[] agent list를 조회합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.bootstrap`
   * 검색용 generated 경로: `aiAgent.bootstrap`
   * 접근 예시: `riido.v2.aiAgent.bootstrap`
   * cache tag: `v2.aiAgent.bootstrap`
   */
  readonly bootstrap: GetAIAgentClientBootstrapV2Endpoint;
  /**
   * account-owned device/runtime을 선택된 workspace agent 권한에 맞춰 읽는 v2 namespace입니다.
   */
  readonly devices: RiidoV2AIAgentDevicesNamespace;
  /**
   * 선택된 workspace 범위로 client가 수신하는 SSE stream namespace입니다.
   */
  readonly events: RiidoV2AIAgentEventsNamespace;
  /**
   * 선택된 workspace에서 빠른 agent 생성을 돕는 onboarding fixture namespace입니다.
   */
  readonly onboarding: RiidoV2AIAgentOnboardingNamespace;
  /**
   * 선택된 workspace의 task thread에서 AI Agent assignment, thread message, compatibility comment action을 다루는 namespace입니다.
   */
  readonly tasks: RiidoV2AIAgentTasksNamespace;
}

/**
 * v2 client API module입니다. workspace-scoped AI Agent API를 riido.v2.aiAgent.* 경로로 제공합니다. v1은 UI 테스트 호환 표면으로 유지됩니다.
 */
export interface RiidoV2Module {
  /**
   * workspace_id path parameter로 범위가 정해지는 AI Agent v2 namespace입니다.
   */
  readonly aiAgent: RiidoV2AIAgentNamespace;
}

/**
 * control-plane API를 DSL client metadata의 module/namespace별로 묶은 config-bound facade입니다.
 */
export interface RiidoControlPlaneClient {
  /**
   * AI Agent client API module입니다. 화면 구성을 소유하지 않고, client가 화면을 구성할 때 필요한 API 의미, 정책, lifecycle, cache 관계를 제공합니다.
   */
  readonly aiAgent: RiidoAIAgentModule;
  /**
   * v2 client API module입니다. workspace-scoped AI Agent API를 riido.v2.aiAgent.* 경로로 제공합니다. v1은 UI 테스트 호환 표면으로 유지됩니다.
   */
  readonly v2: RiidoV2Module;
}

/**
 * control-plane API를 DSL client metadata의 module/namespace별로 묶은 config-bound facade입니다.
 * React QueryClient를 대체하지 않고 request, query/queryOptions, mutation/mutationOptions와 명시적 cache helper만 제공합니다.
 */
export function createRiidoControlPlaneClient(config: RiidoClientConfig): RiidoControlPlaneClient {
  return {
    aiAgent: {
      agents: {
        create: {
          request: (body: CreateAgentConfigurationRequest, options?: RiidoRequestOptions) => createAIAgent(config, body, options),
          mutationKey: createAIAgentMutationKey,
          mutation: (options: RiidoMutationOptions<AgentClientRecordResponse, CreateAIAgentMutationVariables> = {}) => createAIAgentMutationOptions(config, options),
          mutationOptions: (options: RiidoMutationOptions<AgentClientRecordResponse, CreateAIAgentMutationVariables> = {}) => createAIAgentMutationOptions(config, options),
          invalidates: {
            bootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }),
            devicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() }),
            tasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() }),
            all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() })]),
          },
        },
        daemon: {
          details: {
            request: (params: GetAIAgentDaemonPathParams, options?: RiidoRequestOptions) => getAIAgentDaemon(config, params, options),
            queryKeyRoot: getAIAgentDaemonQueryKeyRoot,
            queryKey: getAIAgentDaemonQueryKey,
            query: (params: GetAIAgentDaemonPathParams, options: RiidoQueryOptions<DeviceDaemonDetailResponse> = {}) => getAIAgentDaemonQueryOptions(config, params, options),
            queryOptions: (params: GetAIAgentDaemonPathParams, options: RiidoQueryOptions<DeviceDaemonDetailResponse> = {}) => getAIAgentDaemonQueryOptions(config, params, options),
            invalidate: (queryClient: QueryClient, params: GetAIAgentDaemonPathParams) => queryClient.invalidateQueries({ queryKey: getAIAgentDaemonQueryKey(params) }),
            invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentDaemonQueryKeyRoot() }),
            prefetch: (queryClient: QueryClient, params: GetAIAgentDaemonPathParams, options?: RiidoQueryOptions<DeviceDaemonDetailResponse>) => queryClient.prefetchQuery(getAIAgentDaemonQueryOptions(config, params, options)),
          },
          restart: {
            request: (params: RestartAIAgentDaemonPathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => restartAIAgentDaemon(config, params, body, options),
            mutationKey: restartAIAgentDaemonMutationKey,
            mutation: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDaemonMutationVariables> = {}) => restartAIAgentDaemonMutationOptions(config, options),
            mutationOptions: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDaemonMutationVariables> = {}) => restartAIAgentDaemonMutationOptions(config, options),
            invalidates: {
              agentsDaemon: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentDaemonQueryKeyRoot() }),
              devicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() }),
              all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentDaemonQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() })]),
            },
          },
          start: {
            request: (params: StartAIAgentDaemonPathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => startAIAgentDaemon(config, params, body, options),
            mutationKey: startAIAgentDaemonMutationKey,
            mutation: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDaemonMutationVariables> = {}) => startAIAgentDaemonMutationOptions(config, options),
            mutationOptions: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDaemonMutationVariables> = {}) => startAIAgentDaemonMutationOptions(config, options),
            invalidates: {
              agentsDaemon: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentDaemonQueryKeyRoot() }),
              devicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() }),
              all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentDaemonQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() })]),
            },
          },
          stop: {
            request: (params: StopAIAgentDaemonPathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => stopAIAgentDaemon(config, params, body, options),
            mutationKey: stopAIAgentDaemonMutationKey,
            mutation: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDaemonMutationVariables> = {}) => stopAIAgentDaemonMutationOptions(config, options),
            mutationOptions: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDaemonMutationVariables> = {}) => stopAIAgentDaemonMutationOptions(config, options),
            invalidates: {
              agentsDaemon: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentDaemonQueryKeyRoot() }),
              devicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() }),
              all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentDaemonQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() })]),
            },
          },
        },
        delete: {
          request: (params: DeleteAIAgentPathParams, options?: RiidoRequestOptions) => deleteAIAgent(config, params, options),
          mutationKey: deleteAIAgentMutationKey,
          mutation: (options: RiidoMutationOptions<DeleteAgentResponse, DeleteAIAgentMutationVariables> = {}) => deleteAIAgentMutationOptions(config, options),
          mutationOptions: (options: RiidoMutationOptions<DeleteAgentResponse, DeleteAIAgentMutationVariables> = {}) => deleteAIAgentMutationOptions(config, options),
          invalidates: {
            bootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }),
            devicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() }),
            agentsEditability: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentEditabilityQueryKeyRoot() }),
            tasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() }),
            tasksThreads: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsQueryKeyRoot() }),
            all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: getAIAgentEditabilityQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsQueryKeyRoot() })]),
          },
        },
        editability: {
          request: (params: GetAIAgentEditabilityPathParams, options?: RiidoRequestOptions) => getAIAgentEditability(config, params, options),
          queryKeyRoot: getAIAgentEditabilityQueryKeyRoot,
          queryKey: getAIAgentEditabilityQueryKey,
          query: (params: GetAIAgentEditabilityPathParams, options: RiidoQueryOptions<AgentEditabilityResponse> = {}) => getAIAgentEditabilityQueryOptions(config, params, options),
          queryOptions: (params: GetAIAgentEditabilityPathParams, options: RiidoQueryOptions<AgentEditabilityResponse> = {}) => getAIAgentEditabilityQueryOptions(config, params, options),
          invalidate: (queryClient: QueryClient, params: GetAIAgentEditabilityPathParams) => queryClient.invalidateQueries({ queryKey: getAIAgentEditabilityQueryKey(params) }),
          invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentEditabilityQueryKeyRoot() }),
          prefetch: (queryClient: QueryClient, params: GetAIAgentEditabilityPathParams, options?: RiidoQueryOptions<AgentEditabilityResponse>) => queryClient.prefetchQuery(getAIAgentEditabilityQueryOptions(config, params, options)),
        },
        updateConfiguration: {
          request: (params: UpdateAIAgentConfigurationPathParams, body: UpdateAgentConfigurationRequest, options?: RiidoRequestOptions) => updateAIAgentConfiguration(config, params, body, options),
          mutationKey: updateAIAgentConfigurationMutationKey,
          mutation: (options: RiidoMutationOptions<AgentClientRecordResponse, UpdateAIAgentConfigurationMutationVariables> = {}) => updateAIAgentConfigurationMutationOptions(config, options),
          mutationOptions: (options: RiidoMutationOptions<AgentClientRecordResponse, UpdateAIAgentConfigurationMutationVariables> = {}) => updateAIAgentConfigurationMutationOptions(config, options),
          invalidates: {
            bootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }),
            agentsEditability: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentEditabilityQueryKeyRoot() }),
            tasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() }),
            all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: getAIAgentEditabilityQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() })]),
          },
        },
      },
      bootstrap: {
        request: (options?: RiidoRequestOptions) => getAIAgentClientBootstrap(config, options),
        queryKeyRoot: getAIAgentClientBootstrapQueryKeyRoot,
        queryKey: getAIAgentClientBootstrapQueryKey,
        query: (options: RiidoQueryOptions<ClientBootstrapResponse> = {}) => getAIAgentClientBootstrapQueryOptions(config, options),
        queryOptions: (options: RiidoQueryOptions<ClientBootstrapResponse> = {}) => getAIAgentClientBootstrapQueryOptions(config, options),
        invalidate: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKey() }),
        invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }),
        prefetch: (queryClient: QueryClient, options?: RiidoQueryOptions<ClientBootstrapResponse>) => queryClient.prefetchQuery(getAIAgentClientBootstrapQueryOptions(config, options)),
      },
      devices: {
        daemon: {
          details: {
            request: (params: GetAIAgentDeviceDaemonPathParams, options?: RiidoRequestOptions) => getAIAgentDeviceDaemon(config, params, options),
            queryKeyRoot: getAIAgentDeviceDaemonQueryKeyRoot,
            queryKey: getAIAgentDeviceDaemonQueryKey,
            query: (params: GetAIAgentDeviceDaemonPathParams, options: RiidoQueryOptions<DeviceDaemonDetailResponse> = {}) => getAIAgentDeviceDaemonQueryOptions(config, params, options),
            queryOptions: (params: GetAIAgentDeviceDaemonPathParams, options: RiidoQueryOptions<DeviceDaemonDetailResponse> = {}) => getAIAgentDeviceDaemonQueryOptions(config, params, options),
            invalidate: (queryClient: QueryClient, params: GetAIAgentDeviceDaemonPathParams) => queryClient.invalidateQueries({ queryKey: getAIAgentDeviceDaemonQueryKey(params) }),
            invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentDeviceDaemonQueryKeyRoot() }),
            prefetch: (queryClient: QueryClient, params: GetAIAgentDeviceDaemonPathParams, options?: RiidoQueryOptions<DeviceDaemonDetailResponse>) => queryClient.prefetchQuery(getAIAgentDeviceDaemonQueryOptions(config, params, options)),
          },
          restart: {
            request: (params: RestartAIAgentDeviceDaemonPathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => restartAIAgentDeviceDaemon(config, params, body, options),
            mutationKey: restartAIAgentDeviceDaemonMutationKey,
            mutation: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDeviceDaemonMutationVariables> = {}) => restartAIAgentDeviceDaemonMutationOptions(config, options),
            mutationOptions: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDeviceDaemonMutationVariables> = {}) => restartAIAgentDeviceDaemonMutationOptions(config, options),
            invalidates: {
              devicesDaemon: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentDeviceDaemonQueryKeyRoot() }),
              devicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() }),
              all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentDeviceDaemonQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() })]),
            },
          },
          start: {
            request: (params: StartAIAgentDeviceDaemonPathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => startAIAgentDeviceDaemon(config, params, body, options),
            mutationKey: startAIAgentDeviceDaemonMutationKey,
            mutation: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDeviceDaemonMutationVariables> = {}) => startAIAgentDeviceDaemonMutationOptions(config, options),
            mutationOptions: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDeviceDaemonMutationVariables> = {}) => startAIAgentDeviceDaemonMutationOptions(config, options),
            invalidates: {
              devicesDaemon: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentDeviceDaemonQueryKeyRoot() }),
              devicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() }),
              all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentDeviceDaemonQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() })]),
            },
          },
          stop: {
            request: (params: StopAIAgentDeviceDaemonPathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => stopAIAgentDeviceDaemon(config, params, body, options),
            mutationKey: stopAIAgentDeviceDaemonMutationKey,
            mutation: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDeviceDaemonMutationVariables> = {}) => stopAIAgentDeviceDaemonMutationOptions(config, options),
            mutationOptions: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDeviceDaemonMutationVariables> = {}) => stopAIAgentDeviceDaemonMutationOptions(config, options),
            invalidates: {
              devicesDaemon: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentDeviceDaemonQueryKeyRoot() }),
              devicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() }),
              all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentDeviceDaemonQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() })]),
            },
          },
        },
        runtimes: {
          request: (options?: RiidoRequestOptions) => listAIAgentDeviceRuntimes(config, options),
          queryKeyRoot: listAIAgentDeviceRuntimesQueryKeyRoot,
          queryKey: listAIAgentDeviceRuntimesQueryKey,
          query: (options: RiidoQueryOptions<DeviceRuntimeListResponse> = {}) => listAIAgentDeviceRuntimesQueryOptions(config, options),
          queryOptions: (options: RiidoQueryOptions<DeviceRuntimeListResponse> = {}) => listAIAgentDeviceRuntimesQueryOptions(config, options),
          invalidate: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKey() }),
          invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() }),
          prefetch: (queryClient: QueryClient, options?: RiidoQueryOptions<DeviceRuntimeListResponse>) => queryClient.prefetchQuery(listAIAgentDeviceRuntimesQueryOptions(config, options)),
        },
      },
      events: {
        stream: {
          request: (options?: RiidoRequestOptions) => streamAIAgentClientEvents(config, options),
          queryKeyRoot: streamAIAgentClientEventsQueryKeyRoot,
          queryKey: streamAIAgentClientEventsQueryKey,
          query: (options: RiidoQueryOptions<Response> = {}) => streamAIAgentClientEventsQueryOptions(config, options),
          queryOptions: (options: RiidoQueryOptions<Response> = {}) => streamAIAgentClientEventsQueryOptions(config, options),
          invalidate: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: streamAIAgentClientEventsQueryKey() }),
          invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: streamAIAgentClientEventsQueryKeyRoot() }),
          prefetch: (queryClient: QueryClient, options?: RiidoQueryOptions<Response>) => queryClient.prefetchQuery(streamAIAgentClientEventsQueryOptions(config, options)),
        },
      },
      onboarding: {
        fixtures: {
          request: (options?: RiidoRequestOptions) => listAIAgentOnboardingFixtures(config, options),
          queryKeyRoot: listAIAgentOnboardingFixturesQueryKeyRoot,
          queryKey: listAIAgentOnboardingFixturesQueryKey,
          query: (options: RiidoQueryOptions<AgentOnboardingFixtureListResponse> = {}) => listAIAgentOnboardingFixturesQueryOptions(config, options),
          queryOptions: (options: RiidoQueryOptions<AgentOnboardingFixtureListResponse> = {}) => listAIAgentOnboardingFixturesQueryOptions(config, options),
          invalidate: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentOnboardingFixturesQueryKey() }),
          invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentOnboardingFixturesQueryKeyRoot() }),
          prefetch: (queryClient: QueryClient, options?: RiidoQueryOptions<AgentOnboardingFixtureListResponse>) => queryClient.prefetchQuery(listAIAgentOnboardingFixturesQueryOptions(config, options)),
          createAgent: {
            request: (params: CreateAIAgentFromOnboardingFixturePathParams, body: CreateAgentConfigurationRequest, options?: RiidoRequestOptions) => createAIAgentFromOnboardingFixture(config, params, body, options),
            mutationKey: createAIAgentFromOnboardingFixtureMutationKey,
            mutation: (options: RiidoMutationOptions<AgentClientRecordResponse, CreateAIAgentFromOnboardingFixtureMutationVariables> = {}) => createAIAgentFromOnboardingFixtureMutationOptions(config, options),
            mutationOptions: (options: RiidoMutationOptions<AgentClientRecordResponse, CreateAIAgentFromOnboardingFixtureMutationVariables> = {}) => createAIAgentFromOnboardingFixtureMutationOptions(config, options),
            invalidates: {
              bootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }),
              devicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() }),
              tasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() }),
              all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() })]),
            },
          },
        },
      },
      tasks: {
        assign: {
          request: (params: AssignAIAgentTaskPathParams, body: AssignAIAgentTaskRequest, options?: RiidoRequestOptions) => assignAIAgentTask(config, params, body, options),
          mutationKey: assignAIAgentTaskMutationKey,
          mutation: (options: RiidoMutationOptions<AIAgentTaskActionResponse, AssignAIAgentTaskMutationVariables> = {}) => assignAIAgentTaskMutationOptions(config, options),
          mutationOptions: (options: RiidoMutationOptions<AIAgentTaskActionResponse, AssignAIAgentTaskMutationVariables> = {}) => assignAIAgentTaskMutationOptions(config, options),
          invalidates: {
            bootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }),
            tasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() }),
            tasksThreads: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsQueryKeyRoot() }),
            all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsQueryKeyRoot() })]),
          },
        },
        assignableAgents: {
          request: (params: ListAIAgentTaskAssignableAgentsPathParams, options?: RiidoRequestOptions) => listAIAgentTaskAssignableAgents(config, params, options),
          queryKeyRoot: listAIAgentTaskAssignableAgentsQueryKeyRoot,
          queryKey: listAIAgentTaskAssignableAgentsQueryKey,
          query: (params: ListAIAgentTaskAssignableAgentsPathParams, options: RiidoQueryOptions<AgentClientListResponse> = {}) => listAIAgentTaskAssignableAgentsQueryOptions(config, params, options),
          queryOptions: (params: ListAIAgentTaskAssignableAgentsPathParams, options: RiidoQueryOptions<AgentClientListResponse> = {}) => listAIAgentTaskAssignableAgentsQueryOptions(config, params, options),
          invalidate: (queryClient: QueryClient, params: ListAIAgentTaskAssignableAgentsPathParams) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKey(params) }),
          invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() }),
          prefetch: (queryClient: QueryClient, params: ListAIAgentTaskAssignableAgentsPathParams, options?: RiidoQueryOptions<AgentClientListResponse>) => queryClient.prefetchQuery(listAIAgentTaskAssignableAgentsQueryOptions(config, params, options)),
        },
        stop: {
          request: (params: StopAIAgentTaskPathParams, body: StopAIAgentTaskRequest, options?: RiidoRequestOptions) => stopAIAgentTask(config, params, body, options),
          mutationKey: stopAIAgentTaskMutationKey,
          mutation: (options: RiidoMutationOptions<AIAgentTaskActionResponse, StopAIAgentTaskMutationVariables> = {}) => stopAIAgentTaskMutationOptions(config, options),
          mutationOptions: (options: RiidoMutationOptions<AIAgentTaskActionResponse, StopAIAgentTaskMutationVariables> = {}) => stopAIAgentTaskMutationOptions(config, options),
          invalidates: {
            bootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }),
            tasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() }),
            tasksThreads: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsQueryKeyRoot() }),
            all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsQueryKeyRoot() })]),
          },
        },
        submitComment: {
          request: (params: SubmitAIAgentTaskCommentPathParams, body: SubmitAIAgentTaskCommentRequest, options?: RiidoRequestOptions) => submitAIAgentTaskComment(config, params, body, options),
          mutationKey: submitAIAgentTaskCommentMutationKey,
          mutation: (options: RiidoMutationOptions<AIAgentTaskActionResponse, SubmitAIAgentTaskCommentMutationVariables> = {}) => submitAIAgentTaskCommentMutationOptions(config, options),
          mutationOptions: (options: RiidoMutationOptions<AIAgentTaskActionResponse, SubmitAIAgentTaskCommentMutationVariables> = {}) => submitAIAgentTaskCommentMutationOptions(config, options),
          invalidates: {
            bootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }),
            tasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() }),
            tasksThreads: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsQueryKeyRoot() }),
            all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsQueryKeyRoot() })]),
          },
        },
        threadMessages: {
          create: {
            request: (params: CreateAIAgentTaskThreadMessagePathParams, body: CreateAIAgentTaskThreadMessageRequest, options?: RiidoRequestOptions) => createAIAgentTaskThreadMessage(config, params, body, options),
            mutationKey: createAIAgentTaskThreadMessageMutationKey,
            mutation: (options: RiidoMutationOptions<AIAgentTaskActionResponse, CreateAIAgentTaskThreadMessageMutationVariables> = {}) => createAIAgentTaskThreadMessageMutationOptions(config, options),
            mutationOptions: (options: RiidoMutationOptions<AIAgentTaskActionResponse, CreateAIAgentTaskThreadMessageMutationVariables> = {}) => createAIAgentTaskThreadMessageMutationOptions(config, options),
            invalidates: {
              bootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }),
              tasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() }),
              tasksThreads: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsQueryKeyRoot() }),
              all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsQueryKeyRoot() })]),
            },
          },
        },
        threads: {
          request: (params: ListAIAgentTaskThreadsPathParams, options?: RiidoRequestOptions) => listAIAgentTaskThreads(config, params, options),
          queryKeyRoot: listAIAgentTaskThreadsQueryKeyRoot,
          queryKey: listAIAgentTaskThreadsQueryKey,
          query: (params: ListAIAgentTaskThreadsPathParams, options: RiidoQueryOptions<AIAgentTaskThreadCollectionResponse> = {}) => listAIAgentTaskThreadsQueryOptions(config, params, options),
          queryOptions: (params: ListAIAgentTaskThreadsPathParams, options: RiidoQueryOptions<AIAgentTaskThreadCollectionResponse> = {}) => listAIAgentTaskThreadsQueryOptions(config, params, options),
          invalidate: (queryClient: QueryClient, params: ListAIAgentTaskThreadsPathParams) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsQueryKey(params) }),
          invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsQueryKeyRoot() }),
          prefetch: (queryClient: QueryClient, params: ListAIAgentTaskThreadsPathParams, options?: RiidoQueryOptions<AIAgentTaskThreadCollectionResponse>) => queryClient.prefetchQuery(listAIAgentTaskThreadsQueryOptions(config, params, options)),
        },
        unassign: {
          request: (params: UnassignAIAgentTaskPathParams, body: UnassignAIAgentTaskRequest, options?: RiidoRequestOptions) => unassignAIAgentTask(config, params, body, options),
          mutationKey: unassignAIAgentTaskMutationKey,
          mutation: (options: RiidoMutationOptions<AIAgentTaskActionResponse, UnassignAIAgentTaskMutationVariables> = {}) => unassignAIAgentTaskMutationOptions(config, options),
          mutationOptions: (options: RiidoMutationOptions<AIAgentTaskActionResponse, UnassignAIAgentTaskMutationVariables> = {}) => unassignAIAgentTaskMutationOptions(config, options),
          invalidates: {
            bootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }),
            tasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() }),
            tasksThreads: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsQueryKeyRoot() }),
            all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsQueryKeyRoot() })]),
          },
        },
      },
    },
    v2: {
      aiAgent: {
        agents: {
          create: {
            request: (params: CreateAIAgentV2PathParams, body: CreateAgentConfigurationRequest, options?: RiidoRequestOptions) => createAIAgentV2(config, params, body, options),
            mutationKey: createAIAgentV2MutationKey,
            mutation: (options: RiidoMutationOptions<AgentClientRecordResponseV2, CreateAIAgentV2MutationVariables> = {}) => createAIAgentV2MutationOptions(config, options),
            mutationOptions: (options: RiidoMutationOptions<AgentClientRecordResponseV2, CreateAIAgentV2MutationVariables> = {}) => createAIAgentV2MutationOptions(config, options),
            invalidates: {
              aiAgentBootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }),
              aiAgentDevicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() }),
              aiAgentTasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }),
              all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() })]),
            },
          },
          daemon: {
            details: {
              request: (params: GetAIAgentDaemonV2PathParams, options?: RiidoRequestOptions) => getAIAgentDaemonV2(config, params, options),
              queryKeyRoot: getAIAgentDaemonV2QueryKeyRoot,
              queryKey: getAIAgentDaemonV2QueryKey,
              query: (params: GetAIAgentDaemonV2PathParams, options: RiidoQueryOptions<DeviceDaemonDetailResponse> = {}) => getAIAgentDaemonV2QueryOptions(config, params, options),
              queryOptions: (params: GetAIAgentDaemonV2PathParams, options: RiidoQueryOptions<DeviceDaemonDetailResponse> = {}) => getAIAgentDaemonV2QueryOptions(config, params, options),
              invalidate: (queryClient: QueryClient, params: GetAIAgentDaemonV2PathParams) => queryClient.invalidateQueries({ queryKey: getAIAgentDaemonV2QueryKey(params) }),
              invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentDaemonV2QueryKeyRoot() }),
              prefetch: (queryClient: QueryClient, params: GetAIAgentDaemonV2PathParams, options?: RiidoQueryOptions<DeviceDaemonDetailResponse>) => queryClient.prefetchQuery(getAIAgentDaemonV2QueryOptions(config, params, options)),
            },
            restart: {
              request: (params: RestartAIAgentDaemonV2PathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => restartAIAgentDaemonV2(config, params, body, options),
              mutationKey: restartAIAgentDaemonV2MutationKey,
              mutation: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDaemonV2MutationVariables> = {}) => restartAIAgentDaemonV2MutationOptions(config, options),
              mutationOptions: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDaemonV2MutationVariables> = {}) => restartAIAgentDaemonV2MutationOptions(config, options),
              invalidates: {
                aiAgentAgentsDaemon: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentDaemonV2QueryKeyRoot() }),
                aiAgentDevicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() }),
                all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentDaemonV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() })]),
              },
            },
            start: {
              request: (params: StartAIAgentDaemonV2PathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => startAIAgentDaemonV2(config, params, body, options),
              mutationKey: startAIAgentDaemonV2MutationKey,
              mutation: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDaemonV2MutationVariables> = {}) => startAIAgentDaemonV2MutationOptions(config, options),
              mutationOptions: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDaemonV2MutationVariables> = {}) => startAIAgentDaemonV2MutationOptions(config, options),
              invalidates: {
                aiAgentAgentsDaemon: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentDaemonV2QueryKeyRoot() }),
                aiAgentDevicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() }),
                all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentDaemonV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() })]),
              },
            },
            stop: {
              request: (params: StopAIAgentDaemonV2PathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => stopAIAgentDaemonV2(config, params, body, options),
              mutationKey: stopAIAgentDaemonV2MutationKey,
              mutation: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDaemonV2MutationVariables> = {}) => stopAIAgentDaemonV2MutationOptions(config, options),
              mutationOptions: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDaemonV2MutationVariables> = {}) => stopAIAgentDaemonV2MutationOptions(config, options),
              invalidates: {
                aiAgentAgentsDaemon: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentDaemonV2QueryKeyRoot() }),
                aiAgentDevicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() }),
                all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentDaemonV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() })]),
              },
            },
          },
          delete: {
            request: (params: DeleteAIAgentV2PathParams, options?: RiidoRequestOptions) => deleteAIAgentV2(config, params, options),
            mutationKey: deleteAIAgentV2MutationKey,
            mutation: (options: RiidoMutationOptions<DeleteAgentResponse, DeleteAIAgentV2MutationVariables> = {}) => deleteAIAgentV2MutationOptions(config, options),
            mutationOptions: (options: RiidoMutationOptions<DeleteAgentResponse, DeleteAIAgentV2MutationVariables> = {}) => deleteAIAgentV2MutationOptions(config, options),
            invalidates: {
              aiAgentBootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }),
              aiAgentDevicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() }),
              aiAgentAgentsEditability: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentEditabilityV2QueryKeyRoot() }),
              aiAgentTasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }),
              aiAgentTasksThreads: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() }),
              aiAgentTasksAssignedAgentProfiles: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listWorkspaceAssignedAgentProfilesV2QueryKeyRoot() }),
              all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: getAIAgentEditabilityV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listWorkspaceAssignedAgentProfilesV2QueryKeyRoot() })]),
            },
          },
          editability: {
            request: (params: GetAIAgentEditabilityV2PathParams, options?: RiidoRequestOptions) => getAIAgentEditabilityV2(config, params, options),
            queryKeyRoot: getAIAgentEditabilityV2QueryKeyRoot,
            queryKey: getAIAgentEditabilityV2QueryKey,
            query: (params: GetAIAgentEditabilityV2PathParams, options: RiidoQueryOptions<AgentEditabilityResponse> = {}) => getAIAgentEditabilityV2QueryOptions(config, params, options),
            queryOptions: (params: GetAIAgentEditabilityV2PathParams, options: RiidoQueryOptions<AgentEditabilityResponse> = {}) => getAIAgentEditabilityV2QueryOptions(config, params, options),
            invalidate: (queryClient: QueryClient, params: GetAIAgentEditabilityV2PathParams) => queryClient.invalidateQueries({ queryKey: getAIAgentEditabilityV2QueryKey(params) }),
            invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentEditabilityV2QueryKeyRoot() }),
            prefetch: (queryClient: QueryClient, params: GetAIAgentEditabilityV2PathParams, options?: RiidoQueryOptions<AgentEditabilityResponse>) => queryClient.prefetchQuery(getAIAgentEditabilityV2QueryOptions(config, params, options)),
          },
          updateConfiguration: {
            request: (params: UpdateAIAgentConfigurationV2PathParams, body: UpdateAgentConfigurationRequest, options?: RiidoRequestOptions) => updateAIAgentConfigurationV2(config, params, body, options),
            mutationKey: updateAIAgentConfigurationV2MutationKey,
            mutation: (options: RiidoMutationOptions<AgentClientRecordResponseV2, UpdateAIAgentConfigurationV2MutationVariables> = {}) => updateAIAgentConfigurationV2MutationOptions(config, options),
            mutationOptions: (options: RiidoMutationOptions<AgentClientRecordResponseV2, UpdateAIAgentConfigurationV2MutationVariables> = {}) => updateAIAgentConfigurationV2MutationOptions(config, options),
            invalidates: {
              aiAgentBootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }),
              aiAgentAgentsEditability: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentEditabilityV2QueryKeyRoot() }),
              aiAgentTasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }),
              all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: getAIAgentEditabilityV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() })]),
            },
          },
        },
        bootstrap: {
          request: (params: GetAIAgentClientBootstrapV2PathParams, options?: RiidoRequestOptions) => getAIAgentClientBootstrapV2(config, params, options),
          queryKeyRoot: getAIAgentClientBootstrapV2QueryKeyRoot,
          queryKey: getAIAgentClientBootstrapV2QueryKey,
          query: (params: GetAIAgentClientBootstrapV2PathParams, options: RiidoQueryOptions<ClientBootstrapResponseV2> = {}) => getAIAgentClientBootstrapV2QueryOptions(config, params, options),
          queryOptions: (params: GetAIAgentClientBootstrapV2PathParams, options: RiidoQueryOptions<ClientBootstrapResponseV2> = {}) => getAIAgentClientBootstrapV2QueryOptions(config, params, options),
          invalidate: (queryClient: QueryClient, params: GetAIAgentClientBootstrapV2PathParams) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKey(params) }),
          invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }),
          prefetch: (queryClient: QueryClient, params: GetAIAgentClientBootstrapV2PathParams, options?: RiidoQueryOptions<ClientBootstrapResponseV2>) => queryClient.prefetchQuery(getAIAgentClientBootstrapV2QueryOptions(config, params, options)),
        },
        devices: {
          daemon: {
            details: {
              request: (params: GetAIAgentDeviceDaemonV2PathParams, options?: RiidoRequestOptions) => getAIAgentDeviceDaemonV2(config, params, options),
              queryKeyRoot: getAIAgentDeviceDaemonV2QueryKeyRoot,
              queryKey: getAIAgentDeviceDaemonV2QueryKey,
              query: (params: GetAIAgentDeviceDaemonV2PathParams, options: RiidoQueryOptions<DeviceDaemonDetailResponse> = {}) => getAIAgentDeviceDaemonV2QueryOptions(config, params, options),
              queryOptions: (params: GetAIAgentDeviceDaemonV2PathParams, options: RiidoQueryOptions<DeviceDaemonDetailResponse> = {}) => getAIAgentDeviceDaemonV2QueryOptions(config, params, options),
              invalidate: (queryClient: QueryClient, params: GetAIAgentDeviceDaemonV2PathParams) => queryClient.invalidateQueries({ queryKey: getAIAgentDeviceDaemonV2QueryKey(params) }),
              invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentDeviceDaemonV2QueryKeyRoot() }),
              prefetch: (queryClient: QueryClient, params: GetAIAgentDeviceDaemonV2PathParams, options?: RiidoQueryOptions<DeviceDaemonDetailResponse>) => queryClient.prefetchQuery(getAIAgentDeviceDaemonV2QueryOptions(config, params, options)),
            },
            restart: {
              request: (params: RestartAIAgentDeviceDaemonV2PathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => restartAIAgentDeviceDaemonV2(config, params, body, options),
              mutationKey: restartAIAgentDeviceDaemonV2MutationKey,
              mutation: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDeviceDaemonV2MutationVariables> = {}) => restartAIAgentDeviceDaemonV2MutationOptions(config, options),
              mutationOptions: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, RestartAIAgentDeviceDaemonV2MutationVariables> = {}) => restartAIAgentDeviceDaemonV2MutationOptions(config, options),
              invalidates: {
                aiAgentDevicesDaemon: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentDeviceDaemonV2QueryKeyRoot() }),
                aiAgentDevicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() }),
                all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentDeviceDaemonV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() })]),
              },
            },
            start: {
              request: (params: StartAIAgentDeviceDaemonV2PathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => startAIAgentDeviceDaemonV2(config, params, body, options),
              mutationKey: startAIAgentDeviceDaemonV2MutationKey,
              mutation: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDeviceDaemonV2MutationVariables> = {}) => startAIAgentDeviceDaemonV2MutationOptions(config, options),
              mutationOptions: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, StartAIAgentDeviceDaemonV2MutationVariables> = {}) => startAIAgentDeviceDaemonV2MutationOptions(config, options),
              invalidates: {
                aiAgentDevicesDaemon: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentDeviceDaemonV2QueryKeyRoot() }),
                aiAgentDevicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() }),
                all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentDeviceDaemonV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() })]),
              },
            },
            stop: {
              request: (params: StopAIAgentDeviceDaemonV2PathParams, body: ControlDeviceDaemonRequest, options?: RiidoRequestOptions) => stopAIAgentDeviceDaemonV2(config, params, body, options),
              mutationKey: stopAIAgentDeviceDaemonV2MutationKey,
              mutation: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDeviceDaemonV2MutationVariables> = {}) => stopAIAgentDeviceDaemonV2MutationOptions(config, options),
              mutationOptions: (options: RiidoMutationOptions<DeviceDaemonCommandResponse, StopAIAgentDeviceDaemonV2MutationVariables> = {}) => stopAIAgentDeviceDaemonV2MutationOptions(config, options),
              invalidates: {
                aiAgentDevicesDaemon: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentDeviceDaemonV2QueryKeyRoot() }),
                aiAgentDevicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() }),
                all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentDeviceDaemonV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() })]),
              },
            },
          },
          runtimes: {
            request: (params: ListAIAgentDeviceRuntimesV2PathParams, options?: RiidoRequestOptions) => listAIAgentDeviceRuntimesV2(config, params, options),
            queryKeyRoot: listAIAgentDeviceRuntimesV2QueryKeyRoot,
            queryKey: listAIAgentDeviceRuntimesV2QueryKey,
            query: (params: ListAIAgentDeviceRuntimesV2PathParams, options: RiidoQueryOptions<DeviceRuntimeListResponse> = {}) => listAIAgentDeviceRuntimesV2QueryOptions(config, params, options),
            queryOptions: (params: ListAIAgentDeviceRuntimesV2PathParams, options: RiidoQueryOptions<DeviceRuntimeListResponse> = {}) => listAIAgentDeviceRuntimesV2QueryOptions(config, params, options),
            invalidate: (queryClient: QueryClient, params: ListAIAgentDeviceRuntimesV2PathParams) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKey(params) }),
            invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() }),
            prefetch: (queryClient: QueryClient, params: ListAIAgentDeviceRuntimesV2PathParams, options?: RiidoQueryOptions<DeviceRuntimeListResponse>) => queryClient.prefetchQuery(listAIAgentDeviceRuntimesV2QueryOptions(config, params, options)),
          },
        },
        events: {
          stream: {
            request: (params: StreamAIAgentClientEventsV2PathParams, options?: RiidoRequestOptions) => streamAIAgentClientEventsV2(config, params, options),
            queryKeyRoot: streamAIAgentClientEventsV2QueryKeyRoot,
            queryKey: streamAIAgentClientEventsV2QueryKey,
            query: (params: StreamAIAgentClientEventsV2PathParams, options: RiidoQueryOptions<Response> = {}) => streamAIAgentClientEventsV2QueryOptions(config, params, options),
            queryOptions: (params: StreamAIAgentClientEventsV2PathParams, options: RiidoQueryOptions<Response> = {}) => streamAIAgentClientEventsV2QueryOptions(config, params, options),
            invalidate: (queryClient: QueryClient, params: StreamAIAgentClientEventsV2PathParams) => queryClient.invalidateQueries({ queryKey: streamAIAgentClientEventsV2QueryKey(params) }),
            invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: streamAIAgentClientEventsV2QueryKeyRoot() }),
            prefetch: (queryClient: QueryClient, params: StreamAIAgentClientEventsV2PathParams, options?: RiidoQueryOptions<Response>) => queryClient.prefetchQuery(streamAIAgentClientEventsV2QueryOptions(config, params, options)),
          },
        },
        onboarding: {
          fixtures: {
            request: (params: ListAIAgentOnboardingFixturesV2PathParams, options?: RiidoRequestOptions) => listAIAgentOnboardingFixturesV2(config, params, options),
            queryKeyRoot: listAIAgentOnboardingFixturesV2QueryKeyRoot,
            queryKey: listAIAgentOnboardingFixturesV2QueryKey,
            query: (params: ListAIAgentOnboardingFixturesV2PathParams, options: RiidoQueryOptions<AgentOnboardingFixtureListResponse> = {}) => listAIAgentOnboardingFixturesV2QueryOptions(config, params, options),
            queryOptions: (params: ListAIAgentOnboardingFixturesV2PathParams, options: RiidoQueryOptions<AgentOnboardingFixtureListResponse> = {}) => listAIAgentOnboardingFixturesV2QueryOptions(config, params, options),
            invalidate: (queryClient: QueryClient, params: ListAIAgentOnboardingFixturesV2PathParams) => queryClient.invalidateQueries({ queryKey: listAIAgentOnboardingFixturesV2QueryKey(params) }),
            invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentOnboardingFixturesV2QueryKeyRoot() }),
            prefetch: (queryClient: QueryClient, params: ListAIAgentOnboardingFixturesV2PathParams, options?: RiidoQueryOptions<AgentOnboardingFixtureListResponse>) => queryClient.prefetchQuery(listAIAgentOnboardingFixturesV2QueryOptions(config, params, options)),
            createAgent: {
              request: (params: CreateAIAgentFromOnboardingFixtureV2PathParams, body: CreateAgentConfigurationRequest, options?: RiidoRequestOptions) => createAIAgentFromOnboardingFixtureV2(config, params, body, options),
              mutationKey: createAIAgentFromOnboardingFixtureV2MutationKey,
              mutation: (options: RiidoMutationOptions<AgentClientRecordResponseV2, CreateAIAgentFromOnboardingFixtureV2MutationVariables> = {}) => createAIAgentFromOnboardingFixtureV2MutationOptions(config, options),
              mutationOptions: (options: RiidoMutationOptions<AgentClientRecordResponseV2, CreateAIAgentFromOnboardingFixtureV2MutationVariables> = {}) => createAIAgentFromOnboardingFixtureV2MutationOptions(config, options),
              invalidates: {
                aiAgentBootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }),
                aiAgentDevicesRuntimes: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() }),
                aiAgentTasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }),
                all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() })]),
              },
            },
          },
        },
        tasks: {
          agentAssignments: {
            create: {
              request: (params: CreateAIAgentTaskAgentAssignmentV2PathParams, body: AssignAIAgentTaskRequest, options?: RiidoRequestOptions) => createAIAgentTaskAgentAssignmentV2(config, params, body, options),
              mutationKey: createAIAgentTaskAgentAssignmentV2MutationKey,
              mutation: (options: RiidoMutationOptions<AIAgentTaskActionResponse, CreateAIAgentTaskAgentAssignmentV2MutationVariables> = {}) => createAIAgentTaskAgentAssignmentV2MutationOptions(config, options),
              mutationOptions: (options: RiidoMutationOptions<AIAgentTaskActionResponse, CreateAIAgentTaskAgentAssignmentV2MutationVariables> = {}) => createAIAgentTaskAgentAssignmentV2MutationOptions(config, options),
              invalidates: {
                aiAgentBootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }),
                aiAgentTasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }),
                aiAgentTasksThreads: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() }),
                aiAgentTasksThreadStreamSubscription: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentTaskThreadStreamSubscriptionV2QueryKeyRoot() }),
                all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: getAIAgentTaskThreadStreamSubscriptionV2QueryKeyRoot() })]),
              },
            },
            delete: {
              request: (params: DeleteAIAgentTaskAgentAssignmentV2PathParams, options?: RiidoRequestOptions) => deleteAIAgentTaskAgentAssignmentV2(config, params, options),
              mutationKey: deleteAIAgentTaskAgentAssignmentV2MutationKey,
              mutation: (options: RiidoMutationOptions<AIAgentTaskActionResponse, DeleteAIAgentTaskAgentAssignmentV2MutationVariables> = {}) => deleteAIAgentTaskAgentAssignmentV2MutationOptions(config, options),
              mutationOptions: (options: RiidoMutationOptions<AIAgentTaskActionResponse, DeleteAIAgentTaskAgentAssignmentV2MutationVariables> = {}) => deleteAIAgentTaskAgentAssignmentV2MutationOptions(config, options),
              invalidates: {
                aiAgentBootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }),
                aiAgentTasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }),
                aiAgentTasksThreads: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() }),
                aiAgentTasksThreadStreamSubscription: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentTaskThreadStreamSubscriptionV2QueryKeyRoot() }),
                all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: getAIAgentTaskThreadStreamSubscriptionV2QueryKeyRoot() })]),
              },
            },
            stop: {
              request: (params: StopAIAgentTaskAgentAssignmentV2PathParams, options?: RiidoRequestOptions) => stopAIAgentTaskAgentAssignmentV2(config, params, options),
              mutationKey: stopAIAgentTaskAgentAssignmentV2MutationKey,
              mutation: (options: RiidoMutationOptions<AIAgentTaskActionResponse, StopAIAgentTaskAgentAssignmentV2MutationVariables> = {}) => stopAIAgentTaskAgentAssignmentV2MutationOptions(config, options),
              mutationOptions: (options: RiidoMutationOptions<AIAgentTaskActionResponse, StopAIAgentTaskAgentAssignmentV2MutationVariables> = {}) => stopAIAgentTaskAgentAssignmentV2MutationOptions(config, options),
              invalidates: {
                aiAgentBootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }),
                aiAgentTasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }),
                aiAgentTasksThreads: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() }),
                aiAgentTasksThreadStreamSubscription: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentTaskThreadStreamSubscriptionV2QueryKeyRoot() }),
                all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: getAIAgentTaskThreadStreamSubscriptionV2QueryKeyRoot() })]),
              },
            },
          },
          assign: {
            request: (params: AssignAIAgentTaskV2PathParams, body: AssignAIAgentTaskRequest, options?: RiidoRequestOptions) => assignAIAgentTaskV2(config, params, body, options),
            mutationKey: assignAIAgentTaskV2MutationKey,
            mutation: (options: RiidoMutationOptions<AIAgentTaskActionResponse, AssignAIAgentTaskV2MutationVariables> = {}) => assignAIAgentTaskV2MutationOptions(config, options),
            mutationOptions: (options: RiidoMutationOptions<AIAgentTaskActionResponse, AssignAIAgentTaskV2MutationVariables> = {}) => assignAIAgentTaskV2MutationOptions(config, options),
            invalidates: {
              aiAgentBootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }),
              aiAgentTasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }),
              aiAgentTasksThreads: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() }),
              aiAgentTasksAssignedAgentProfiles: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listWorkspaceAssignedAgentProfilesV2QueryKeyRoot() }),
              all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listWorkspaceAssignedAgentProfilesV2QueryKeyRoot() })]),
            },
          },
          assignableAgents: {
            request: (params: ListAIAgentTaskAssignableAgentsV2PathParams, options?: RiidoRequestOptions) => listAIAgentTaskAssignableAgentsV2(config, params, options),
            queryKeyRoot: listAIAgentTaskAssignableAgentsV2QueryKeyRoot,
            queryKey: listAIAgentTaskAssignableAgentsV2QueryKey,
            query: (params: ListAIAgentTaskAssignableAgentsV2PathParams, options: RiidoQueryOptions<AgentClientListResponseV2> = {}) => listAIAgentTaskAssignableAgentsV2QueryOptions(config, params, options),
            queryOptions: (params: ListAIAgentTaskAssignableAgentsV2PathParams, options: RiidoQueryOptions<AgentClientListResponseV2> = {}) => listAIAgentTaskAssignableAgentsV2QueryOptions(config, params, options),
            invalidate: (queryClient: QueryClient, params: ListAIAgentTaskAssignableAgentsV2PathParams) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKey(params) }),
            invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }),
            prefetch: (queryClient: QueryClient, params: ListAIAgentTaskAssignableAgentsV2PathParams, options?: RiidoQueryOptions<AgentClientListResponseV2>) => queryClient.prefetchQuery(listAIAgentTaskAssignableAgentsV2QueryOptions(config, params, options)),
          },
          assignedAgentProfiles: {
            request: (params: ListWorkspaceAssignedAgentProfilesV2PathParams, options?: RiidoRequestOptions) => listWorkspaceAssignedAgentProfilesV2(config, params, options),
            queryKeyRoot: listWorkspaceAssignedAgentProfilesV2QueryKeyRoot,
            queryKey: listWorkspaceAssignedAgentProfilesV2QueryKey,
            query: (params: ListWorkspaceAssignedAgentProfilesV2PathParams, options: RiidoQueryOptions<AssignedAgentProfileMapResponse> = {}) => listWorkspaceAssignedAgentProfilesV2QueryOptions(config, params, options),
            queryOptions: (params: ListWorkspaceAssignedAgentProfilesV2PathParams, options: RiidoQueryOptions<AssignedAgentProfileMapResponse> = {}) => listWorkspaceAssignedAgentProfilesV2QueryOptions(config, params, options),
            invalidate: (queryClient: QueryClient, params: ListWorkspaceAssignedAgentProfilesV2PathParams) => queryClient.invalidateQueries({ queryKey: listWorkspaceAssignedAgentProfilesV2QueryKey(params) }),
            invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listWorkspaceAssignedAgentProfilesV2QueryKeyRoot() }),
            prefetch: (queryClient: QueryClient, params: ListWorkspaceAssignedAgentProfilesV2PathParams, options?: RiidoQueryOptions<AssignedAgentProfileMapResponse>) => queryClient.prefetchQuery(listWorkspaceAssignedAgentProfilesV2QueryOptions(config, params, options)),
          },
          stop: {
            request: (params: StopAIAgentTaskV2PathParams, body: StopAIAgentTaskRequest, options?: RiidoRequestOptions) => stopAIAgentTaskV2(config, params, body, options),
            mutationKey: stopAIAgentTaskV2MutationKey,
            mutation: (options: RiidoMutationOptions<AIAgentTaskActionResponse, StopAIAgentTaskV2MutationVariables> = {}) => stopAIAgentTaskV2MutationOptions(config, options),
            mutationOptions: (options: RiidoMutationOptions<AIAgentTaskActionResponse, StopAIAgentTaskV2MutationVariables> = {}) => stopAIAgentTaskV2MutationOptions(config, options),
            invalidates: {
              aiAgentBootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }),
              aiAgentTasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }),
              aiAgentTasksThreads: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() }),
              all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() })]),
            },
          },
          submitComment: {
            request: (params: SubmitAIAgentTaskCommentV2PathParams, body: SubmitAIAgentTaskCommentRequest, options?: RiidoRequestOptions) => submitAIAgentTaskCommentV2(config, params, body, options),
            mutationKey: submitAIAgentTaskCommentV2MutationKey,
            mutation: (options: RiidoMutationOptions<AIAgentTaskActionResponse, SubmitAIAgentTaskCommentV2MutationVariables> = {}) => submitAIAgentTaskCommentV2MutationOptions(config, options),
            mutationOptions: (options: RiidoMutationOptions<AIAgentTaskActionResponse, SubmitAIAgentTaskCommentV2MutationVariables> = {}) => submitAIAgentTaskCommentV2MutationOptions(config, options),
            invalidates: {
              aiAgentBootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }),
              aiAgentTasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }),
              aiAgentTasksThreads: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() }),
              aiAgentTasksAssignedAgentProfiles: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listWorkspaceAssignedAgentProfilesV2QueryKeyRoot() }),
              all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listWorkspaceAssignedAgentProfilesV2QueryKeyRoot() })]),
            },
          },
          threadMessages: {
            create: {
              request: (params: CreateAIAgentTaskThreadMessageV2PathParams, body: CreateAIAgentTaskThreadMessageRequest, options?: RiidoRequestOptions) => createAIAgentTaskThreadMessageV2(config, params, body, options),
              mutationKey: createAIAgentTaskThreadMessageV2MutationKey,
              mutation: (options: RiidoMutationOptions<AIAgentTaskActionResponse, CreateAIAgentTaskThreadMessageV2MutationVariables> = {}) => createAIAgentTaskThreadMessageV2MutationOptions(config, options),
              mutationOptions: (options: RiidoMutationOptions<AIAgentTaskActionResponse, CreateAIAgentTaskThreadMessageV2MutationVariables> = {}) => createAIAgentTaskThreadMessageV2MutationOptions(config, options),
              invalidates: {
                aiAgentBootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }),
                aiAgentTasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }),
                aiAgentTasksThreads: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() }),
                all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() })]),
              },
            },
          },
          threadStreamSubscription: {
            request: (params: GetAIAgentTaskThreadStreamSubscriptionV2PathParams, options?: RiidoRequestOptions) => getAIAgentTaskThreadStreamSubscriptionV2(config, params, options),
            queryKeyRoot: getAIAgentTaskThreadStreamSubscriptionV2QueryKeyRoot,
            queryKey: getAIAgentTaskThreadStreamSubscriptionV2QueryKey,
            query: (params: GetAIAgentTaskThreadStreamSubscriptionV2PathParams, options: RiidoQueryOptions<AIAgentTaskThreadStreamSubscriptionResponse> = {}) => getAIAgentTaskThreadStreamSubscriptionV2QueryOptions(config, params, options),
            queryOptions: (params: GetAIAgentTaskThreadStreamSubscriptionV2PathParams, options: RiidoQueryOptions<AIAgentTaskThreadStreamSubscriptionResponse> = {}) => getAIAgentTaskThreadStreamSubscriptionV2QueryOptions(config, params, options),
            invalidate: (queryClient: QueryClient, params: GetAIAgentTaskThreadStreamSubscriptionV2PathParams) => queryClient.invalidateQueries({ queryKey: getAIAgentTaskThreadStreamSubscriptionV2QueryKey(params) }),
            invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentTaskThreadStreamSubscriptionV2QueryKeyRoot() }),
            prefetch: (queryClient: QueryClient, params: GetAIAgentTaskThreadStreamSubscriptionV2PathParams, options?: RiidoQueryOptions<AIAgentTaskThreadStreamSubscriptionResponse>) => queryClient.prefetchQuery(getAIAgentTaskThreadStreamSubscriptionV2QueryOptions(config, params, options)),
          },
          threads: {
            request: (params: ListAIAgentTaskThreadsV2PathParams, options?: RiidoRequestOptions) => listAIAgentTaskThreadsV2(config, params, options),
            queryKeyRoot: listAIAgentTaskThreadsV2QueryKeyRoot,
            queryKey: listAIAgentTaskThreadsV2QueryKey,
            query: (params: ListAIAgentTaskThreadsV2PathParams, options: RiidoQueryOptions<AIAgentTaskThreadCollectionResponse> = {}) => listAIAgentTaskThreadsV2QueryOptions(config, params, options),
            queryOptions: (params: ListAIAgentTaskThreadsV2PathParams, options: RiidoQueryOptions<AIAgentTaskThreadCollectionResponse> = {}) => listAIAgentTaskThreadsV2QueryOptions(config, params, options),
            invalidate: (queryClient: QueryClient, params: ListAIAgentTaskThreadsV2PathParams) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKey(params) }),
            invalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() }),
            prefetch: (queryClient: QueryClient, params: ListAIAgentTaskThreadsV2PathParams, options?: RiidoQueryOptions<AIAgentTaskThreadCollectionResponse>) => queryClient.prefetchQuery(listAIAgentTaskThreadsV2QueryOptions(config, params, options)),
          },
          unassign: {
            request: (params: UnassignAIAgentTaskV2PathParams, body: UnassignAIAgentTaskRequest, options?: RiidoRequestOptions) => unassignAIAgentTaskV2(config, params, body, options),
            mutationKey: unassignAIAgentTaskV2MutationKey,
            mutation: (options: RiidoMutationOptions<AIAgentTaskActionResponse, UnassignAIAgentTaskV2MutationVariables> = {}) => unassignAIAgentTaskV2MutationOptions(config, options),
            mutationOptions: (options: RiidoMutationOptions<AIAgentTaskActionResponse, UnassignAIAgentTaskV2MutationVariables> = {}) => unassignAIAgentTaskV2MutationOptions(config, options),
            invalidates: {
              aiAgentBootstrap: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }),
              aiAgentTasksAssignableAgents: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }),
              aiAgentTasksThreads: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() }),
              aiAgentTasksAssignedAgentProfiles: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: listWorkspaceAssignedAgentProfilesV2QueryKeyRoot() }),
              all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskThreadsV2QueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listWorkspaceAssignedAgentProfilesV2QueryKeyRoot() })]),
            },
          },
        },
      },
    },
  };
}
