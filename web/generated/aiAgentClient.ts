// 이 파일은 tools/reactquerygen으로 생성됩니다. 직접 수정하지 마세요.
// 원본: contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json

import { useMutation, useQuery, type UseMutationOptions, type UseQueryOptions } from '@tanstack/react-query';

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
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function deleteAIAgentQueryKey(params: DeleteAIAgentPathParams): readonly unknown[] {
  return ["deleteAIAgent", params] as const;
}

/**
 * agent를 삭제하고 queued/running assignment를 강제로 정리합니다
 * React Query mutation hook입니다.
 */
export function useDeleteAIAgent(config: RiidoClientConfig, options: UseMutationOptions<DeleteAgentResponse, Error, { params: DeleteAIAgentPathParams }> = {}) {
  return useMutation<DeleteAgentResponse, Error, { params: DeleteAIAgentPathParams }>({
    ...options,
    mutationFn: (variables) => deleteAIAgent(config, variables.params, {}),
  });
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
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function updateAIAgentConfigurationQueryKey(params: UpdateAIAgentConfigurationPathParams, body: UpdateAgentConfigurationRequest): readonly unknown[] {
  return ["updateAIAgentConfiguration", params, body] as const;
}

/**
 * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다
 * React Query mutation hook입니다.
 */
export function useUpdateAIAgentConfiguration(config: RiidoClientConfig, options: UseMutationOptions<AgentClientRecordResponse, Error, { params: UpdateAIAgentConfigurationPathParams; body: UpdateAgentConfigurationRequest }> = {}) {
  return useMutation<AgentClientRecordResponse, Error, { params: UpdateAIAgentConfigurationPathParams; body: UpdateAgentConfigurationRequest }>({
    ...options,
    mutationFn: (variables) => updateAIAgentConfiguration(config, variables.params, variables.body, {}),
  });
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
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function getAIAgentEditabilityQueryKey(params: GetAIAgentEditabilityPathParams): readonly unknown[] {
  return ["getAIAgentEditability", params] as const;
}

/**
 * client control 활성화 전에 agent 수정 가능 여부를 조회합니다
 * React Query query hook입니다.
 */
export function useGetAIAgentEditability(config: RiidoClientConfig, params: GetAIAgentEditabilityPathParams, options: Omit<UseQueryOptions<AgentEditabilityResponse>, 'queryKey' | 'queryFn'> & RiidoRequestOptions = {}) {
  return useQuery<AgentEditabilityResponse>({
    ...options,
    queryKey: getAIAgentEditabilityQueryKey(params),
    queryFn: () => getAIAgentEditability(config, params, options),
  });
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
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function getAIAgentClientBootstrapQueryKey(): readonly unknown[] {
  return ["getAIAgentClientBootstrap"] as const;
}

/**
 * web 또는 desktop webview client의 AI Agent 화면 초기 데이터를 조회합니다
 * React Query query hook입니다.
 */
export function useGetAIAgentClientBootstrap(config: RiidoClientConfig, options: Omit<UseQueryOptions<ClientBootstrapResponse>, 'queryKey' | 'queryFn'> & RiidoRequestOptions = {}) {
  return useQuery<ClientBootstrapResponse>({
    ...options,
    queryKey: getAIAgentClientBootstrapQueryKey(),
    queryFn: () => getAIAgentClientBootstrap(config, options),
  });
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
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function listAIAgentDeviceRuntimesQueryKey(): readonly unknown[] {
  return ["listAIAgentDeviceRuntimes"] as const;
}

/**
 * 권한이 확인된 principal의 device runtime 상태를 조회합니다
 * React Query query hook입니다.
 */
export function useListAIAgentDeviceRuntimes(config: RiidoClientConfig, options: Omit<UseQueryOptions<DeviceRuntimeListResponse>, 'queryKey' | 'queryFn'> & RiidoRequestOptions = {}) {
  return useQuery<DeviceRuntimeListResponse>({
    ...options,
    queryKey: listAIAgentDeviceRuntimesQueryKey(),
    queryFn: () => listAIAgentDeviceRuntimes(config, options),
  });
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
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function streamAIAgentClientEventsQueryKey(): readonly unknown[] {
  return ["streamAIAgentClientEvents"] as const;
}

/**
 * editability, work status, runtime snapshot에 대한 AI Agent client update를 스트리밍합니다
 * React Query query hook입니다.
 */
export function useStreamAIAgentClientEvents(config: RiidoClientConfig, options: Omit<UseQueryOptions<Response>, 'queryKey' | 'queryFn'> & RiidoRequestOptions = {}) {
  return useQuery<Response>({
    ...options,
    queryKey: streamAIAgentClientEventsQueryKey(),
    queryFn: () => streamAIAgentClientEvents(config, options),
  });
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
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function listAIAgentTaskAssignableAgentsQueryKey(params: ListAIAgentTaskAssignableAgentsPathParams): readonly unknown[] {
  return ["listAIAgentTaskAssignableAgents", params] as const;
}

/**
 * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다
 * React Query query hook입니다.
 */
export function useListAIAgentTaskAssignableAgents(config: RiidoClientConfig, params: ListAIAgentTaskAssignableAgentsPathParams, options: Omit<UseQueryOptions<AgentClientListResponse>, 'queryKey' | 'queryFn'> & RiidoRequestOptions = {}) {
  return useQuery<AgentClientListResponse>({
    ...options,
    queryKey: listAIAgentTaskAssignableAgentsQueryKey(params),
    queryFn: () => listAIAgentTaskAssignableAgents(config, params, options),
  });
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
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function submitAIAgentTaskCommentQueryKey(params: SubmitAIAgentTaskCommentPathParams, body: SubmitAIAgentTaskCommentRequest): readonly unknown[] {
  return ["submitAIAgentTaskComment", params, body] as const;
}

/**
 * task thread comment를 할당된 AI agent에게 전달합니다
 * React Query mutation hook입니다.
 */
export function useSubmitAIAgentTaskComment(config: RiidoClientConfig, options: UseMutationOptions<AIAgentTaskActionResponse, Error, { params: SubmitAIAgentTaskCommentPathParams; body: SubmitAIAgentTaskCommentRequest }> = {}) {
  return useMutation<AIAgentTaskActionResponse, Error, { params: SubmitAIAgentTaskCommentPathParams; body: SubmitAIAgentTaskCommentRequest }>({
    ...options,
    mutationFn: (variables) => submitAIAgentTaskComment(config, variables.params, variables.body, {}),
  });
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
 * 이 호출에 사용하는 React Query 키입니다.
 */
export function stopAIAgentTaskQueryKey(params: StopAIAgentTaskPathParams, body: StopAIAgentTaskRequest): readonly unknown[] {
  return ["stopAIAgentTask", params, body] as const;
}

/**
 * task thread의 stop action으로 AI agent 작업을 중단합니다
 * React Query mutation hook입니다.
 */
export function useStopAIAgentTask(config: RiidoClientConfig, options: UseMutationOptions<AIAgentTaskActionResponse, Error, { params: StopAIAgentTaskPathParams; body: StopAIAgentTaskRequest }> = {}) {
  return useMutation<AIAgentTaskActionResponse, Error, { params: StopAIAgentTaskPathParams; body: StopAIAgentTaskRequest }>({
    ...options,
    mutationFn: (variables) => stopAIAgentTask(config, variables.params, variables.body, {}),
  });
}
