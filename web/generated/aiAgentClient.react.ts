'use client';

// 이 파일은 tools/reactquerygen으로 생성됩니다. 직접 수정하지 마세요.
// 원본: contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json

/* eslint-disable react-hooks/rules-of-hooks */

import { useMemo } from 'react';
import { useMutation, useQuery, type UseMutationResult, type UseQueryResult } from '@/lib/react-query';
import * as core from './aiAgentClient';

/**
 * Figma agent 추가 화면에서 AI agent 설정을 생성합니다
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
 * client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.
 */
export interface UpdateAIAgentConfigurationReactEndpoint extends core.UpdateAIAgentConfigurationEndpoint {
  /**
   * React Query useMutation hook입니다.
   */
  readonly useMutation: (options?: core.RiidoMutationOptions<core.AgentClientRecordResponse, core.UpdateAIAgentConfigurationMutationVariables>) => UseMutationResult<core.AgentClientRecordResponse, Error, core.UpdateAIAgentConfigurationMutationVariables>;
}

/**
 * client control 활성화 전에 agent 수정 가능 여부를 조회합니다
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface GetAIAgentEditabilityReactEndpoint extends core.GetAIAgentEditabilityEndpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (params: core.GetAIAgentEditabilityPathParams, options?: core.RiidoQueryOptions<core.AgentEditabilityResponse>) => UseQueryResult<core.AgentEditabilityResponse, Error>;
}

/**
 * web 또는 desktop webview client의 AI Agent 화면 초기 데이터를 조회합니다
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
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface StreamAIAgentClientEventsReactEndpoint extends core.StreamAIAgentClientEventsEndpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (options?: core.RiidoQueryOptions<Response>) => UseQueryResult<Response, Error>;
}

/**
 * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface ListAIAgentTaskAssignableAgentsReactEndpoint extends core.ListAIAgentTaskAssignableAgentsEndpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (params: core.ListAIAgentTaskAssignableAgentsPathParams, options?: core.RiidoQueryOptions<core.AgentClientListResponse>) => UseQueryResult<core.AgentClientListResponse, Error>;
}

/**
 * task thread comment를 할당된 AI agent에게 전달합니다
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
 * client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.
 */
export interface ListAIAgentTaskThreadsReactEndpoint extends core.ListAIAgentTaskThreadsEndpoint {
  /**
   * React Query useQuery hook입니다.
   */
  readonly useQuery: (params: core.ListAIAgentTaskThreadsPathParams, options?: core.RiidoQueryOptions<core.AIAgentTaskThreadCollectionResponse>) => UseQueryResult<core.AIAgentTaskThreadCollectionResponse, Error>;
}

/**
 * agent 설정과 mutation을 다루는 namespace입니다. agent 이름은 중복될 수 있고 assigned task가 있으면 수정할 수 없습니다.
 */
export interface RiidoAIAgentAgentsReactNamespace {
  /**
   * Figma agent 추가 화면에서 AI agent 설정을 생성합니다 invalidates: `aiAgent.bootstrap`, `aiAgent.devices.runtimes`, `aiAgent.tasks.assignableAgents`
   */
  readonly create: CreateAIAgentReactEndpoint;
  /**
   * agent를 삭제하고 queued/running assignment를 강제로 정리합니다 invalidates: `aiAgent.bootstrap`, `aiAgent.devices.runtimes`, `aiAgent.agents.editability`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly delete: DeleteAIAgentReactEndpoint;
  /**
   * client control 활성화 전에 agent 수정 가능 여부를 조회합니다 cache tag: `aiAgent.agents.editability`
   */
  readonly editability: GetAIAgentEditabilityReactEndpoint;
  /**
   * 할당된 task가 없을 때 agent 표시/설정 필드를 수정합니다 invalidates: `aiAgent.bootstrap`, `aiAgent.agents.editability`, `aiAgent.tasks.assignableAgents`
   */
  readonly updateConfiguration: UpdateAIAgentConfigurationReactEndpoint;
}

/**
 * device와 runtime 상태를 다루는 namespace입니다.
 */
export interface RiidoAIAgentDevicesReactNamespace {
  /**
   * 권한이 확인된 principal의 device runtime 상태를 조회합니다 cache tag: `aiAgent.devices.runtimes`
   */
  readonly runtimes: ListAIAgentDeviceRuntimesReactEndpoint;
}

/**
 * client가 SSE로 수신하는 상태 변경 stream namespace입니다.
 */
export interface RiidoAIAgentEventsReactNamespace {
  /**
   * editability, work status, runtime snapshot, task-thread progress에 대한 AI Agent client update를 스트리밍합니다 cache tag: `aiAgent.events.stream`
   */
  readonly stream: StreamAIAgentClientEventsReactEndpoint;
}

/**
 * task thread에서 AI Agent assignment와 comment action을 다루는 namespace입니다.
 */
export interface RiidoAIAgentTasksReactNamespace {
  /**
   * task participant dropdown에서 할당 가능한 agent 목록을 조회합니다 cache tag: `aiAgent.tasks.assignableAgents`
   */
  readonly assignableAgents: ListAIAgentTaskAssignableAgentsReactEndpoint;
  /**
   * task thread의 stop action으로 AI agent 작업을 중단합니다 invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly stop: StopAIAgentTaskReactEndpoint;
  /**
   * task thread comment를 할당된 AI agent에게 전달합니다 invalidates: `aiAgent.bootstrap`, `aiAgent.tasks.assignableAgents`, `aiAgent.tasks.threads`
   */
  readonly submitComment: SubmitAIAgentTaskCommentReactEndpoint;
  /**
   * active stream link가 있을 때만 이어서 연결할 수 있도록 AI Agent task thread 목록을 조회합니다 cache tag: `aiAgent.tasks.threads`
   */
  readonly threads: ListAIAgentTaskThreadsReactEndpoint;
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
   * web 또는 desktop webview client의 AI Agent 화면 초기 데이터를 조회합니다 cache tag: `aiAgent.bootstrap`
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
   * task thread에서 AI Agent assignment와 comment action을 다루는 namespace입니다.
   */
  readonly tasks: RiidoAIAgentTasksReactNamespace;
}

/**
 * control-plane API를 DSL client metadata의 module/namespace별로 묶은 config-bound facade입니다.
 */
export interface RiidoControlPlaneReactClient {
  /**
   * AI Agent client API module입니다. 화면 구성을 소유하지 않고, client가 화면을 구성할 때 필요한 API 의미, 정책, lifecycle, cache 관계를 제공합니다.
   */
  readonly aiAgent: RiidoAIAgentReactModule;
}

/**
 * control-plane API facade에 React Query hook을 얹은 client 전용 wrapper입니다.
 * hook은 반드시 `@/lib/react-query`를 통과하므로 riido-client의 workspace/demo 정책을 우회하지 않습니다.
 */
export function useRiidoControlPlaneClient(config: core.RiidoClientConfig): RiidoControlPlaneReactClient {
  const coreClient = useMemo(() => core.createRiidoControlPlaneClient(config), [config.baseUrl, config.fetcher, config.token]);

  return useMemo(
    () => ({
      aiAgent: {
        agents: {
          create: {
            ...coreClient.aiAgent.agents.create,
            useMutation: (options: core.RiidoMutationOptions<core.AgentClientRecordResponse, core.CreateAIAgentMutationVariables> = {}) => useMutation<core.AgentClientRecordResponse, Error, core.CreateAIAgentMutationVariables>(coreClient.aiAgent.agents.create.mutation(options)),
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
        tasks: {
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
          threads: {
            ...coreClient.aiAgent.tasks.threads,
            useQuery: (params: core.ListAIAgentTaskThreadsPathParams, options?: core.RiidoQueryOptions<core.AIAgentTaskThreadCollectionResponse>) => useQuery<core.AIAgentTaskThreadCollectionResponse, Error>(coreClient.aiAgent.tasks.threads.query(params, options)),
          },
        },
      },
    }),
    [coreClient],
  );
}
