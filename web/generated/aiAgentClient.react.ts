'use client';

// 이 파일은 tools/reactquerygen으로 생성됩니다. 직접 수정하지 마세요.
// 원본: contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json

/* eslint-disable react-hooks/rules-of-hooks */

import { useMemo } from 'react';
import { useMutation, useQuery, type UseMutationResult, type UseQueryResult } from '@/lib/react-query';
import * as core from './aiAgentClient';

/**
 * Figma agent 추가 화면에서 AI agent 설정을 생성합니다
 * 계약 generated path: `aiAgent.agents.create`
 * 검색용 generated 경로: `agents.create`
 * 접근 예시: `riido.aiAgent.agents.create`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface CreateAIAgentReactEndpoint extends core.CreateAIAgentEndpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.AgentClientRecordResponse, core.CreateAIAgentMutationVariables>) => UseMutationResult<core.AgentClientRecordResponse, Error, core.CreateAIAgentMutationVariables>;
}

/**
 * agent를 삭제하고 queued/running assignment를 강제로 정리합니다
 * 계약 generated path: `aiAgent.agents.delete`
 * 검색용 generated 경로: `agents.delete`
 * 접근 예시: `riido.aiAgent.agents.delete`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface DeleteAIAgentReactEndpoint extends core.DeleteAIAgentEndpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.DeleteAgentResponse, core.DeleteAIAgentMutationVariables>) => UseMutationResult<core.DeleteAgentResponse, Error, core.DeleteAIAgentMutationVariables>;
}

/**
 * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다
 * 계약 generated path: `aiAgent.agents.updateConfiguration`
 * 검색용 generated 경로: `agents.updateConfiguration`
 * 접근 예시: `riido.aiAgent.agents.updateConfiguration`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface UpdateAIAgentConfigurationReactEndpoint extends core.UpdateAIAgentConfigurationEndpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.AgentClientRecordResponse, core.UpdateAIAgentConfigurationMutationVariables>) => UseMutationResult<core.AgentClientRecordResponse, Error, core.UpdateAIAgentConfigurationMutationVariables>;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다
 * 계약 generated path: `aiAgent.agents.daemon.details`
 * 검색용 generated 경로: `agents.daemon.details`
 * 접근 예시: `riido.aiAgent.agents.daemon.details`
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface GetAIAgentDaemonReactEndpoint extends core.GetAIAgentDaemonEndpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (params: core.GetAIAgentDaemonPathParams, options?: core.RiidoQueryOptions<core.DeviceDaemonDetailResponse>) => UseQueryResult<core.DeviceDaemonDetailResponse, Error>;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다
 * 계약 generated path: `aiAgent.agents.daemon.restart`
 * 검색용 generated 경로: `agents.daemon.restart`
 * 접근 예시: `riido.aiAgent.agents.daemon.restart`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface RestartAIAgentDaemonReactEndpoint extends core.RestartAIAgentDaemonEndpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.DeviceDaemonCommandResponse, core.RestartAIAgentDaemonMutationVariables>) => UseMutationResult<core.DeviceDaemonCommandResponse, Error, core.RestartAIAgentDaemonMutationVariables>;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다
 * 계약 generated path: `aiAgent.agents.daemon.start`
 * 검색용 generated 경로: `agents.daemon.start`
 * 접근 예시: `riido.aiAgent.agents.daemon.start`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface StartAIAgentDaemonReactEndpoint extends core.StartAIAgentDaemonEndpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.DeviceDaemonCommandResponse, core.StartAIAgentDaemonMutationVariables>) => UseMutationResult<core.DeviceDaemonCommandResponse, Error, core.StartAIAgentDaemonMutationVariables>;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다
 * 계약 generated path: `aiAgent.agents.daemon.stop`
 * 검색용 generated 경로: `agents.daemon.stop`
 * 접근 예시: `riido.aiAgent.agents.daemon.stop`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface StopAIAgentDaemonReactEndpoint extends core.StopAIAgentDaemonEndpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.DeviceDaemonCommandResponse, core.StopAIAgentDaemonMutationVariables>) => UseMutationResult<core.DeviceDaemonCommandResponse, Error, core.StopAIAgentDaemonMutationVariables>;
}

/**
 * client control 활성화 전에 agent 수정 가능 여부를 조회합니다
 * 계약 generated path: `aiAgent.agents.editability`
 * 검색용 generated 경로: `agents.editability`
 * 접근 예시: `riido.aiAgent.agents.editability`
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface GetAIAgentEditabilityReactEndpoint extends core.GetAIAgentEditabilityEndpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (params: core.GetAIAgentEditabilityPathParams, options?: core.RiidoQueryOptions<core.AgentEditabilityResponse>) => UseQueryResult<core.AgentEditabilityResponse, Error>;
}

/**
 * web 또는 desktop webview client의 AI Agent 설정/온보딩 초기 데이터와 bootstrap.agents[] agent list를 조회합니다
 * 계약 generated path: `aiAgent.bootstrap`
 * 검색용 generated 경로: `bootstrap`
 * 접근 예시: `riido.aiAgent.bootstrap`
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface GetAIAgentClientBootstrapReactEndpoint extends core.GetAIAgentClientBootstrapEndpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (options?: core.RiidoQueryOptions<core.ClientBootstrapResponse>) => UseQueryResult<core.ClientBootstrapResponse, Error>;
}

/**
 * 권한이 확인된 principal의 device runtime 상태를 조회합니다
 * 계약 generated path: `aiAgent.devices.runtimes`
 * 검색용 generated 경로: `devices.runtimes`
 * 접근 예시: `riido.aiAgent.devices.runtimes`
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface ListAIAgentDeviceRuntimesReactEndpoint extends core.ListAIAgentDeviceRuntimesEndpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (options?: core.RiidoQueryOptions<core.DeviceRuntimeListResponse>) => UseQueryResult<core.DeviceRuntimeListResponse, Error>;
}

/**
 * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다
 * 계약 generated path: `aiAgent.events.stream`
 * 검색용 generated 경로: `events.stream`
 * 접근 예시: `riido.aiAgent.events.stream`
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface StreamAIAgentClientEventsReactEndpoint extends core.StreamAIAgentClientEventsEndpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (options?: core.RiidoQueryOptions<Response>) => UseQueryResult<Response, Error>;
}

/**
 * AI Agent 온보딩에서 사용할 서버 제공 fixture 목록을 조회합니다
 * 계약 generated path: `aiAgent.onboarding.fixtures`
 * 검색용 generated 경로: `onboarding.fixtures`
 * 접근 예시: `riido.aiAgent.onboarding.fixtures`
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface ListAIAgentOnboardingFixturesReactEndpoint extends core.ListAIAgentOnboardingFixturesEndpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (options?: core.RiidoQueryOptions<core.AgentOnboardingFixtureListResponse>) => UseQueryResult<core.AgentOnboardingFixtureListResponse, Error>;
}

/**
 * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다
 * 계약 generated path: `aiAgent.onboarding.fixtures.createAgent`
 * 검색용 generated 경로: `onboarding.fixtures.createAgent`
 * 접근 예시: `riido.aiAgent.onboarding.fixtures.createAgent`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface CreateAIAgentFromOnboardingFixtureReactEndpoint extends core.CreateAIAgentFromOnboardingFixtureEndpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.AgentClientRecordResponse, core.CreateAIAgentFromOnboardingFixtureMutationVariables>) => UseMutationResult<core.AgentClientRecordResponse, Error, core.CreateAIAgentFromOnboardingFixtureMutationVariables>;
}

/**
 * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다
 * 계약 generated path: `aiAgent.tasks.assignableAgents`
 * 검색용 generated 경로: `tasks.assignableAgents`
 * 접근 예시: `riido.aiAgent.tasks.assignableAgents`
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface ListAIAgentTaskAssignableAgentsReactEndpoint extends core.ListAIAgentTaskAssignableAgentsEndpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (params: core.ListAIAgentTaskAssignableAgentsPathParams, options?: core.RiidoQueryOptions<core.AgentClientListResponse>) => UseQueryResult<core.AgentClientListResponse, Error>;
}

/**
 * task participant에서 AI agent를 제거하고 작업을 중단합니다
 * 계약 generated path: `aiAgent.tasks.unassign`
 * 검색용 generated 경로: `tasks.unassign`
 * 접근 예시: `riido.aiAgent.tasks.unassign`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface UnassignAIAgentTaskReactEndpoint extends core.UnassignAIAgentTaskEndpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.UnassignAIAgentTaskMutationVariables>) => UseMutationResult<core.AIAgentTaskActionResponse, Error, core.UnassignAIAgentTaskMutationVariables>;
}

/**
 * task participant dropdown에서 AI agent를 배정합니다
 * 계약 generated path: `aiAgent.tasks.assign`
 * 검색용 generated 경로: `tasks.assign`
 * 접근 예시: `riido.aiAgent.tasks.assign`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface AssignAIAgentTaskReactEndpoint extends core.AssignAIAgentTaskEndpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.AssignAIAgentTaskMutationVariables>) => UseMutationResult<core.AIAgentTaskActionResponse, Error, core.AssignAIAgentTaskMutationVariables>;
}

/**
 * 호환 task comment route로 AI agent에게 메시지를 전달합니다
 * 계약 generated path: `aiAgent.tasks.submitComment`
 * 검색용 generated 경로: `tasks.submitComment`
 * 접근 예시: `riido.aiAgent.tasks.submitComment`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface SubmitAIAgentTaskCommentReactEndpoint extends core.SubmitAIAgentTaskCommentEndpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.SubmitAIAgentTaskCommentMutationVariables>) => UseMutationResult<core.AIAgentTaskActionResponse, Error, core.SubmitAIAgentTaskCommentMutationVariables>;
}

/**
 * task thread의 stop action으로 AI agent 작업을 중단합니다
 * 계약 generated path: `aiAgent.tasks.stop`
 * 검색용 generated 경로: `tasks.stop`
 * 접근 예시: `riido.aiAgent.tasks.stop`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface StopAIAgentTaskReactEndpoint extends core.StopAIAgentTaskEndpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.StopAIAgentTaskMutationVariables>) => UseMutationResult<core.AIAgentTaskActionResponse, Error, core.StopAIAgentTaskMutationVariables>;
}

/**
 * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다
 * 계약 generated path: `aiAgent.tasks.threads`
 * 검색용 generated 경로: `tasks.threads`
 * 접근 예시: `riido.aiAgent.tasks.threads`
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface ListAIAgentTaskThreadsReactEndpoint extends core.ListAIAgentTaskThreadsEndpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (params: core.ListAIAgentTaskThreadsPathParams, options?: core.RiidoQueryOptions<core.AIAgentTaskThreadCollectionResponse>) => UseQueryResult<core.AIAgentTaskThreadCollectionResponse, Error>;
}

/**
 * task thread message로 AI agent에게 다음 작업 지시를 전달합니다
 * 계약 generated path: `aiAgent.tasks.threadMessages.create`
 * 검색용 generated 경로: `tasks.threadMessages.create`
 * 접근 예시: `riido.aiAgent.tasks.threadMessages.create`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface CreateAIAgentTaskThreadMessageReactEndpoint extends core.CreateAIAgentTaskThreadMessageEndpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.CreateAIAgentTaskThreadMessageMutationVariables>) => UseMutationResult<core.AIAgentTaskActionResponse, Error, core.CreateAIAgentTaskThreadMessageMutationVariables>;
}

/**
 * Figma agent 추가 화면에서 AI agent 설정을 생성합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.agents.create`
 * 검색용 generated 경로: `aiAgent.agents.create`
 * 접근 예시: `riido.v2.aiAgent.agents.create`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface CreateAIAgentV2ReactEndpoint extends core.CreateAIAgentV2Endpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.AgentClientRecordResponseV2, core.CreateAIAgentV2MutationVariables>) => UseMutationResult<core.AgentClientRecordResponseV2, Error, core.CreateAIAgentV2MutationVariables>;
}

/**
 * agent를 삭제하고 queued/running assignment를 강제로 정리합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.agents.delete`
 * 검색용 generated 경로: `aiAgent.agents.delete`
 * 접근 예시: `riido.v2.aiAgent.agents.delete`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface DeleteAIAgentV2ReactEndpoint extends core.DeleteAIAgentV2Endpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.DeleteAgentResponse, core.DeleteAIAgentV2MutationVariables>) => UseMutationResult<core.DeleteAgentResponse, Error, core.DeleteAIAgentV2MutationVariables>;
}

/**
 * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.agents.updateConfiguration`
 * 검색용 generated 경로: `aiAgent.agents.updateConfiguration`
 * 접근 예시: `riido.v2.aiAgent.agents.updateConfiguration`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface UpdateAIAgentConfigurationV2ReactEndpoint extends core.UpdateAIAgentConfigurationV2Endpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.AgentClientRecordResponseV2, core.UpdateAIAgentConfigurationV2MutationVariables>) => UseMutationResult<core.AgentClientRecordResponseV2, Error, core.UpdateAIAgentConfigurationV2MutationVariables>;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.agents.daemon.details`
 * 검색용 generated 경로: `aiAgent.agents.daemon.details`
 * 접근 예시: `riido.v2.aiAgent.agents.daemon.details`
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface GetAIAgentDaemonV2ReactEndpoint extends core.GetAIAgentDaemonV2Endpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (params: core.GetAIAgentDaemonV2PathParams, options?: core.RiidoQueryOptions<core.DeviceDaemonDetailResponse>) => UseQueryResult<core.DeviceDaemonDetailResponse, Error>;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.agents.daemon.restart`
 * 검색용 generated 경로: `aiAgent.agents.daemon.restart`
 * 접근 예시: `riido.v2.aiAgent.agents.daemon.restart`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface RestartAIAgentDaemonV2ReactEndpoint extends core.RestartAIAgentDaemonV2Endpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.DeviceDaemonCommandResponse, core.RestartAIAgentDaemonV2MutationVariables>) => UseMutationResult<core.DeviceDaemonCommandResponse, Error, core.RestartAIAgentDaemonV2MutationVariables>;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.agents.daemon.start`
 * 검색용 generated 경로: `aiAgent.agents.daemon.start`
 * 접근 예시: `riido.v2.aiAgent.agents.daemon.start`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface StartAIAgentDaemonV2ReactEndpoint extends core.StartAIAgentDaemonV2Endpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.DeviceDaemonCommandResponse, core.StartAIAgentDaemonV2MutationVariables>) => UseMutationResult<core.DeviceDaemonCommandResponse, Error, core.StartAIAgentDaemonV2MutationVariables>;
}

/**
 * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.agents.daemon.stop`
 * 검색용 generated 경로: `aiAgent.agents.daemon.stop`
 * 접근 예시: `riido.v2.aiAgent.agents.daemon.stop`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface StopAIAgentDaemonV2ReactEndpoint extends core.StopAIAgentDaemonV2Endpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.DeviceDaemonCommandResponse, core.StopAIAgentDaemonV2MutationVariables>) => UseMutationResult<core.DeviceDaemonCommandResponse, Error, core.StopAIAgentDaemonV2MutationVariables>;
}

/**
 * client control 활성화 전에 agent 수정 가능 여부를 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.agents.editability`
 * 검색용 generated 경로: `aiAgent.agents.editability`
 * 접근 예시: `riido.v2.aiAgent.agents.editability`
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface GetAIAgentEditabilityV2ReactEndpoint extends core.GetAIAgentEditabilityV2Endpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (params: core.GetAIAgentEditabilityV2PathParams, options?: core.RiidoQueryOptions<core.AgentEditabilityResponse>) => UseQueryResult<core.AgentEditabilityResponse, Error>;
}

/**
 * web 또는 desktop webview client의 AI Agent 설정/온보딩 초기 데이터와 v2.aiAgent.bootstrap.agents[] agent list를 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.bootstrap`
 * 검색용 generated 경로: `aiAgent.bootstrap`
 * 접근 예시: `riido.v2.aiAgent.bootstrap`
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface GetAIAgentClientBootstrapV2ReactEndpoint extends core.GetAIAgentClientBootstrapV2Endpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (params: core.GetAIAgentClientBootstrapV2PathParams, options?: core.RiidoQueryOptions<core.ClientBootstrapResponseV2>) => UseQueryResult<core.ClientBootstrapResponseV2, Error>;
}

/**
 * 권한이 확인된 principal의 device runtime 상태를 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.devices.runtimes`
 * 검색용 generated 경로: `aiAgent.devices.runtimes`
 * 접근 예시: `riido.v2.aiAgent.devices.runtimes`
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface ListAIAgentDeviceRuntimesV2ReactEndpoint extends core.ListAIAgentDeviceRuntimesV2Endpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (params: core.ListAIAgentDeviceRuntimesV2PathParams, options?: core.RiidoQueryOptions<core.DeviceRuntimeListResponse>) => UseQueryResult<core.DeviceRuntimeListResponse, Error>;
}

/**
 * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.events.stream`
 * 검색용 generated 경로: `aiAgent.events.stream`
 * 접근 예시: `riido.v2.aiAgent.events.stream`
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface StreamAIAgentClientEventsV2ReactEndpoint extends core.StreamAIAgentClientEventsV2Endpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (params: core.StreamAIAgentClientEventsV2PathParams, options?: core.RiidoQueryOptions<Response>) => UseQueryResult<Response, Error>;
}

/**
 * AI Agent 온보딩에서 사용할 서버 제공 fixture 목록을 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.onboarding.fixtures`
 * 검색용 generated 경로: `aiAgent.onboarding.fixtures`
 * 접근 예시: `riido.v2.aiAgent.onboarding.fixtures`
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface ListAIAgentOnboardingFixturesV2ReactEndpoint extends core.ListAIAgentOnboardingFixturesV2Endpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (params: core.ListAIAgentOnboardingFixturesV2PathParams, options?: core.RiidoQueryOptions<core.AgentOnboardingFixtureListResponse>) => UseQueryResult<core.AgentOnboardingFixtureListResponse, Error>;
}

/**
 * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.onboarding.fixtures.createAgent`
 * 검색용 generated 경로: `aiAgent.onboarding.fixtures.createAgent`
 * 접근 예시: `riido.v2.aiAgent.onboarding.fixtures.createAgent`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface CreateAIAgentFromOnboardingFixtureV2ReactEndpoint extends core.CreateAIAgentFromOnboardingFixtureV2Endpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.AgentClientRecordResponseV2, core.CreateAIAgentFromOnboardingFixtureV2MutationVariables>) => UseMutationResult<core.AgentClientRecordResponseV2, Error, core.CreateAIAgentFromOnboardingFixtureV2MutationVariables>;
}

/**
 * workspace task/card 목록에서 현재 배정된 AI Agent profile 표시값을 component_id key 해시맵으로 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.assignedAgentProfiles`
 * 검색용 generated 경로: `aiAgent.tasks.assignedAgentProfiles`
 * 접근 예시: `riido.v2.aiAgent.tasks.assignedAgentProfiles`
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface ListWorkspaceAssignedAgentProfilesV2ReactEndpoint extends core.ListWorkspaceAssignedAgentProfilesV2Endpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (params: core.ListWorkspaceAssignedAgentProfilesV2PathParams, options?: core.RiidoQueryOptions<core.AssignedAgentProfileMapResponse>) => UseQueryResult<core.AssignedAgentProfileMapResponse, Error>;
}

/**
 * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.assignableAgents`
 * 검색용 generated 경로: `aiAgent.tasks.assignableAgents`
 * 접근 예시: `riido.v2.aiAgent.tasks.assignableAgents`
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface ListAIAgentTaskAssignableAgentsV2ReactEndpoint extends core.ListAIAgentTaskAssignableAgentsV2Endpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (params: core.ListAIAgentTaskAssignableAgentsV2PathParams, options?: core.RiidoQueryOptions<core.AgentClientListResponseV2>) => UseQueryResult<core.AgentClientListResponseV2, Error>;
}

/**
 * task participant에서 AI agent를 제거하고 작업을 중단합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.unassign`
 * 검색용 generated 경로: `aiAgent.tasks.unassign`
 * 접근 예시: `riido.v2.aiAgent.tasks.unassign`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface UnassignAIAgentTaskV2ReactEndpoint extends core.UnassignAIAgentTaskV2Endpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.UnassignAIAgentTaskV2MutationVariables>) => UseMutationResult<core.AIAgentTaskActionResponse, Error, core.UnassignAIAgentTaskV2MutationVariables>;
}

/**
 * task participant dropdown에서 AI agent를 배정합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.assign`
 * 검색용 generated 경로: `aiAgent.tasks.assign`
 * 접근 예시: `riido.v2.aiAgent.tasks.assign`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface AssignAIAgentTaskV2ReactEndpoint extends core.AssignAIAgentTaskV2Endpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.AssignAIAgentTaskV2MutationVariables>) => UseMutationResult<core.AIAgentTaskActionResponse, Error, core.AssignAIAgentTaskV2MutationVariables>;
}

/**
 * 호환 task comment route로 AI agent에게 메시지를 전달합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.submitComment`
 * 검색용 generated 경로: `aiAgent.tasks.submitComment`
 * 접근 예시: `riido.v2.aiAgent.tasks.submitComment`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface SubmitAIAgentTaskCommentV2ReactEndpoint extends core.SubmitAIAgentTaskCommentV2Endpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.SubmitAIAgentTaskCommentV2MutationVariables>) => UseMutationResult<core.AIAgentTaskActionResponse, Error, core.SubmitAIAgentTaskCommentV2MutationVariables>;
}

/**
 * task thread의 stop action으로 AI agent 작업을 중단합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.stop`
 * 검색용 generated 경로: `aiAgent.tasks.stop`
 * 접근 예시: `riido.v2.aiAgent.tasks.stop`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface StopAIAgentTaskV2ReactEndpoint extends core.StopAIAgentTaskV2Endpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.StopAIAgentTaskV2MutationVariables>) => UseMutationResult<core.AIAgentTaskActionResponse, Error, core.StopAIAgentTaskV2MutationVariables>;
}

/**
 * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.threads`
 * 검색용 generated 경로: `aiAgent.tasks.threads`
 * 접근 예시: `riido.v2.aiAgent.tasks.threads`
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface ListAIAgentTaskThreadsV2ReactEndpoint extends core.ListAIAgentTaskThreadsV2Endpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (params: core.ListAIAgentTaskThreadsV2PathParams, options?: core.RiidoQueryOptions<core.AIAgentTaskThreadCollectionResponse>) => UseQueryResult<core.AIAgentTaskThreadCollectionResponse, Error>;
}

/**
 * task thread message로 AI agent에게 다음 작업 지시를 전달합니다 (v2 workspace-scoped)
 * 계약 generated path: `v2.aiAgent.tasks.threadMessages.create`
 * 검색용 generated 경로: `aiAgent.tasks.threadMessages.create`
 * 접근 예시: `riido.v2.aiAgent.tasks.threadMessages.create`
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface CreateAIAgentTaskThreadMessageV2ReactEndpoint extends core.CreateAIAgentTaskThreadMessageV2Endpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.CreateAIAgentTaskThreadMessageV2MutationVariables>) => UseMutationResult<core.AIAgentTaskActionResponse, Error, core.CreateAIAgentTaskThreadMessageV2MutationVariables>;
}

/**
 * agent visibility/access 권한을 통해 해당 agent에 연결된 desktop local daemon 상세와 제어 command를 다루는 namespace입니다.
 */
export interface RiidoAIAgentAgentsDaemonReactNamespace {
  /**
   * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다
   * 계약 generated path: `aiAgent.agents.daemon.details`
   * 검색용 generated 경로: `agents.daemon.details`
   * 접근 예시: `riido.aiAgent.agents.daemon.details`
   * cache tag: `aiAgent.agents.daemon`
   */
  readonly details: GetAIAgentDaemonReactEndpoint;
  /**
   * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다
   * 계약 generated path: `aiAgent.agents.daemon.restart`
   * 검색용 generated 경로: `agents.daemon.restart`
   * 접근 예시: `riido.aiAgent.agents.daemon.restart`
   * invalidates: `aiAgent.agents.daemon`, `aiAgent.devices.runtimes`
   */
  readonly restart: RestartAIAgentDaemonReactEndpoint;
  /**
   * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다
   * 계약 generated path: `aiAgent.agents.daemon.start`
   * 검색용 generated 경로: `agents.daemon.start`
   * 접근 예시: `riido.aiAgent.agents.daemon.start`
   * invalidates: `aiAgent.agents.daemon`, `aiAgent.devices.runtimes`
   */
  readonly start: StartAIAgentDaemonReactEndpoint;
  /**
   * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다
   * 계약 generated path: `aiAgent.agents.daemon.stop`
   * 검색용 generated 경로: `agents.daemon.stop`
   * 접근 예시: `riido.aiAgent.agents.daemon.stop`
   * invalidates: `aiAgent.agents.daemon`, `aiAgent.devices.runtimes`
   */
  readonly stop: StopAIAgentDaemonReactEndpoint;
}

/**
 * agent 설정과 mutation을 다루는 namespace입니다. agent 이름은 중복될 수 있고 assigned task가 있으면 수정할 수 없습니다.
 */
export interface RiidoAIAgentAgentsReactNamespace {
  /**
   * Figma agent 추가 화면에서 AI agent 설정을 생성합니다
   * 계약 generated path: `aiAgent.agents.create`
   * 검색용 generated 경로: `agents.create`
   * 접근 예시: `riido.aiAgent.agents.create`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.devices.runtimes`, `aiAgent.tasks.assignableAgents`
   */
  readonly create: CreateAIAgentReactEndpoint;
  /**
   * agent visibility/access 권한을 통해 해당 agent에 연결된 desktop local daemon 상세와 제어 command를 다루는 namespace입니다.
   */
  readonly daemon: RiidoAIAgentAgentsDaemonReactNamespace;
  /**
   * agent를 삭제하고 queued/running assignment를 강제로 정리합니다
   * 계약 generated path: `aiAgent.agents.delete`
   * 검색용 generated 경로: `agents.delete`
   * 접근 예시: `riido.aiAgent.agents.delete`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.devices.runtimes`, `aiAgent.agents.editability`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly delete: DeleteAIAgentReactEndpoint;
  /**
   * client control 활성화 전에 agent 수정 가능 여부를 조회합니다
   * 계약 generated path: `aiAgent.agents.editability`
   * 검색용 generated 경로: `agents.editability`
   * 접근 예시: `riido.aiAgent.agents.editability`
   * cache tag: `aiAgent.agents.editability`
   */
  readonly editability: GetAIAgentEditabilityReactEndpoint;
  /**
   * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다
   * 계약 generated path: `aiAgent.agents.updateConfiguration`
   * 검색용 generated 경로: `agents.updateConfiguration`
   * 접근 예시: `riido.aiAgent.agents.updateConfiguration`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.agents.editability`, `aiAgent.tasks.assignableAgents`
   */
  readonly updateConfiguration: UpdateAIAgentConfigurationReactEndpoint;
}

/**
 * device와 runtime 상태를 다루는 namespace입니다.
 */
export interface RiidoAIAgentDevicesReactNamespace {
  /**
   * 권한이 확인된 principal의 device runtime 상태를 조회합니다
   * 계약 generated path: `aiAgent.devices.runtimes`
   * 검색용 generated 경로: `devices.runtimes`
   * 접근 예시: `riido.aiAgent.devices.runtimes`
   * cache tag: `aiAgent.devices.runtimes`
   */
  readonly runtimes: ListAIAgentDeviceRuntimesReactEndpoint;
}

/**
 * client가 SSE로 수신하는 상태 변경 stream namespace입니다.
 */
export interface RiidoAIAgentEventsReactNamespace {
  /**
   * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다
   * 계약 generated path: `aiAgent.events.stream`
   * 검색용 generated 경로: `events.stream`
   * 접근 예시: `riido.aiAgent.events.stream`
   * cache tag: `aiAgent.events.stream`
   */
  readonly stream: StreamAIAgentClientEventsReactEndpoint;
}

/**
 * 리도, 영실, 홍도, 지원처럼 제품이 제공하는 고정 onboarding fixture 목록과 fixture 기반 agent 생성 진입점입니다.
 */
export interface RiidoAIAgentOnboardingFixturesReactNamespace extends ListAIAgentOnboardingFixturesReactEndpoint {
  /**
   * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다
   * 계약 generated path: `aiAgent.onboarding.fixtures.createAgent`
   * 검색용 generated 경로: `onboarding.fixtures.createAgent`
   * 접근 예시: `riido.aiAgent.onboarding.fixtures.createAgent`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.devices.runtimes`, `aiAgent.tasks.assignableAgents`
   */
  readonly createAgent: CreateAIAgentFromOnboardingFixtureReactEndpoint;
}

/**
 * AI Agent 온보딩에서 필요한 서버 제공 초기값을 다루는 namespace입니다. 템플릿 엔티티를 만들거나 관리하지 않습니다.
 */
export interface RiidoAIAgentOnboardingReactNamespace {
  /**
   * 리도, 영실, 홍도, 지원처럼 제품이 제공하는 고정 onboarding fixture 목록과 fixture 기반 agent 생성 진입점입니다.
   */
  readonly fixtures: RiidoAIAgentOnboardingFixturesReactNamespace;
}

/**
 * task thread에 사용자가 다음 작업 지시를 남기는 정식 message command namespace입니다. Figma의 댓글 표현은 이 thread message로 투영됩니다.
 */
export interface RiidoAIAgentTasksThreadMessagesReactNamespace {
  /**
   * task thread message로 AI agent에게 다음 작업 지시를 전달합니다
   * 계약 generated path: `aiAgent.tasks.threadMessages.create`
   * 검색용 generated 경로: `tasks.threadMessages.create`
   * 접근 예시: `riido.aiAgent.tasks.threadMessages.create`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly create: CreateAIAgentTaskThreadMessageReactEndpoint;
}

/**
 * task thread에서 AI Agent assignment, thread message, compatibility comment action을 다루는 namespace입니다.
 */
export interface RiidoAIAgentTasksReactNamespace {
  /**
   * task participant dropdown에서 AI agent를 배정합니다
   * 계약 generated path: `aiAgent.tasks.assign`
   * 검색용 generated 경로: `tasks.assign`
   * 접근 예시: `riido.aiAgent.tasks.assign`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly assign: AssignAIAgentTaskReactEndpoint;
  /**
   * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다
   * 계약 generated path: `aiAgent.tasks.assignableAgents`
   * 검색용 generated 경로: `tasks.assignableAgents`
   * 접근 예시: `riido.aiAgent.tasks.assignableAgents`
   * cache tag: `aiAgent.tasks.assignableAgents`
   */
  readonly assignableAgents: ListAIAgentTaskAssignableAgentsReactEndpoint;
  /**
   * task thread의 stop action으로 AI agent 작업을 중단합니다
   * 계약 generated path: `aiAgent.tasks.stop`
   * 검색용 generated 경로: `tasks.stop`
   * 접근 예시: `riido.aiAgent.tasks.stop`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly stop: StopAIAgentTaskReactEndpoint;
  /**
   * 호환 task comment route로 AI agent에게 메시지를 전달합니다
   * 계약 generated path: `aiAgent.tasks.submitComment`
   * 검색용 generated 경로: `tasks.submitComment`
   * 접근 예시: `riido.aiAgent.tasks.submitComment`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly submitComment: SubmitAIAgentTaskCommentReactEndpoint;
  /**
   * task thread에 사용자가 다음 작업 지시를 남기는 정식 message command namespace입니다. Figma의 댓글 표현은 이 thread message로 투영됩니다.
   */
  readonly threadMessages: RiidoAIAgentTasksThreadMessagesReactNamespace;
  /**
   * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다
   * 계약 generated path: `aiAgent.tasks.threads`
   * 검색용 generated 경로: `tasks.threads`
   * 접근 예시: `riido.aiAgent.tasks.threads`
   * cache tag: `aiAgent.tasks.threads`
   */
  readonly threads: ListAIAgentTaskThreadsReactEndpoint;
  /**
   * task participant에서 AI agent를 제거하고 작업을 중단합니다
   * 계약 generated path: `aiAgent.tasks.unassign`
   * 검색용 generated 경로: `tasks.unassign`
   * 접근 예시: `riido.aiAgent.tasks.unassign`
   * invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly unassign: UnassignAIAgentTaskReactEndpoint;
}

/**
 * AI Agent client API module입니다. 화면 구성을 소유하지 않고, client가 화면을 구성할 때 필요한 API 의미, 정책, lifecycle, cache 관계를 제공합니다.
 */
export interface RiidoAIAgentReactModule {
  /**
   * agent 설정과 mutation을 다루는 namespace입니다. agent 이름은 중복될 수 있고 assigned task가 있으면 수정할 수 없습니다.
   */
  readonly agents: RiidoAIAgentAgentsReactNamespace;
  /**
   * web 또는 desktop webview client의 AI Agent 설정/온보딩 초기 데이터와 bootstrap.agents[] agent list를 조회합니다
   * 계약 generated path: `aiAgent.bootstrap`
   * 검색용 generated 경로: `bootstrap`
   * 접근 예시: `riido.aiAgent.bootstrap`
   * cache tag: `aiAgent.bootstrap`
   */
  readonly bootstrap: GetAIAgentClientBootstrapReactEndpoint;
  /**
   * device와 runtime 상태를 다루는 namespace입니다.
   */
  readonly devices: RiidoAIAgentDevicesReactNamespace;
  /**
   * client가 SSE로 수신하는 상태 변경 stream namespace입니다.
   */
  readonly events: RiidoAIAgentEventsReactNamespace;
  /**
   * AI Agent 온보딩에서 필요한 서버 제공 초기값을 다루는 namespace입니다. 템플릿 엔티티를 만들거나 관리하지 않습니다.
   */
  readonly onboarding: RiidoAIAgentOnboardingReactNamespace;
  /**
   * task thread에서 AI Agent assignment, thread message, compatibility comment action을 다루는 namespace입니다.
   */
  readonly tasks: RiidoAIAgentTasksReactNamespace;
}

/**
 * workspace-scoped agent 권한을 통해 daemon 상세와 제어 command를 다루는 v2 namespace입니다.
 */
export interface RiidoV2AIAgentAgentsDaemonReactNamespace {
  /**
   * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 상세를 조회합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.agents.daemon.details`
   * 검색용 generated 경로: `aiAgent.agents.daemon.details`
   * 접근 예시: `riido.v2.aiAgent.agents.daemon.details`
   * cache tag: `v2.aiAgent.agents.daemon`
   */
  readonly details: GetAIAgentDaemonV2ReactEndpoint;
  /**
   * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 재시작을 요청합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.agents.daemon.restart`
   * 검색용 generated 경로: `aiAgent.agents.daemon.restart`
   * 접근 예시: `riido.v2.aiAgent.agents.daemon.restart`
   * invalidates: `v2.aiAgent.agents.daemon`, `v2.aiAgent.devices.runtimes`
   */
  readonly restart: RestartAIAgentDaemonV2ReactEndpoint;
  /**
   * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 시작을 요청합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.agents.daemon.start`
   * 검색용 generated 경로: `aiAgent.agents.daemon.start`
   * 접근 예시: `riido.v2.aiAgent.agents.daemon.start`
   * invalidates: `v2.aiAgent.agents.daemon`, `v2.aiAgent.devices.runtimes`
   */
  readonly start: StartAIAgentDaemonV2ReactEndpoint;
  /**
   * runtime 설정 화면에서 agent 권한으로 접근 가능한 daemon 중지를 요청합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.agents.daemon.stop`
   * 검색용 generated 경로: `aiAgent.agents.daemon.stop`
   * 접근 예시: `riido.v2.aiAgent.agents.daemon.stop`
   * invalidates: `v2.aiAgent.agents.daemon`, `v2.aiAgent.devices.runtimes`
   */
  readonly stop: StopAIAgentDaemonV2ReactEndpoint;
}

/**
 * workspace 안에 생성되는 agent 설정과 mutation을 다루는 v2 namespace입니다.
 */
export interface RiidoV2AIAgentAgentsReactNamespace {
  /**
   * Figma agent 추가 화면에서 AI agent 설정을 생성합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.agents.create`
   * 검색용 generated 경로: `aiAgent.agents.create`
   * 접근 예시: `riido.v2.aiAgent.agents.create`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.devices.runtimes`, `v2.aiAgent.tasks.assignableAgents`
   */
  readonly create: CreateAIAgentV2ReactEndpoint;
  /**
   * workspace-scoped agent 권한을 통해 daemon 상세와 제어 command를 다루는 v2 namespace입니다.
   */
  readonly daemon: RiidoV2AIAgentAgentsDaemonReactNamespace;
  /**
   * agent를 삭제하고 queued/running assignment를 강제로 정리합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.agents.delete`
   * 검색용 generated 경로: `aiAgent.agents.delete`
   * 접근 예시: `riido.v2.aiAgent.agents.delete`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.devices.runtimes`, `v2.aiAgent.agents.editability`, `v2.aiAgent.tasks.assignableAgents`, `v2.aiAgent.tasks.threads`, `v2.aiAgent.tasks.assignedAgentProfiles`
   */
  readonly delete: DeleteAIAgentV2ReactEndpoint;
  /**
   * client control 활성화 전에 agent 수정 가능 여부를 조회합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.agents.editability`
   * 검색용 generated 경로: `aiAgent.agents.editability`
   * 접근 예시: `riido.v2.aiAgent.agents.editability`
   * cache tag: `v2.aiAgent.agents.editability`
   */
  readonly editability: GetAIAgentEditabilityV2ReactEndpoint;
  /**
   * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.agents.updateConfiguration`
   * 검색용 generated 경로: `aiAgent.agents.updateConfiguration`
   * 접근 예시: `riido.v2.aiAgent.agents.updateConfiguration`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.agents.editability`, `v2.aiAgent.tasks.assignableAgents`
   */
  readonly updateConfiguration: UpdateAIAgentConfigurationV2ReactEndpoint;
}

/**
 * account-owned device/runtime을 선택된 workspace agent 권한에 맞춰 읽는 v2 namespace입니다.
 */
export interface RiidoV2AIAgentDevicesReactNamespace {
  /**
   * 권한이 확인된 principal의 device runtime 상태를 조회합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.devices.runtimes`
   * 검색용 generated 경로: `aiAgent.devices.runtimes`
   * 접근 예시: `riido.v2.aiAgent.devices.runtimes`
   * cache tag: `v2.aiAgent.devices.runtimes`
   */
  readonly runtimes: ListAIAgentDeviceRuntimesV2ReactEndpoint;
}

/**
 * 선택된 workspace 범위로 client가 수신하는 SSE stream namespace입니다.
 */
export interface RiidoV2AIAgentEventsReactNamespace {
  /**
   * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.events.stream`
   * 검색용 generated 경로: `aiAgent.events.stream`
   * 접근 예시: `riido.v2.aiAgent.events.stream`
   * cache tag: `v2.aiAgent.events.stream`
   */
  readonly stream: StreamAIAgentClientEventsV2ReactEndpoint;
}

/**
 * 서버 제공 fixture 목록과 fixture 기반 agent 생성 진입점입니다. 템플릿 엔티티를 만들거나 관리하지 않습니다.
 */
export interface RiidoV2AIAgentOnboardingFixturesReactNamespace extends ListAIAgentOnboardingFixturesV2ReactEndpoint {
  /**
   * 선택한 onboarding fixture를 기준으로 일반 AI agent를 생성합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.onboarding.fixtures.createAgent`
   * 검색용 generated 경로: `aiAgent.onboarding.fixtures.createAgent`
   * 접근 예시: `riido.v2.aiAgent.onboarding.fixtures.createAgent`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.devices.runtimes`, `v2.aiAgent.tasks.assignableAgents`
   */
  readonly createAgent: CreateAIAgentFromOnboardingFixtureV2ReactEndpoint;
}

/**
 * 선택된 workspace에서 빠른 agent 생성을 돕는 onboarding fixture namespace입니다.
 */
export interface RiidoV2AIAgentOnboardingReactNamespace {
  /**
   * 서버 제공 fixture 목록과 fixture 기반 agent 생성 진입점입니다. 템플릿 엔티티를 만들거나 관리하지 않습니다.
   */
  readonly fixtures: RiidoV2AIAgentOnboardingFixturesReactNamespace;
}

/**
 * 선택된 workspace의 task thread에 사용자가 다음 작업 지시를 남기는 정식 message command namespace입니다.
 */
export interface RiidoV2AIAgentTasksThreadMessagesReactNamespace {
  /**
   * task thread message로 AI agent에게 다음 작업 지시를 전달합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.threadMessages.create`
   * 검색용 generated 경로: `aiAgent.tasks.threadMessages.create`
   * 접근 예시: `riido.v2.aiAgent.tasks.threadMessages.create`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.tasks.assignableAgents`, `v2.aiAgent.tasks.threads`
   */
  readonly create: CreateAIAgentTaskThreadMessageV2ReactEndpoint;
}

/**
 * 선택된 workspace의 task thread에서 AI Agent assignment, thread message, compatibility comment action을 다루는 namespace입니다.
 */
export interface RiidoV2AIAgentTasksReactNamespace {
  /**
   * task participant dropdown에서 AI agent를 배정합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.assign`
   * 검색용 generated 경로: `aiAgent.tasks.assign`
   * 접근 예시: `riido.v2.aiAgent.tasks.assign`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.tasks.assignableAgents`, `v2.aiAgent.tasks.threads`, `v2.aiAgent.tasks.assignedAgentProfiles`
   */
  readonly assign: AssignAIAgentTaskV2ReactEndpoint;
  /**
   * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.assignableAgents`
   * 검색용 generated 경로: `aiAgent.tasks.assignableAgents`
   * 접근 예시: `riido.v2.aiAgent.tasks.assignableAgents`
   * cache tag: `v2.aiAgent.tasks.assignableAgents`
   */
  readonly assignableAgents: ListAIAgentTaskAssignableAgentsV2ReactEndpoint;
  /**
   * workspace task/card 목록에서 현재 배정된 AI Agent profile 표시값을 component_id key 해시맵으로 조회합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.assignedAgentProfiles`
   * 검색용 generated 경로: `aiAgent.tasks.assignedAgentProfiles`
   * 접근 예시: `riido.v2.aiAgent.tasks.assignedAgentProfiles`
   * cache tag: `v2.aiAgent.tasks.assignedAgentProfiles`
   */
  readonly assignedAgentProfiles: ListWorkspaceAssignedAgentProfilesV2ReactEndpoint;
  /**
   * task thread의 stop action으로 AI agent 작업을 중단합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.stop`
   * 검색용 generated 경로: `aiAgent.tasks.stop`
   * 접근 예시: `riido.v2.aiAgent.tasks.stop`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.tasks.assignableAgents`, `v2.aiAgent.tasks.threads`
   */
  readonly stop: StopAIAgentTaskV2ReactEndpoint;
  /**
   * 호환 task comment route로 AI agent에게 메시지를 전달합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.submitComment`
   * 검색용 generated 경로: `aiAgent.tasks.submitComment`
   * 접근 예시: `riido.v2.aiAgent.tasks.submitComment`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.tasks.assignableAgents`, `v2.aiAgent.tasks.threads`, `v2.aiAgent.tasks.assignedAgentProfiles`
   */
  readonly submitComment: SubmitAIAgentTaskCommentV2ReactEndpoint;
  /**
   * 선택된 workspace의 task thread에 사용자가 다음 작업 지시를 남기는 정식 message command namespace입니다.
   */
  readonly threadMessages: RiidoV2AIAgentTasksThreadMessagesReactNamespace;
  /**
   * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.threads`
   * 검색용 generated 경로: `aiAgent.tasks.threads`
   * 접근 예시: `riido.v2.aiAgent.tasks.threads`
   * cache tag: `v2.aiAgent.tasks.threads`
   */
  readonly threads: ListAIAgentTaskThreadsV2ReactEndpoint;
  /**
   * task participant에서 AI agent를 제거하고 작업을 중단합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.tasks.unassign`
   * 검색용 generated 경로: `aiAgent.tasks.unassign`
   * 접근 예시: `riido.v2.aiAgent.tasks.unassign`
   * invalidates: `v2.aiAgent.bootstrap`, `v2.aiAgent.tasks.assignableAgents`, `v2.aiAgent.tasks.threads`, `v2.aiAgent.tasks.assignedAgentProfiles`
   */
  readonly unassign: UnassignAIAgentTaskV2ReactEndpoint;
}

/**
 * workspace_id path parameter로 범위가 정해지는 AI Agent v2 namespace입니다.
 */
export interface RiidoV2AIAgentReactNamespace {
  /**
   * workspace 안에 생성되는 agent 설정과 mutation을 다루는 v2 namespace입니다.
   */
  readonly agents: RiidoV2AIAgentAgentsReactNamespace;
  /**
   * web 또는 desktop webview client의 AI Agent 설정/온보딩 초기 데이터와 v2.aiAgent.bootstrap.agents[] agent list를 조회합니다 (v2 workspace-scoped)
   * 계약 generated path: `v2.aiAgent.bootstrap`
   * 검색용 generated 경로: `aiAgent.bootstrap`
   * 접근 예시: `riido.v2.aiAgent.bootstrap`
   * cache tag: `v2.aiAgent.bootstrap`
   */
  readonly bootstrap: GetAIAgentClientBootstrapV2ReactEndpoint;
  /**
   * account-owned device/runtime을 선택된 workspace agent 권한에 맞춰 읽는 v2 namespace입니다.
   */
  readonly devices: RiidoV2AIAgentDevicesReactNamespace;
  /**
   * 선택된 workspace 범위로 client가 수신하는 SSE stream namespace입니다.
   */
  readonly events: RiidoV2AIAgentEventsReactNamespace;
  /**
   * 선택된 workspace에서 빠른 agent 생성을 돕는 onboarding fixture namespace입니다.
   */
  readonly onboarding: RiidoV2AIAgentOnboardingReactNamespace;
  /**
   * 선택된 workspace의 task thread에서 AI Agent assignment, thread message, compatibility comment action을 다루는 namespace입니다.
   */
  readonly tasks: RiidoV2AIAgentTasksReactNamespace;
}

/**
 * v2 client API module입니다. workspace-scoped AI Agent API를 riido.v2.aiAgent.* 경로로 제공합니다. v1은 UI 테스트 호환 표면으로 유지됩니다.
 */
export interface RiidoV2ReactModule {
  /**
   * workspace_id path parameter로 범위가 정해지는 AI Agent v2 namespace입니다.
   */
  readonly aiAgent: RiidoV2AIAgentReactNamespace;
}

/**
 * control-plane API를 DSL client metadata의 module/namespace별로 묶은 config-bound facade입니다.
 */
export interface RiidoControlPlaneReactClient {
  /**
   * AI Agent client API module입니다. 화면 구성을 소유하지 않고, client가 화면을 구성할 때 필요한 API 의미, 정책, lifecycle, cache 관계를 제공합니다.
   */
  readonly aiAgent: RiidoAIAgentReactModule;
  /**
   * v2 client API module입니다. workspace-scoped AI Agent API를 riido.v2.aiAgent.* 경로로 제공합니다. v1은 UI 테스트 호환 표면으로 유지됩니다.
   */
  readonly v2: RiidoV2ReactModule;
}

/**
 * control-plane API facade에 React Query hook을 얹은 client 전용 wrapper입니다.
 * hook은 반드시 `@/lib/react-query`를 통과하므로 riido-client의 workspace/demo 정책을 우회하지 않습니다.
 */
export function useRiidoControlPlaneClient(config: core.RiidoClientConfig): RiidoControlPlaneReactClient {
  const coreClient = useMemo(() => core.createRiidoControlPlaneClient(config), [config.baseUrl, config.fetcher, config.aiAgentToken]);

  return useMemo(
    () => ({
      aiAgent: {
        agents: {
          create: {
            ...coreClient.aiAgent.agents.create,
            useMutation: (options: core.RiidoMutationOptions<core.AgentClientRecordResponse, core.CreateAIAgentMutationVariables> = {}) => useMutation<core.AgentClientRecordResponse, Error, core.CreateAIAgentMutationVariables>(coreClient.aiAgent.agents.create.mutation(options)),
          },
          daemon: {
            details: {
              ...coreClient.aiAgent.agents.daemon.details,
              useQuery: (params: core.GetAIAgentDaemonPathParams, options?: core.RiidoQueryOptions<core.DeviceDaemonDetailResponse>) => useQuery<core.DeviceDaemonDetailResponse, Error>(coreClient.aiAgent.agents.daemon.details.query(params, options)),
            },
            restart: {
              ...coreClient.aiAgent.agents.daemon.restart,
              useMutation: (options: core.RiidoMutationOptions<core.DeviceDaemonCommandResponse, core.RestartAIAgentDaemonMutationVariables> = {}) => useMutation<core.DeviceDaemonCommandResponse, Error, core.RestartAIAgentDaemonMutationVariables>(coreClient.aiAgent.agents.daemon.restart.mutation(options)),
            },
            start: {
              ...coreClient.aiAgent.agents.daemon.start,
              useMutation: (options: core.RiidoMutationOptions<core.DeviceDaemonCommandResponse, core.StartAIAgentDaemonMutationVariables> = {}) => useMutation<core.DeviceDaemonCommandResponse, Error, core.StartAIAgentDaemonMutationVariables>(coreClient.aiAgent.agents.daemon.start.mutation(options)),
            },
            stop: {
              ...coreClient.aiAgent.agents.daemon.stop,
              useMutation: (options: core.RiidoMutationOptions<core.DeviceDaemonCommandResponse, core.StopAIAgentDaemonMutationVariables> = {}) => useMutation<core.DeviceDaemonCommandResponse, Error, core.StopAIAgentDaemonMutationVariables>(coreClient.aiAgent.agents.daemon.stop.mutation(options)),
            },
          },
          delete: {
            ...coreClient.aiAgent.agents.delete,
            useMutation: (options: core.RiidoMutationOptions<core.DeleteAgentResponse, core.DeleteAIAgentMutationVariables> = {}) => useMutation<core.DeleteAgentResponse, Error, core.DeleteAIAgentMutationVariables>(coreClient.aiAgent.agents.delete.mutation(options)),
          },
          editability: {
            ...coreClient.aiAgent.agents.editability,
            useQuery: (params: core.GetAIAgentEditabilityPathParams, options?: core.RiidoQueryOptions<core.AgentEditabilityResponse>) => useQuery<core.AgentEditabilityResponse, Error>(coreClient.aiAgent.agents.editability.query(params, options)),
          },
          updateConfiguration: {
            ...coreClient.aiAgent.agents.updateConfiguration,
            useMutation: (options: core.RiidoMutationOptions<core.AgentClientRecordResponse, core.UpdateAIAgentConfigurationMutationVariables> = {}) => useMutation<core.AgentClientRecordResponse, Error, core.UpdateAIAgentConfigurationMutationVariables>(coreClient.aiAgent.agents.updateConfiguration.mutation(options)),
          },
        },
        bootstrap: {
          ...coreClient.aiAgent.bootstrap,
          useQuery: (options?: core.RiidoQueryOptions<core.ClientBootstrapResponse>) => useQuery<core.ClientBootstrapResponse, Error>(coreClient.aiAgent.bootstrap.query(options)),
        },
        devices: {
          runtimes: {
            ...coreClient.aiAgent.devices.runtimes,
            useQuery: (options?: core.RiidoQueryOptions<core.DeviceRuntimeListResponse>) => useQuery<core.DeviceRuntimeListResponse, Error>(coreClient.aiAgent.devices.runtimes.query(options)),
          },
        },
        events: {
          stream: {
            ...coreClient.aiAgent.events.stream,
            useQuery: (options?: core.RiidoQueryOptions<Response>) => useQuery<Response, Error>(coreClient.aiAgent.events.stream.query(options)),
          },
        },
        onboarding: {
          fixtures: {
            ...coreClient.aiAgent.onboarding.fixtures,
            useQuery: (options?: core.RiidoQueryOptions<core.AgentOnboardingFixtureListResponse>) => useQuery<core.AgentOnboardingFixtureListResponse, Error>(coreClient.aiAgent.onboarding.fixtures.query(options)),
            createAgent: {
              ...coreClient.aiAgent.onboarding.fixtures.createAgent,
              useMutation: (options: core.RiidoMutationOptions<core.AgentClientRecordResponse, core.CreateAIAgentFromOnboardingFixtureMutationVariables> = {}) => useMutation<core.AgentClientRecordResponse, Error, core.CreateAIAgentFromOnboardingFixtureMutationVariables>(coreClient.aiAgent.onboarding.fixtures.createAgent.mutation(options)),
            },
          },
        },
        tasks: {
          assign: {
            ...coreClient.aiAgent.tasks.assign,
            useMutation: (options: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.AssignAIAgentTaskMutationVariables> = {}) => useMutation<core.AIAgentTaskActionResponse, Error, core.AssignAIAgentTaskMutationVariables>(coreClient.aiAgent.tasks.assign.mutation(options)),
          },
          assignableAgents: {
            ...coreClient.aiAgent.tasks.assignableAgents,
            useQuery: (params: core.ListAIAgentTaskAssignableAgentsPathParams, options?: core.RiidoQueryOptions<core.AgentClientListResponse>) => useQuery<core.AgentClientListResponse, Error>(coreClient.aiAgent.tasks.assignableAgents.query(params, options)),
          },
          stop: {
            ...coreClient.aiAgent.tasks.stop,
            useMutation: (options: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.StopAIAgentTaskMutationVariables> = {}) => useMutation<core.AIAgentTaskActionResponse, Error, core.StopAIAgentTaskMutationVariables>(coreClient.aiAgent.tasks.stop.mutation(options)),
          },
          submitComment: {
            ...coreClient.aiAgent.tasks.submitComment,
            useMutation: (options: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.SubmitAIAgentTaskCommentMutationVariables> = {}) => useMutation<core.AIAgentTaskActionResponse, Error, core.SubmitAIAgentTaskCommentMutationVariables>(coreClient.aiAgent.tasks.submitComment.mutation(options)),
          },
          threadMessages: {
            create: {
              ...coreClient.aiAgent.tasks.threadMessages.create,
              useMutation: (options: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.CreateAIAgentTaskThreadMessageMutationVariables> = {}) => useMutation<core.AIAgentTaskActionResponse, Error, core.CreateAIAgentTaskThreadMessageMutationVariables>(coreClient.aiAgent.tasks.threadMessages.create.mutation(options)),
            },
          },
          threads: {
            ...coreClient.aiAgent.tasks.threads,
            useQuery: (params: core.ListAIAgentTaskThreadsPathParams, options?: core.RiidoQueryOptions<core.AIAgentTaskThreadCollectionResponse>) => useQuery<core.AIAgentTaskThreadCollectionResponse, Error>(coreClient.aiAgent.tasks.threads.query(params, options)),
          },
          unassign: {
            ...coreClient.aiAgent.tasks.unassign,
            useMutation: (options: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.UnassignAIAgentTaskMutationVariables> = {}) => useMutation<core.AIAgentTaskActionResponse, Error, core.UnassignAIAgentTaskMutationVariables>(coreClient.aiAgent.tasks.unassign.mutation(options)),
          },
        },
      },
      v2: {
        aiAgent: {
          agents: {
            create: {
              ...coreClient.v2.aiAgent.agents.create,
              useMutation: (options: core.RiidoMutationOptions<core.AgentClientRecordResponseV2, core.CreateAIAgentV2MutationVariables> = {}) => useMutation<core.AgentClientRecordResponseV2, Error, core.CreateAIAgentV2MutationVariables>(coreClient.v2.aiAgent.agents.create.mutation(options)),
            },
            daemon: {
              details: {
                ...coreClient.v2.aiAgent.agents.daemon.details,
                useQuery: (params: core.GetAIAgentDaemonV2PathParams, options?: core.RiidoQueryOptions<core.DeviceDaemonDetailResponse>) => useQuery<core.DeviceDaemonDetailResponse, Error>(coreClient.v2.aiAgent.agents.daemon.details.query(params, options)),
              },
              restart: {
                ...coreClient.v2.aiAgent.agents.daemon.restart,
                useMutation: (options: core.RiidoMutationOptions<core.DeviceDaemonCommandResponse, core.RestartAIAgentDaemonV2MutationVariables> = {}) => useMutation<core.DeviceDaemonCommandResponse, Error, core.RestartAIAgentDaemonV2MutationVariables>(coreClient.v2.aiAgent.agents.daemon.restart.mutation(options)),
              },
              start: {
                ...coreClient.v2.aiAgent.agents.daemon.start,
                useMutation: (options: core.RiidoMutationOptions<core.DeviceDaemonCommandResponse, core.StartAIAgentDaemonV2MutationVariables> = {}) => useMutation<core.DeviceDaemonCommandResponse, Error, core.StartAIAgentDaemonV2MutationVariables>(coreClient.v2.aiAgent.agents.daemon.start.mutation(options)),
              },
              stop: {
                ...coreClient.v2.aiAgent.agents.daemon.stop,
                useMutation: (options: core.RiidoMutationOptions<core.DeviceDaemonCommandResponse, core.StopAIAgentDaemonV2MutationVariables> = {}) => useMutation<core.DeviceDaemonCommandResponse, Error, core.StopAIAgentDaemonV2MutationVariables>(coreClient.v2.aiAgent.agents.daemon.stop.mutation(options)),
              },
            },
            delete: {
              ...coreClient.v2.aiAgent.agents.delete,
              useMutation: (options: core.RiidoMutationOptions<core.DeleteAgentResponse, core.DeleteAIAgentV2MutationVariables> = {}) => useMutation<core.DeleteAgentResponse, Error, core.DeleteAIAgentV2MutationVariables>(coreClient.v2.aiAgent.agents.delete.mutation(options)),
            },
            editability: {
              ...coreClient.v2.aiAgent.agents.editability,
              useQuery: (params: core.GetAIAgentEditabilityV2PathParams, options?: core.RiidoQueryOptions<core.AgentEditabilityResponse>) => useQuery<core.AgentEditabilityResponse, Error>(coreClient.v2.aiAgent.agents.editability.query(params, options)),
            },
            updateConfiguration: {
              ...coreClient.v2.aiAgent.agents.updateConfiguration,
              useMutation: (options: core.RiidoMutationOptions<core.AgentClientRecordResponseV2, core.UpdateAIAgentConfigurationV2MutationVariables> = {}) => useMutation<core.AgentClientRecordResponseV2, Error, core.UpdateAIAgentConfigurationV2MutationVariables>(coreClient.v2.aiAgent.agents.updateConfiguration.mutation(options)),
            },
          },
          bootstrap: {
            ...coreClient.v2.aiAgent.bootstrap,
            useQuery: (params: core.GetAIAgentClientBootstrapV2PathParams, options?: core.RiidoQueryOptions<core.ClientBootstrapResponseV2>) => useQuery<core.ClientBootstrapResponseV2, Error>(coreClient.v2.aiAgent.bootstrap.query(params, options)),
          },
          devices: {
            runtimes: {
              ...coreClient.v2.aiAgent.devices.runtimes,
              useQuery: (params: core.ListAIAgentDeviceRuntimesV2PathParams, options?: core.RiidoQueryOptions<core.DeviceRuntimeListResponse>) => useQuery<core.DeviceRuntimeListResponse, Error>(coreClient.v2.aiAgent.devices.runtimes.query(params, options)),
            },
          },
          events: {
            stream: {
              ...coreClient.v2.aiAgent.events.stream,
              useQuery: (params: core.StreamAIAgentClientEventsV2PathParams, options?: core.RiidoQueryOptions<Response>) => useQuery<Response, Error>(coreClient.v2.aiAgent.events.stream.query(params, options)),
            },
          },
          onboarding: {
            fixtures: {
              ...coreClient.v2.aiAgent.onboarding.fixtures,
              useQuery: (params: core.ListAIAgentOnboardingFixturesV2PathParams, options?: core.RiidoQueryOptions<core.AgentOnboardingFixtureListResponse>) => useQuery<core.AgentOnboardingFixtureListResponse, Error>(coreClient.v2.aiAgent.onboarding.fixtures.query(params, options)),
              createAgent: {
                ...coreClient.v2.aiAgent.onboarding.fixtures.createAgent,
                useMutation: (options: core.RiidoMutationOptions<core.AgentClientRecordResponseV2, core.CreateAIAgentFromOnboardingFixtureV2MutationVariables> = {}) => useMutation<core.AgentClientRecordResponseV2, Error, core.CreateAIAgentFromOnboardingFixtureV2MutationVariables>(coreClient.v2.aiAgent.onboarding.fixtures.createAgent.mutation(options)),
              },
            },
          },
          tasks: {
            assign: {
              ...coreClient.v2.aiAgent.tasks.assign,
              useMutation: (options: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.AssignAIAgentTaskV2MutationVariables> = {}) => useMutation<core.AIAgentTaskActionResponse, Error, core.AssignAIAgentTaskV2MutationVariables>(coreClient.v2.aiAgent.tasks.assign.mutation(options)),
            },
            assignableAgents: {
              ...coreClient.v2.aiAgent.tasks.assignableAgents,
              useQuery: (params: core.ListAIAgentTaskAssignableAgentsV2PathParams, options?: core.RiidoQueryOptions<core.AgentClientListResponseV2>) => useQuery<core.AgentClientListResponseV2, Error>(coreClient.v2.aiAgent.tasks.assignableAgents.query(params, options)),
            },
            assignedAgentProfiles: {
              ...coreClient.v2.aiAgent.tasks.assignedAgentProfiles,
              useQuery: (params: core.ListWorkspaceAssignedAgentProfilesV2PathParams, options?: core.RiidoQueryOptions<core.AssignedAgentProfileMapResponse>) => useQuery<core.AssignedAgentProfileMapResponse, Error>(coreClient.v2.aiAgent.tasks.assignedAgentProfiles.query(params, options)),
            },
            stop: {
              ...coreClient.v2.aiAgent.tasks.stop,
              useMutation: (options: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.StopAIAgentTaskV2MutationVariables> = {}) => useMutation<core.AIAgentTaskActionResponse, Error, core.StopAIAgentTaskV2MutationVariables>(coreClient.v2.aiAgent.tasks.stop.mutation(options)),
            },
            submitComment: {
              ...coreClient.v2.aiAgent.tasks.submitComment,
              useMutation: (options: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.SubmitAIAgentTaskCommentV2MutationVariables> = {}) => useMutation<core.AIAgentTaskActionResponse, Error, core.SubmitAIAgentTaskCommentV2MutationVariables>(coreClient.v2.aiAgent.tasks.submitComment.mutation(options)),
            },
            threadMessages: {
              create: {
                ...coreClient.v2.aiAgent.tasks.threadMessages.create,
                useMutation: (options: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.CreateAIAgentTaskThreadMessageV2MutationVariables> = {}) => useMutation<core.AIAgentTaskActionResponse, Error, core.CreateAIAgentTaskThreadMessageV2MutationVariables>(coreClient.v2.aiAgent.tasks.threadMessages.create.mutation(options)),
              },
            },
            threads: {
              ...coreClient.v2.aiAgent.tasks.threads,
              useQuery: (params: core.ListAIAgentTaskThreadsV2PathParams, options?: core.RiidoQueryOptions<core.AIAgentTaskThreadCollectionResponse>) => useQuery<core.AIAgentTaskThreadCollectionResponse, Error>(coreClient.v2.aiAgent.tasks.threads.query(params, options)),
            },
            unassign: {
              ...coreClient.v2.aiAgent.tasks.unassign,
              useMutation: (options: core.RiidoMutationOptions<core.AIAgentTaskActionResponse, core.UnassignAIAgentTaskV2MutationVariables> = {}) => useMutation<core.AIAgentTaskActionResponse, Error, core.UnassignAIAgentTaskV2MutationVariables>(coreClient.v2.aiAgent.tasks.unassign.mutation(options)),
            },
          },
        },
      },
    }),
    [coreClient],
  );
}
