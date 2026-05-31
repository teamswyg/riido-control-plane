// 이 파일은 tools/reactquerygen으로 생성됩니다. 직접 수정하지 마세요.
// 원본: contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json

import type { QueryClient, UseMutationOptions, UseQueryOptions } from '@/lib/react-query';

/**
 * task thread action 요청 이후 client가 즉시 반영할 agent 작업 상태 응답입니다.
 */
export interface AIAgentTaskActionResponse {
  agent_id: string;
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
  assignment_state: AgentAssignmentState;
  comment_kind: AgentTaskCommentKind;
  completed_at?: string;
  lines: AgentThreadProgressLine[];
  message: string;
  run_id: string;
  source_comment_id?: string;
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
 * agent 삭제와 runtime lifecycle에 의해 변경되는 task assignment 상태입니다.
 */
export type AgentAssignmentState = "queued" | "running" | "stopping" | "stopped" | "completed" | "failed" | "unassigned";

/**
 * task 또는 화면에서 표시할 agent 목록 응답입니다.
 */
export interface AgentClientListResponse {
  agents: AgentClientRecord[];
  schema_version: string;
}

/**
 * client 화면에 표시하는 agent 요약 record입니다.
 */
export interface AgentClientRecord {
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
 * AI Agent 온보딩에서 선택할 수 있는 starter agent template입니다.
 */
export interface AgentOnboardingTemplate {
  description: string;
  instruction: string;
  name: string;
  profile_thumbnail_url?: string;
  role_label?: string;
  template_id: string;
}

/**
 * Figma comment 소통 흐름에서 task thread에 기록되는 상태 update 종류입니다.
 */
export type AgentTaskCommentKind = "queued_by_busy_agent" | "assignment_started" | "stopped_by_agent_deleted" | "stopped_by_user_request" | "runtime_progress" | "task_completed" | "task_failed";

/**
 * 활성 task thread에 추가된 AI Agent 진행 상태를 client SSE로 전달하는 event입니다.
 */
export interface AgentThreadProgressEvent {
  agent_id: string;
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
 * agent 작업 상태와 task thread comment 상태 변경을 전달하는 SSE event입니다.
 */
export interface AgentWorkStatusChangedEvent {
  agent_id: string;
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
 * task participant dropdown에서 agent를 배정하기 위한 요청입니다.
 */
export interface AssignAIAgentTaskRequest {
  agent_id: string;
}

/**
 * AI Agent 화면 진입 시 필요한 agent와 device runtime 초기 데이터입니다.
 */
export interface ClientBootstrapResponse {
  agent_templates: AgentOnboardingTemplate[];
  agents: AgentClientRecord[];
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
 * current-device daemon start/restart/stop command 요청입니다. reason은 audit 표시용이며 화면 표시 정책은 client가 결정합니다.
 */
export interface ControlDeviceDaemonRequest {
  reason?: string;
}

/**
 * agent 이름, 공개 범위, runtime, model, profile field를 저장하기 위한 생성 요청입니다.
 */
export interface CreateAgentConfigurationRequest {
  description?: string;
  instruction?: string;
  model_id?: string;
  name: string;
  profile_thumbnail_url?: string;
  runtime_id: string;
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
 * runtime 설정 화면의 daemon 상세 row와 상세 패널을 위한 응답입니다.
 */
export interface DeviceDaemonDetailResponse {
  daemon: DeviceDaemonRecord;
  schema_version: string;
}

/**
 * runtime 설정 화면에서 현재 device의 daemon 상세 표시와 제어 버튼 상태를 구성하는 read model입니다.
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
 * current-device daemon 상세/제어 상태가 변경되었음을 client SSE로 전달하는 event입니다.
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
 * task thread comment를 agent에게 전달하기 위한 요청입니다.
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
 * agent 이름, 공개 범위, runtime 연결을 수정하기 위한 요청입니다.
 */
export interface UpdateAgentConfigurationRequest {
  description?: string;
  instruction?: string;
  model_id?: string;
  name?: string;
  profile_thumbnail_url?: string;
  runtime_id?: string;
  visibility?: AgentVisibility;
}

/**
 * 앱에서 사용하는 fetch 구현을 주입하기 위한 타입입니다.
 */
export type RiidoFetcher = typeof fetch;

/**
 * control-plane 호출에 필요한 기본 설정입니다.
 * `baseUrl`은 예: `http://ai-api.riido.io`처럼 마지막 슬래시 없이 전달해도 됩니다.
 * `fetcher`는 테스트나 앱 공통 transport 래핑이 필요할 때만 주입합니다.
 */
export interface RiidoClientConfig {
  baseUrl: string;
  token: string;
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
      Authorization: `Bearer ${config.token}`,
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
      Authorization: `Bearer ${config.token}`,
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
 * web 또는 desktop webview client의 AI Agent 화면 초기 데이터를 조회합니다
 */
export async function getAIAgentClientBootstrap(config: RiidoClientConfig, options: RiidoRequestOptions = {}): Promise<ClientBootstrapResponse> {
  const path = "/v1/client/ai-agent/bootstrap";
  return riidoRequest<ClientBootstrapResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * web 또는 desktop webview client의 AI Agent 화면 초기 데이터를 조회합니다
 * cache tag: `aiAgent.bootstrap`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function getAIAgentClientBootstrapQueryKeyRoot(): readonly unknown[] {
  return ["aiAgent.bootstrap"] as const;
}

/**
 * web 또는 desktop webview client의 AI Agent 화면 초기 데이터를 조회합니다
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function getAIAgentClientBootstrapQueryKey(): readonly unknown[] {
  return [...getAIAgentClientBootstrapQueryKeyRoot()] as const;
}

/**
 * web 또는 desktop webview client의 AI Agent 화면 초기 데이터를 조회합니다
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
 * runtime 설정 화면에서 현재 device의 daemon 상세를 조회합니다
 * 경로 파라미터입니다.
 */
export interface GetAIAgentDeviceDaemonPathParams {
  device_id: string;
}

/**
 * runtime 설정 화면에서 현재 device의 daemon 상세를 조회합니다
 */
export async function getAIAgentDeviceDaemon(config: RiidoClientConfig, params: GetAIAgentDeviceDaemonPathParams, options: RiidoRequestOptions = {}): Promise<DeviceDaemonDetailResponse> {
  const path = `/v1/client/ai-agent/devices/${params.device_id}/daemon`;
  return riidoRequest<DeviceDaemonDetailResponse>(config, path, { method: 'GET', signal: options.signal });
}

/**
 * runtime 설정 화면에서 현재 device의 daemon 상세를 조회합니다
 * cache tag: `aiAgent.devices.daemon`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function getAIAgentDeviceDaemonQueryKeyRoot(): readonly unknown[] {
  return ["aiAgent.devices.daemon"] as const;
}

/**
 * runtime 설정 화면에서 현재 device의 daemon 상세를 조회합니다
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function getAIAgentDeviceDaemonQueryKey(params: GetAIAgentDeviceDaemonPathParams): readonly unknown[] {
  return [...getAIAgentDeviceDaemonQueryKeyRoot(), params] as const;
}

/**
 * runtime 설정 화면에서 현재 device의 daemon 상세를 조회합니다
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
 * runtime 설정 화면에서 현재 device의 daemon 재시작을 요청합니다
 * 경로 파라미터입니다.
 */
export interface RestartAIAgentDeviceDaemonPathParams {
  device_id: string;
}

/**
 * runtime 설정 화면에서 현재 device의 daemon 재시작을 요청합니다
 */
export async function restartAIAgentDeviceDaemon(config: RiidoClientConfig, params: RestartAIAgentDeviceDaemonPathParams, body: ControlDeviceDaemonRequest, options: RiidoRequestOptions = {}): Promise<DeviceDaemonCommandResponse> {
  const path = `/v1/client/ai-agent/devices/${params.device_id}/daemon/restart`;
  return riidoRequest<DeviceDaemonCommandResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * runtime 설정 화면에서 현재 device의 daemon 재시작을 요청합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface RestartAIAgentDeviceDaemonMutationVariables {
  params: RestartAIAgentDeviceDaemonPathParams;
  body: ControlDeviceDaemonRequest;
}

/**
 * runtime 설정 화면에서 현재 device의 daemon 재시작을 요청합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function restartAIAgentDeviceDaemonMutationKey(): readonly unknown[] {
  return ["restartAIAgentDeviceDaemon"] as const;
}

/**
 * runtime 설정 화면에서 현재 device의 daemon 재시작을 요청합니다
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
 * runtime 설정 화면에서 현재 device의 daemon 시작을 요청합니다
 * 경로 파라미터입니다.
 */
export interface StartAIAgentDeviceDaemonPathParams {
  device_id: string;
}

/**
 * runtime 설정 화면에서 현재 device의 daemon 시작을 요청합니다
 */
export async function startAIAgentDeviceDaemon(config: RiidoClientConfig, params: StartAIAgentDeviceDaemonPathParams, body: ControlDeviceDaemonRequest, options: RiidoRequestOptions = {}): Promise<DeviceDaemonCommandResponse> {
  const path = `/v1/client/ai-agent/devices/${params.device_id}/daemon/start`;
  return riidoRequest<DeviceDaemonCommandResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * runtime 설정 화면에서 현재 device의 daemon 시작을 요청합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface StartAIAgentDeviceDaemonMutationVariables {
  params: StartAIAgentDeviceDaemonPathParams;
  body: ControlDeviceDaemonRequest;
}

/**
 * runtime 설정 화면에서 현재 device의 daemon 시작을 요청합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function startAIAgentDeviceDaemonMutationKey(): readonly unknown[] {
  return ["startAIAgentDeviceDaemon"] as const;
}

/**
 * runtime 설정 화면에서 현재 device의 daemon 시작을 요청합니다
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
 * runtime 설정 화면에서 현재 device의 daemon 중지를 요청합니다
 * 경로 파라미터입니다.
 */
export interface StopAIAgentDeviceDaemonPathParams {
  device_id: string;
}

/**
 * runtime 설정 화면에서 현재 device의 daemon 중지를 요청합니다
 */
export async function stopAIAgentDeviceDaemon(config: RiidoClientConfig, params: StopAIAgentDeviceDaemonPathParams, body: ControlDeviceDaemonRequest, options: RiidoRequestOptions = {}): Promise<DeviceDaemonCommandResponse> {
  const path = `/v1/client/ai-agent/devices/${params.device_id}/daemon/stop`;
  return riidoRequest<DeviceDaemonCommandResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * runtime 설정 화면에서 현재 device의 daemon 중지를 요청합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface StopAIAgentDeviceDaemonMutationVariables {
  params: StopAIAgentDeviceDaemonPathParams;
  body: ControlDeviceDaemonRequest;
}

/**
 * runtime 설정 화면에서 현재 device의 daemon 중지를 요청합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function stopAIAgentDeviceDaemonMutationKey(): readonly unknown[] {
  return ["stopAIAgentDeviceDaemon"] as const;
}

/**
 * runtime 설정 화면에서 현재 device의 daemon 중지를 요청합니다
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
 * task thread comment를 할당된 AI agent에게 전달합니다
 * 경로 파라미터입니다.
 */
export interface SubmitAIAgentTaskCommentPathParams {
  task_id: string;
}

/**
 * task thread comment를 할당된 AI agent에게 전달합니다
 */
export async function submitAIAgentTaskComment(config: RiidoClientConfig, params: SubmitAIAgentTaskCommentPathParams, body: SubmitAIAgentTaskCommentRequest, options: RiidoRequestOptions = {}): Promise<AIAgentTaskActionResponse> {
  const path = `/v1/client/ai-agent/tasks/${params.task_id}/comments`;
  return riidoRequest<AIAgentTaskActionResponse>(config, path, { method: 'POST', signal: options.signal, body: JSON.stringify(body) });
}

/**
 * task thread comment를 할당된 AI agent에게 전달합니다
 * mutation 함수에 전달하는 변수입니다.
 */
export interface SubmitAIAgentTaskCommentMutationVariables {
  params: SubmitAIAgentTaskCommentPathParams;
  body: SubmitAIAgentTaskCommentRequest;
}

/**
 * task thread comment를 할당된 AI agent에게 전달합니다
 * 이 mutation을 구분하는 React Query mutation key입니다.
 */
export function submitAIAgentTaskCommentMutationKey(): readonly unknown[] {
  return ["submitAIAgentTaskComment"] as const;
}

/**
 * task thread comment를 할당된 AI agent에게 전달합니다
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
 * Figma agent 추가 화면에서 AI agent 설정을 생성합니다
 * DSL facade path: `aiAgent.agents.create`
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
 * DSL facade path: `aiAgent.agents.delete`
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
 * DSL facade path: `aiAgent.agents.updateConfiguration`
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
 * client control 활성화 전에 agent 수정 가능 여부를 조회합니다
 * DSL facade path: `aiAgent.agents.editability`
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
 * web 또는 desktop webview client의 AI Agent 화면 초기 데이터를 조회합니다
 * DSL facade path: `aiAgent.bootstrap`
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
 * DSL facade path: `aiAgent.devices.runtimes`
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
 * runtime 설정 화면에서 현재 device의 daemon 상세를 조회합니다
 * DSL facade path: `aiAgent.devices.daemon.details`
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
 * runtime 설정 화면에서 현재 device의 daemon 재시작을 요청합니다
 * DSL facade path: `aiAgent.devices.daemon.restart`
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
 * runtime 설정 화면에서 현재 device의 daemon 시작을 요청합니다
 * DSL facade path: `aiAgent.devices.daemon.start`
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
 * runtime 설정 화면에서 현재 device의 daemon 중지를 요청합니다
 * DSL facade path: `aiAgent.devices.daemon.stop`
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
 * DSL facade path: `aiAgent.events.stream`
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
 * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다
 * DSL facade path: `aiAgent.tasks.assignableAgents`
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
 * DSL facade path: `aiAgent.tasks.unassign`
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
 * DSL facade path: `aiAgent.tasks.assign`
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
 * task thread comment를 할당된 AI agent에게 전달합니다
 * DSL facade path: `aiAgent.tasks.submitComment`
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
 * DSL facade path: `aiAgent.tasks.stop`
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
 * DSL facade path: `aiAgent.tasks.threads`
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
 * agent 설정과 mutation을 다루는 namespace입니다. agent 이름은 중복될 수 있고 assigned task가 있으면 수정할 수 없습니다.
 */
export interface RiidoAIAgentAgentsNamespace {
  /**
   * Figma agent 추가 화면에서 AI agent 설정을 생성합니다 invalidates: `aiAgent.bootstrap`, `aiAgent.devices.runtimes`, `aiAgent.tasks.assignableAgents`
   */
  readonly create: CreateAIAgentEndpoint;
  /**
   * agent를 삭제하고 queued/running assignment를 강제로 정리합니다 invalidates: `aiAgent.bootstrap`, `aiAgent.devices.runtimes`, `aiAgent.agents.editability`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly delete: DeleteAIAgentEndpoint;
  /**
   * client control 활성화 전에 agent 수정 가능 여부를 조회합니다 cache tag: `aiAgent.agents.editability`
   */
  readonly editability: GetAIAgentEditabilityEndpoint;
  /**
   * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다 invalidates: `aiAgent.bootstrap`, `aiAgent.agents.editability`, `aiAgent.tasks.assignableAgents`
   */
  readonly updateConfiguration: UpdateAIAgentConfigurationEndpoint;
}

/**
 * runtime 설정 화면에서 현재 device의 desktop local daemon 상세와 제어 command를 다루는 namespace입니다.
 */
export interface RiidoAIAgentDevicesDaemonNamespace {
  /**
   * runtime 설정 화면에서 현재 device의 daemon 상세를 조회합니다 cache tag: `aiAgent.devices.daemon`
   */
  readonly details: GetAIAgentDeviceDaemonEndpoint;
  /**
   * runtime 설정 화면에서 현재 device의 daemon 재시작을 요청합니다 invalidates: `aiAgent.devices.daemon`, `aiAgent.devices.runtimes`
   */
  readonly restart: RestartAIAgentDeviceDaemonEndpoint;
  /**
   * runtime 설정 화면에서 현재 device의 daemon 시작을 요청합니다 invalidates: `aiAgent.devices.daemon`, `aiAgent.devices.runtimes`
   */
  readonly start: StartAIAgentDeviceDaemonEndpoint;
  /**
   * runtime 설정 화면에서 현재 device의 daemon 중지를 요청합니다 invalidates: `aiAgent.devices.daemon`, `aiAgent.devices.runtimes`
   */
  readonly stop: StopAIAgentDeviceDaemonEndpoint;
}

/**
 * device와 runtime 상태를 다루는 namespace입니다.
 */
export interface RiidoAIAgentDevicesNamespace {
  /**
   * runtime 설정 화면에서 현재 device의 desktop local daemon 상세와 제어 command를 다루는 namespace입니다.
   */
  readonly daemon: RiidoAIAgentDevicesDaemonNamespace;
  /**
   * 권한이 확인된 principal의 device runtime 상태를 조회합니다 cache tag: `aiAgent.devices.runtimes`
   */
  readonly runtimes: ListAIAgentDeviceRuntimesEndpoint;
}

/**
 * client가 SSE로 수신하는 상태 변경 stream namespace입니다.
 */
export interface RiidoAIAgentEventsNamespace {
  /**
   * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다 cache tag: `aiAgent.events.stream`
   */
  readonly stream: StreamAIAgentClientEventsEndpoint;
}

/**
 * task thread에서 AI Agent assignment와 comment action을 다루는 namespace입니다.
 */
export interface RiidoAIAgentTasksNamespace {
  /**
   * task participant dropdown에서 AI agent를 배정합니다 invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly assign: AssignAIAgentTaskEndpoint;
  /**
   * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다 cache tag: `aiAgent.tasks.assignableAgents`
   */
  readonly assignableAgents: ListAIAgentTaskAssignableAgentsEndpoint;
  /**
   * task thread의 stop action으로 AI agent 작업을 중단합니다 invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly stop: StopAIAgentTaskEndpoint;
  /**
   * task thread comment를 할당된 AI agent에게 전달합니다 invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly submitComment: SubmitAIAgentTaskCommentEndpoint;
  /**
   * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다 cache tag: `aiAgent.tasks.threads`
   */
  readonly threads: ListAIAgentTaskThreadsEndpoint;
  /**
   * task participant에서 AI agent를 제거하고 작업을 중단합니다 invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
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
   * web 또는 desktop webview client의 AI Agent 화면 초기 데이터를 조회합니다 cache tag: `aiAgent.bootstrap`
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
   * task thread에서 AI Agent assignment와 comment action을 다루는 namespace입니다.
   */
  readonly tasks: RiidoAIAgentTasksNamespace;
}

/**
 * control-plane API를 DSL client metadata의 module/namespace별로 묶은 config-bound facade입니다.
 */
export interface RiidoControlPlaneClient {
  /**
   * AI Agent client API module입니다. 화면 구성을 소유하지 않고, client가 화면을 구성할 때 필요한 API 의미, 정책, lifecycle, cache 관계를 제공합니다.
   */
  readonly aiAgent: RiidoAIAgentModule;
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
  };
}
