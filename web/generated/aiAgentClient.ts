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
  work_status: AgentWorkStatus;
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
  editability: AgentEditability;
  is_owned_by_viewer: boolean;
  name: string;
  owner_principal_id: string;
  runtime_id?: string;
  runtime_kind?: RuntimeKind;
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
 * Figma comment 소통 흐름에서 task thread에 기록되는 상태 update 종류입니다.
 */
export type AgentTaskCommentKind = "queued_by_busy_agent" | "stopped_by_agent_deleted" | "stopped_by_user_request" | "runtime_progress" | "task_completed" | "task_failed";

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
  work_status: AgentWorkStatus;
}

/**
 * AI Agent 화면 진입 시 필요한 agent와 device runtime 초기 데이터입니다.
 */
export interface ClientBootstrapResponse {
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
export type ClientStreamEvent = DeviceRuntimeSnapshotEvent | AgentEditabilityChangedEvent | AgentWorkStatusChangedEvent;

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
 * device에 설치되었거나 assignment 정책상 offline으로 유지되는 runtime record입니다.
 */
export interface RuntimeRecord {
  availability: RuntimeAvailability;
  detection_state: RuntimeDetectionState;
  device_id: string;
  has_assigned_agent: boolean;
  kind: RuntimeKind;
  last_detected_at?: string;
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
 * agent 이름, 공개 범위, runtime 연결을 수정하기 위한 요청입니다.
 */
export interface UpdateAgentConfigurationRequest {
  name?: string;
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
 * editability, work status, runtime snapshot에 대한 AI Agent client update를 스트리밍합니다
 */
export async function streamAIAgentClientEvents(config: RiidoClientConfig, options: RiidoRequestOptions = {}): Promise<Response> {
  const path = "/v1/client/ai-agent/events";
  return riidoRawRequest(config, path, { method: 'GET', signal: options.signal });
}

/**
 * editability, work status, runtime snapshot에 대한 AI Agent client update를 스트리밍합니다
 * cache tag: `aiAgent.events.stream`
 * 이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.
 */
export function streamAIAgentClientEventsQueryKeyRoot(): readonly unknown[] {
  return ["aiAgent.events.stream"] as const;
}

/**
 * editability, work status, runtime snapshot에 대한 AI Agent client update를 스트리밍합니다
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function streamAIAgentClientEventsQueryKey(): readonly unknown[] {
  return [...streamAIAgentClientEventsQueryKeyRoot()] as const;
}

/**
 * editability, work status, runtime snapshot에 대한 AI Agent client update를 스트리밍합니다
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
 * editability, work status, runtime snapshot에 대한 AI Agent client update를 스트리밍합니다
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
     * 선언된 모든 cache tag를 한 번에 무효화합니다.
     */
    readonly all: (queryClient: QueryClient) => Promise<void[]>;
  };
}

/**
 * agent 설정과 mutation을 다루는 namespace입니다. agent 이름은 중복될 수 있고 assigned task가 있으면 수정할 수 없습니다.
 */
export interface RiidoAIAgentAgentsNamespace {
  /**
   * agent를 삭제하고 queued/running assignment를 강제로 정리합니다 invalidates: `aiAgent.bootstrap`, `aiAgent.devices.runtimes`, `aiAgent.agents.editability`, `aiAgent.tasks.assignableAgents`
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
 * device와 runtime 상태를 다루는 namespace입니다.
 */
export interface RiidoAIAgentDevicesNamespace {
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
   * editability, work status, runtime snapshot에 대한 AI Agent client update를 스트리밍합니다 cache tag: `aiAgent.events.stream`
   */
  readonly stream: StreamAIAgentClientEventsEndpoint;
}

/**
 * task thread에서 AI Agent assignment와 comment action을 다루는 namespace입니다.
 */
export interface RiidoAIAgentTasksNamespace {
  /**
   * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다 cache tag: `aiAgent.tasks.assignableAgents`
   */
  readonly assignableAgents: ListAIAgentTaskAssignableAgentsEndpoint;
  /**
   * task thread의 stop action으로 AI agent 작업을 중단합니다 invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`
   */
  readonly stop: StopAIAgentTaskEndpoint;
  /**
   * task thread comment를 할당된 AI agent에게 전달합니다 invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`
   */
  readonly submitComment: SubmitAIAgentTaskCommentEndpoint;
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
            all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentDeviceRuntimesQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: getAIAgentEditabilityQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() })]),
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
            all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() })]),
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
            all: (queryClient: QueryClient) => Promise.all([queryClient.invalidateQueries({ queryKey: getAIAgentClientBootstrapQueryKeyRoot() }), queryClient.invalidateQueries({ queryKey: listAIAgentTaskAssignableAgentsQueryKeyRoot() })]),
          },
        },
      },
    },
  };
}
