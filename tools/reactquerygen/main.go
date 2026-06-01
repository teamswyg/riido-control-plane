package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type openAPISpec struct {
	Paths         map[string]map[string]operation `json:"paths"`
	ClientModules []clientModuleMetadata          `json:"x-riido-client-modules"`
	Components    struct {
		Schemas map[string]schema `json:"schemas"`
	} `json:"components"`
}

type clientModuleMetadata struct {
	Module      string                    `json:"module"`
	Description string                    `json:"description"`
	Namespaces  []clientNamespaceMetadata `json:"namespaces"`
}

type clientNamespaceMetadata struct {
	Path        []string `json:"path"`
	Description string   `json:"description"`
}

type operation struct {
	OperationID string              `json:"operationId"`
	Summary     string              `json:"summary"`
	Parameters  []parameter         `json:"parameters"`
	RequestBody *requestBody        `json:"requestBody"`
	Responses   map[string]response `json:"responses"`
	Client      clientMetadata      `json:"x-riido-client"`
}

type clientMetadata struct {
	Module        string   `json:"module"`
	FacadePath    []string `json:"facade_path"`
	GeneratedPath string   `json:"generated_path"`
	CacheTag      string   `json:"cache_tag"`
	Invalidates   []string `json:"invalidates"`
}

type parameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Required bool   `json:"required"`
	Schema   schema `json:"schema"`
}

type requestBody struct {
	Required bool                  `json:"required"`
	Content  map[string]mediaValue `json:"content"`
}

type response struct {
	Content map[string]mediaValue `json:"content"`
}

type mediaValue struct {
	Schema schema `json:"schema"`
}

type schema struct {
	Ref         string            `json:"$ref"`
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Format      string            `json:"format"`
	Enum        []string          `json:"enum"`
	Required    []string          `json:"required"`
	Properties  map[string]schema `json:"properties"`
	Items       *schema           `json:"items"`
	OneOf       []schema          `json:"oneOf"`
}

type routeOperation struct {
	Method string
	Path   string
	Op     operation
}

type operationInfo struct {
	Route             routeOperation
	Name              string
	PathParams        []string
	ParamTypeName     string
	RequestType       string
	ResponseType      string
	MutationVariables string
	EventStream       bool
}

type facadeNode struct {
	Children map[string]*facadeNode
	Op       *routeOperation
}

var pathParamPattern = regexp.MustCompile(`\{([^}/]+)\}`)

func main() {
	openAPIPath := flag.String("openapi", "", "OpenAPI JSON path")
	outPath := flag.String("out", "", "generated TypeScript path")
	flag.Parse()
	if err := run(*openAPIPath, *outPath); err != nil {
		fmt.Fprintln(os.Stderr, "reactquerygen:", err)
		os.Exit(1)
	}
}

func run(openAPIPath, outPath string) error {
	if strings.TrimSpace(openAPIPath) == "" || strings.TrimSpace(outPath) == "" {
		return errors.New("usage: go run ./tools/reactquerygen -openapi <path> -out <path>")
	}
	spec, err := loadOpenAPI(openAPIPath)
	if err != nil {
		return err
	}
	coreBody, err := generate(spec)
	if err != nil {
		return err
	}
	reactBody, err := generateReact(spec)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, coreBody, 0o644); err != nil {
		return err
	}
	return os.WriteFile(reactOutPath(outPath), reactBody, 0o644)
}

func reactOutPath(outPath string) string {
	return strings.TrimSuffix(outPath, filepath.Ext(outPath)) + ".react.ts"
}

func loadOpenAPI(path string) (openAPISpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return openAPISpec{}, fmt.Errorf("read %s: %w", path, err)
	}
	var spec openAPISpec
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&spec); err != nil {
		return openAPISpec{}, fmt.Errorf("decode %s: %w", path, err)
	}
	if len(spec.Paths) == 0 {
		return openAPISpec{}, errors.New("OpenAPI paths are required")
	}
	return spec, nil
}

func generate(spec openAPISpec) ([]byte, error) {
	ops := flattenOperations(spec.Paths)
	if err := validateClientMetadata(spec, ops); err != nil {
		return nil, err
	}

	var b strings.Builder
	writeGeneratedHeader(&b)
	b.WriteString("import type { QueryClient, UseMutationOptions, UseQueryOptions } from '@/lib/react-query';\n\n")
	writeTypes(&b, spec.Components.Schemas)
	writeCoreRuntime(&b)
	for _, op := range ops {
		if err := writeOperation(&b, op); err != nil {
			return nil, err
		}
	}
	if err := writeFacadeInterfaces(&b, spec, ops, false); err != nil {
		return nil, err
	}
	if err := writeFacade(&b, ops); err != nil {
		return nil, err
	}
	out := bytes.TrimRight([]byte(b.String()), "\n")
	return append(out, '\n'), nil
}

func generateReact(spec openAPISpec) ([]byte, error) {
	ops := flattenOperations(spec.Paths)
	if err := validateClientMetadata(spec, ops); err != nil {
		return nil, err
	}

	var b strings.Builder
	b.WriteString("'use client';\n\n")
	writeGeneratedHeader(&b)
	b.WriteString("/* eslint-disable react-hooks/rules-of-hooks */\n\n")
	b.WriteString("import { useMemo } from 'react';\n")
	b.WriteString("import { useMutation, useQuery, type UseMutationResult, type UseQueryResult } from '@/lib/react-query';\n")
	b.WriteString("import * as core from './aiAgentClient';\n\n")
	if err := writeFacadeInterfaces(&b, spec, ops, true); err != nil {
		return nil, err
	}
	if err := writeReactFacade(&b, ops); err != nil {
		return nil, err
	}
	out := bytes.TrimRight([]byte(b.String()), "\n")
	return append(out, '\n'), nil
}

func validateClientMetadata(spec openAPISpec, ops []routeOperation) error {
	if len(spec.ClientModules) == 0 {
		return errors.New("OpenAPI x-riido-client-modules is required")
	}
	modules := map[string]struct{}{}
	for _, module := range spec.ClientModules {
		name := strings.TrimSpace(module.Module)
		if name == "" {
			return errors.New("x-riido-client-modules contains an empty module")
		}
		modules[name] = struct{}{}
	}

	cacheTags := map[string]routeOperation{}
	for _, op := range ops {
		if strings.EqualFold(op.Method, "GET") {
			cacheTag := strings.TrimSpace(op.Op.Client.CacheTag)
			if cacheTag == "" {
				return fmt.Errorf("%s %s missing x-riido-client.cache_tag", strings.ToUpper(op.Method), op.Path)
			}
			if prev, exists := cacheTags[cacheTag]; exists {
				return fmt.Errorf("duplicate x-riido-client.cache_tag %q on %s and %s", cacheTag, prev.Op.OperationID, op.Op.OperationID)
			}
			cacheTags[cacheTag] = op
		}
	}

	for _, op := range ops {
		module := strings.TrimSpace(op.Op.Client.Module)
		if module == "" {
			return fmt.Errorf("%s %s missing x-riido-client.module", strings.ToUpper(op.Method), op.Path)
		}
		if _, ok := modules[module]; !ok {
			return fmt.Errorf("%s %s references unknown x-riido-client.module %q", strings.ToUpper(op.Method), op.Path, module)
		}
		if len(op.Op.Client.FacadePath) == 0 {
			return fmt.Errorf("%s %s missing x-riido-client.facade_path", strings.ToUpper(op.Method), op.Path)
		}
		for _, segment := range op.Op.Client.FacadePath {
			if strings.TrimSpace(segment) == "" {
				return fmt.Errorf("%s %s has empty x-riido-client.facade_path segment", strings.ToUpper(op.Method), op.Path)
			}
		}
		if generatedPath := strings.TrimSpace(op.Op.Client.GeneratedPath); generatedPath != "" {
			want := generatedPathFromClient(op.Op.Client)
			if generatedPath != want {
				return fmt.Errorf("%s %s has x-riido-client.generated_path %q, want %q", strings.ToUpper(op.Method), op.Path, generatedPath, want)
			}
		}
		for _, invalidates := range op.Op.Client.Invalidates {
			if _, ok := cacheTags[invalidates]; !ok {
				return fmt.Errorf("%s invalidates unknown cache tag %q", op.Op.OperationID, invalidates)
			}
		}
	}
	return nil
}

func writeGeneratedHeader(b *strings.Builder) {
	b.WriteString("// 이 파일은 tools/reactquerygen으로 생성됩니다. 직접 수정하지 마세요.\n")
	b.WriteString("// 원본: contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json\n\n")
}

func writeTypes(b *strings.Builder, schemas map[string]schema) {
	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		s := schemas[name]
		if len(s.Enum) > 0 {
			writeJSDoc(b, typeDescription(name, s))
			fmt.Fprintf(b, "export type %s = %s;\n\n", name, stringUnion(s.Enum))
			continue
		}
		if len(s.OneOf) > 0 {
			parts := make([]string, 0, len(s.OneOf))
			for _, item := range s.OneOf {
				parts = append(parts, schemaType(item, true))
			}
			writeJSDoc(b, typeDescription(name, s))
			fmt.Fprintf(b, "export type %s = %s;\n\n", name, strings.Join(parts, " | "))
			continue
		}
		if s.Type == "object" || len(s.Properties) > 0 {
			required := map[string]struct{}{}
			for _, field := range s.Required {
				required[field] = struct{}{}
			}
			writeJSDoc(b, typeDescription(name, s))
			fmt.Fprintf(b, "export interface %s {\n", name)
			fields := make([]string, 0, len(s.Properties))
			for field := range s.Properties {
				fields = append(fields, field)
			}
			sort.Strings(fields)
			for _, field := range fields {
				_, ok := required[field]
				if description := strings.TrimSpace(s.Properties[field].Description); description != "" {
					writeIndentedJSDoc(b, "  ", description)
				}
				fmt.Fprintf(b, "  %s%s: %s;\n", quoteProperty(field), optionalMark(ok), schemaType(s.Properties[field], true))
			}
			b.WriteString("}\n\n")
		}
	}
}

func writeCoreRuntime(b *strings.Builder) {
	writeJSDoc(b, "앱에서 사용하는 fetch 구현을 주입하기 위한 타입입니다.")
	b.WriteString("export type RiidoFetcher = typeof fetch;\n\n")
	writeJSDoc(b,
		"control-plane 호출에 필요한 기본 설정입니다.",
		"`baseUrl`은 예: `http://ai-api.riido.io`처럼 마지막 슬래시 없이 전달해도 됩니다.",
		"`aiAgentToken`은 기존 Riido 앱 로그인 토큰과 구분되는 AI Agent SaaS 전용 토큰입니다.",
		"요청에는 `X-Riido-AI-Agent-Token` 헤더로 전달됩니다.",
		"`fetcher`는 테스트나 앱 공통 transport 래핑이 필요할 때만 주입합니다.",
	)
	b.WriteString("export interface RiidoClientConfig {\n  baseUrl: string;\n  aiAgentToken: string;\n  fetcher?: RiidoFetcher;\n}\n\n")
	writeJSDoc(b, "요청 단위로 전달할 수 있는 옵션입니다. 현재는 AbortSignal만 전달합니다.")
	b.WriteString("export interface RiidoRequestOptions {\n  signal?: AbortSignal;\n}\n\n")
	writeJSDoc(b, "React Query query option에 Riido 요청 옵션을 함께 전달하기 위한 타입입니다.")
	b.WriteString("export type RiidoQueryOptions<TData> = Omit<UseQueryOptions<TData>, 'queryKey' | 'queryFn'> & RiidoRequestOptions;\n\n")
	writeJSDoc(b, "React Query mutation option을 Riido endpoint 변수 타입과 묶은 타입입니다.")
	b.WriteString("export type RiidoMutationOptions<TData, TVariables> = UseMutationOptions<TData, Error, TVariables>;\n\n")
	b.WriteString("async function riidoRequest<T>(config: RiidoClientConfig, path: string, init: RequestInit = {}): Promise<T> {\n")
	b.WriteString("  const fetcher = config.fetcher ?? fetch;\n")
	b.WriteString("  const response = await fetcher(`${config.baseUrl.replace(/\\/$/, '')}${path}`, {\n")
	b.WriteString("    ...init,\n    headers: {\n      Accept: 'application/json',\n      'X-Riido-AI-Agent-Token': config.aiAgentToken,\n      ...(init.body ? { 'Content-Type': 'application/json' } : {}),\n      ...init.headers,\n    },\n  });\n")
	b.WriteString("  if (!response.ok) {\n    throw new Error(`Riido API ${response.status}: ${await response.text()}`);\n  }\n")
	b.WriteString("  return response.json() as Promise<T>;\n}\n\n")
	b.WriteString("async function riidoRawRequest(config: RiidoClientConfig, path: string, init: RequestInit = {}): Promise<Response> {\n")
	b.WriteString("  const fetcher = config.fetcher ?? fetch;\n")
	b.WriteString("  const response = await fetcher(`${config.baseUrl.replace(/\\/$/, '')}${path}`, {\n")
	b.WriteString("    ...init,\n    headers: {\n      'X-Riido-AI-Agent-Token': config.aiAgentToken,\n      ...init.headers,\n    },\n  });\n")
	b.WriteString("  if (!response.ok) {\n    throw new Error(`Riido API ${response.status}: ${await response.text()}`);\n  }\n")
	b.WriteString("  return response;\n}\n\n")
}

func writeOperation(b *strings.Builder, op routeOperation) error {
	info, err := newOperationInfo(op)
	if err != nil {
		return err
	}
	if len(info.PathParams) > 0 {
		writeJSDoc(b, operationSummary(op), "경로 파라미터입니다.")
		fmt.Fprintf(b, "export interface %s {\n", info.ParamTypeName)
		for _, param := range info.PathParams {
			fmt.Fprintf(b, "  %s: string;\n", safeIdentifier(param))
		}
		b.WriteString("}\n\n")
	}

	args := []string{"config: RiidoClientConfig"}
	if len(info.PathParams) > 0 {
		args = append(args, "params: "+info.ParamTypeName)
	}
	if info.RequestType != "" {
		args = append(args, "body: "+info.RequestType)
	}
	args = append(args, "options: RiidoRequestOptions = {}")
	writeJSDoc(b, operationSummary(op))
	fmt.Fprintf(b, "export async function %s(%s): Promise<%s> {\n", info.Name, strings.Join(args, ", "), info.ResponseType)
	fmt.Fprintf(b, "  const path = %s;\n", pathTemplate(op.Path, info.PathParams))
	initParts := []string{fmt.Sprintf("method: '%s'", strings.ToUpper(op.Method)), "signal: options.signal"}
	if info.RequestType != "" {
		initParts = append(initParts, "body: JSON.stringify(body)")
	}
	if info.EventStream {
		fmt.Fprintf(b, "  return riidoRawRequest(config, path, { %s });\n", strings.Join(initParts, ", "))
	} else {
		fmt.Fprintf(b, "  return riidoRequest<%s>(config, path, { %s });\n", info.ResponseType, strings.Join(initParts, ", "))
	}
	b.WriteString("}\n\n")

	if strings.EqualFold(op.Method, "GET") {
		writeQueryOperation(b, info)
		return nil
	}
	writeMutationOperation(b, info)
	return nil
}

func writeQueryOperation(b *strings.Builder, info operationInfo) {
	writeJSDoc(b,
		operationSummary(info.Route),
		fmt.Sprintf("cache tag: `%s`", info.Route.Op.Client.CacheTag),
		"이 endpoint cache 전체를 무효화할 때 사용하는 root query key입니다.",
	)
	fmt.Fprintf(b, "export function %s(): readonly unknown[] {\n", queryKeyRootFunctionName(info.Name))
	fmt.Fprintf(b, "  return [%q] as const;\n", info.Route.Op.Client.CacheTag)
	b.WriteString("}\n\n")

	writeJSDoc(b, operationSummary(info.Route), "이 호출에 사용하는 React Query 키입니다.")
	fmt.Fprintf(b, "export function %s(%s): readonly unknown[] {\n", queryKeyFunctionName(info.Name), queryKeyArgs(info.PathParams, info.ParamTypeName, info.RequestType))
	fmt.Fprintf(b, "  return [...%s()%s] as const;\n", queryKeyRootFunctionName(info.Name), queryKeyTail(info.PathParams, info.RequestType))
	b.WriteString("}\n\n")

	writeJSDoc(b, operationSummary(info.Route), "useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.")
	fmt.Fprintf(b, "export function %s(%s) {\n", queryOptionsFunctionName(info.Name), queryOptionsSignature(info, true))
	b.WriteString("  const { signal, ...queryOptions } = options;\n")
	fmt.Fprintf(b, "  return {\n    ...queryOptions,\n    queryKey: %s(%s),\n    queryFn: () => %s(%s),\n  };\n",
		queryKeyFunctionName(info.Name),
		strings.Join(queryKeyCallArgs(info.PathParams, info.RequestType), ", "),
		info.Name,
		strings.Join(queryRequestCallArgs(info, "signal"), ", "),
	)
	b.WriteString("}\n\n")
}

func writeMutationOperation(b *strings.Builder, info operationInfo) {
	info.MutationVariables = writeMutationVariables(b, info)
	writeJSDoc(b, operationSummary(info.Route), "이 mutation을 구분하는 React Query mutation key입니다.")
	fmt.Fprintf(b, "export function %s(): readonly unknown[] {\n", mutationKeyFunctionName(info.Name))
	fmt.Fprintf(b, "  return [%q] as const;\n", info.Name)
	b.WriteString("}\n\n")

	writeJSDoc(b, operationSummary(info.Route), "useMutation에 전달할 수 있는 옵션입니다.")
	fmt.Fprintf(b, "export function %s(config: RiidoClientConfig, options: RiidoMutationOptions<%s, %s> = {}) {\n", mutationOptionsFunctionName(info.Name), info.ResponseType, info.MutationVariables)
	fmt.Fprintf(b, "  return {\n    ...options,\n    mutationKey: %s(),\n    mutationFn: (%s) => ", mutationKeyFunctionName(info.Name), mutationFunctionVariable(info.MutationVariables))
	callArgs := []string{"config"}
	if len(info.PathParams) > 0 {
		callArgs = append(callArgs, "variables.params")
	}
	if info.RequestType != "" {
		callArgs = append(callArgs, "variables.body")
	}
	callArgs = append(callArgs, "{}")
	fmt.Fprintf(b, "%s(%s),\n", info.Name, strings.Join(callArgs, ", "))
	b.WriteString("  };\n")
	b.WriteString("}\n\n")
}

func writeMutationVariables(b *strings.Builder, info operationInfo) string {
	if len(info.PathParams) == 0 && info.RequestType == "" {
		return "void"
	}
	typeName := mutationVariableTypeName(info.Name, info.PathParams, info.RequestType)
	writeJSDoc(b, operationSummary(info.Route), "mutation 함수에 전달하는 변수입니다.")
	fmt.Fprintf(b, "export interface %s {\n", typeName)
	if len(info.PathParams) > 0 {
		fmt.Fprintf(b, "  params: %s;\n", info.ParamTypeName)
	}
	if info.RequestType != "" {
		fmt.Fprintf(b, "  body: %s;\n", info.RequestType)
	}
	b.WriteString("}\n\n")
	return typeName
}

func writeFacadeInterfaces(b *strings.Builder, spec openAPISpec, ops []routeOperation, react bool) error {
	root, err := buildFacadeTree(ops)
	if err != nil {
		return err
	}
	if react {
		if err := writeReactEndpointInterfaces(b, ops); err != nil {
			return err
		}
	} else {
		if err := writeCoreEndpointInterfaces(b, ops); err != nil {
			return err
		}
	}

	descriptions := namespaceDescriptions(spec)
	moduleDescriptions := moduleDescriptions(spec)
	for _, moduleName := range sortedNodeNames(root) {
		writeNamespaceInterface(b, moduleName, nil, root.Children[moduleName], descriptions, moduleDescriptions[moduleName], react)
	}
	writeControlPlaneClientInterface(b, root, moduleDescriptions, react)
	return nil
}

func writeCoreEndpointInterfaces(b *strings.Builder, ops []routeOperation) error {
	cacheTargets := queryOperationByCacheTag(ops)
	for _, op := range ops {
		info, err := newOperationInfo(op)
		if err != nil {
			return err
		}
		if strings.EqualFold(op.Method, "GET") {
			writeCoreQueryEndpointInterface(b, info)
			continue
		}
		info.MutationVariables = mutationVariableTypeName(info.Name, info.PathParams, info.RequestType)
		writeCoreMutationEndpointInterface(b, info, cacheTargets)
	}
	return nil
}

func writeCoreQueryEndpointInterface(b *strings.Builder, info operationInfo) {
	writeJSDoc(b, append(operationGeneratedPathCommentLines(info.Route),
		fmt.Sprintf("cache tag: `%s`", info.Route.Op.Client.CacheTag),
	)...)
	fmt.Fprintf(b, "export interface %s {\n", endpointInterfaceName(info.Name))
	writeIndentedJSDoc(b, "  ", "HTTP 요청을 직접 실행합니다.")
	fmt.Fprintf(b, "  readonly request: (%s) => Promise<%s>;\n", facadeQueryRequestSignature(info.PathParams, info.ParamTypeName), info.ResponseType)
	writeIndentedJSDoc(b, "  ", "이 endpoint cache 전체를 가리키는 root query key입니다.")
	fmt.Fprintf(b, "  readonly queryKeyRoot: () => readonly unknown[];\n")
	writeIndentedJSDoc(b, "  ", "특정 호출을 가리키는 query key입니다.")
	fmt.Fprintf(b, "  readonly queryKey: (%s) => readonly unknown[];\n", queryKeyArgs(info.PathParams, info.ParamTypeName, info.RequestType))
	writeIndentedJSDoc(b, "  ", "useQuery에 전달할 수 있는 query option입니다.")
	fmt.Fprintf(b, "  readonly query: (%s) => ReturnType<typeof %s>;\n", facadeQueryOptionsSignature(info, false), queryOptionsFunctionName(info.Name))
	writeIndentedJSDoc(b, "  ", "query와 동일합니다. prefetchQuery 등 명시적인 React Query API에 넘길 때 사용합니다.")
	fmt.Fprintf(b, "  readonly queryOptions: (%s) => ReturnType<typeof %s>;\n", facadeQueryOptionsSignature(info, false), queryOptionsFunctionName(info.Name))
	writeIndentedJSDoc(b, "  ", "특정 query key만 무효화합니다. 화면 정책에 맞춰 client 코드가 호출 여부를 결정합니다.")
	fmt.Fprintf(b, "  readonly invalidate: (%s) => Promise<void>;\n", facadeInvalidateSignature(info))
	writeIndentedJSDoc(b, "  ", "이 endpoint의 root cache tag 전체를 무효화합니다.")
	fmt.Fprintf(b, "  readonly invalidateAll: (queryClient: QueryClient) => Promise<void>;\n")
	writeIndentedJSDoc(b, "  ", "현재 endpoint를 prefetch합니다.")
	fmt.Fprintf(b, "  readonly prefetch: (%s) => Promise<void>;\n", facadePrefetchSignature(info))
	b.WriteString("}\n\n")
}

func writeCoreMutationEndpointInterface(b *strings.Builder, info operationInfo, cacheTargets map[string]routeOperation) {
	writeJSDoc(b, append(operationGeneratedPathCommentLines(info.Route),
		"자동 무효화는 하지 않습니다. 화면 정책에 맞춰 invalidates helper를 명시적으로 호출합니다.",
	)...)
	fmt.Fprintf(b, "export interface %s {\n", endpointInterfaceName(info.Name))
	writeIndentedJSDoc(b, "  ", "HTTP 요청을 직접 실행합니다.")
	fmt.Fprintf(b, "  readonly request: (%s) => Promise<%s>;\n", facadeMutationRequestSignature(info.PathParams, info.ParamTypeName, info.RequestType), info.ResponseType)
	writeIndentedJSDoc(b, "  ", "이 mutation을 구분하는 key입니다.")
	fmt.Fprintf(b, "  readonly mutationKey: () => readonly unknown[];\n")
	writeIndentedJSDoc(b, "  ", "useMutation에 전달할 수 있는 mutation option입니다.")
	fmt.Fprintf(b, "  readonly mutation: (options?: RiidoMutationOptions<%s, %s>) => ReturnType<typeof %s>;\n", info.ResponseType, info.MutationVariables, mutationOptionsFunctionName(info.Name))
	writeIndentedJSDoc(b, "  ", "mutation과 동일합니다. React Query API에 명시적으로 넘길 때 사용합니다.")
	fmt.Fprintf(b, "  readonly mutationOptions: (options?: RiidoMutationOptions<%s, %s>) => ReturnType<typeof %s>;\n", info.ResponseType, info.MutationVariables, mutationOptionsFunctionName(info.Name))
	writeIndentedJSDoc(b, "  ", "이 command 이후 client가 선택적으로 무효화할 수 있는 cache helper입니다.")
	b.WriteString("  readonly invalidates: {\n")
	for _, tag := range info.Route.Op.Client.Invalidates {
		target := cacheTargets[tag]
		writeIndentedJSDoc(b, "    ", fmt.Sprintf("`%s` cache tag를 무효화합니다.", tag))
		fmt.Fprintf(b, "    readonly %s: (queryClient: QueryClient) => Promise<void>;\n", invalidationPropertyName(tag, info.Route.Op.Client.Module))
		_ = target
	}
	if len(info.Route.Op.Client.Invalidates) > 0 {
		writeIndentedJSDoc(b, "    ", "선언된 모든 cache tag를 한 번에 무효화합니다.")
		b.WriteString("    readonly all: (queryClient: QueryClient) => Promise<void[]>;\n")
	}
	b.WriteString("  };\n")
	b.WriteString("}\n\n")
}

func writeReactEndpointInterfaces(b *strings.Builder, ops []routeOperation) error {
	for _, op := range ops {
		info, err := newOperationInfo(op)
		if err != nil {
			return err
		}
		if strings.EqualFold(op.Method, "GET") {
			writeJSDoc(b, append(operationGeneratedPathCommentLines(op), "client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.")...)
			fmt.Fprintf(b, "export interface %s extends core.%s {\n", reactEndpointInterfaceName(info.Name), endpointInterfaceName(info.Name))
			writeIndentedJSDoc(b, "  ", "React Query useQuery hook입니다.")
			fmt.Fprintf(b, "  readonly useQuery: (%s) => UseQueryResult<%s, Error>;\n", reactQueryHookSignature(info), reactType(info.ResponseType))
			b.WriteString("}\n\n")
			continue
		}
		info.MutationVariables = mutationVariableTypeName(info.Name, info.PathParams, info.RequestType)
		writeJSDoc(b, append(operationGeneratedPathCommentLines(op), "client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.")...)
		fmt.Fprintf(b, "export interface %s extends core.%s {\n", reactEndpointInterfaceName(info.Name), endpointInterfaceName(info.Name))
		writeIndentedJSDoc(b, "  ", "React Query useMutation hook입니다.")
		fmt.Fprintf(b, "  readonly useMutation: (options?: core.RiidoMutationOptions<%s, %s>) => UseMutationResult<%s, Error, %s>;\n", reactType(info.ResponseType), reactType(info.MutationVariables), reactType(info.ResponseType), reactType(info.MutationVariables))
		b.WriteString("}\n\n")
	}
	return nil
}

func writeNamespaceInterface(b *strings.Builder, module string, path []string, node *facadeNode, descriptions map[string]string, rootDescription string, react bool) {
	for _, name := range sortedNodeNames(node) {
		child := node.Children[name]
		if len(child.Children) > 0 {
			writeNamespaceInterface(b, module, append(path, name), child, descriptions, rootDescription, react)
		}
	}

	description := rootDescription
	if len(path) > 0 {
		description = descriptions[namespaceKey(module, path)]
	}
	if description == "" {
		description = strings.Join(append([]string{module}, path...), ".") + " namespace입니다."
	}
	writeJSDoc(b, description)
	interfaceName := namespaceInterfaceName(module, path, react)
	if node.Op != nil {
		info, _ := newOperationInfo(*node.Op)
		fmt.Fprintf(b, "export interface %s extends %s {\n", interfaceName, operationEndpointType(info.Name, react))
	} else {
		fmt.Fprintf(b, "export interface %s {\n", interfaceName)
	}
	for _, name := range sortedNodeNames(node) {
		child := node.Children[name]
		if child.Op != nil && len(child.Children) == 0 {
			info, _ := newOperationInfo(*child.Op)
			writeIndentedJSDoc(b, "  ", operationPropertyDescriptionLines(info)...)
			fmt.Fprintf(b, "  readonly %s: %s;\n", quoteProperty(name), operationEndpointType(info.Name, react))
			continue
		}
		childPath := append(path, name)
		desc := descriptions[namespaceKey(module, childPath)]
		if desc == "" {
			if child.Op != nil {
				info, _ := newOperationInfo(*child.Op)
				desc = strings.Join(operationPropertyDescriptionLines(info), " ")
			} else {
				desc = strings.Join(append([]string{module}, childPath...), ".") + " namespace입니다."
			}
		}
		writeIndentedJSDoc(b, "  ", desc)
		fmt.Fprintf(b, "  readonly %s: %s;\n", quoteProperty(name), namespaceInterfaceName(module, childPath, react))
	}
	b.WriteString("}\n\n")
}

func writeControlPlaneClientInterface(b *strings.Builder, root *facadeNode, moduleDescriptions map[string]string, react bool) {
	name := "RiidoControlPlaneClient"
	if react {
		name = "RiidoControlPlaneReactClient"
	}
	writeJSDoc(b, "control-plane API를 DSL client metadata의 module/namespace별로 묶은 config-bound facade입니다.")
	fmt.Fprintf(b, "export interface %s {\n", name)
	for _, module := range sortedNodeNames(root) {
		description := moduleDescriptions[module]
		if description == "" {
			description = module + " module입니다."
		}
		writeIndentedJSDoc(b, "  ", description)
		fmt.Fprintf(b, "  readonly %s: %s;\n", quoteProperty(module), namespaceInterfaceName(module, nil, react))
	}
	b.WriteString("}\n\n")
}

func writeFacade(b *strings.Builder, ops []routeOperation) error {
	root, err := buildFacadeTree(ops)
	if err != nil {
		return err
	}
	cacheTargets := queryOperationByCacheTag(ops)
	writeJSDoc(b,
		"control-plane API를 DSL client metadata의 module/namespace별로 묶은 config-bound facade입니다.",
		"React QueryClient를 대체하지 않고 request, query/queryOptions, mutation/mutationOptions와 명시적 cache helper만 제공합니다.",
	)
	b.WriteString("export function createRiidoControlPlaneClient(config: RiidoClientConfig): RiidoControlPlaneClient {\n")
	b.WriteString("  return {\n")
	writeFacadeChildren(b, root, nil, "    ", cacheTargets)
	b.WriteString("  };\n")
	b.WriteString("}\n\n")
	return nil
}

func writeFacadeChildren(b *strings.Builder, node *facadeNode, path []string, indent string, cacheTargets map[string]routeOperation) {
	for _, name := range sortedNodeNames(node) {
		child := node.Children[name]
		fmt.Fprintf(b, "%s%s: {\n", indent, quoteProperty(name))
		if child.Op != nil {
			writeFacadeOperation(b, *child.Op, append(path, name), indent+"  ", cacheTargets)
		}
		writeFacadeChildren(b, child, append(path, name), indent+"  ", cacheTargets)
		fmt.Fprintf(b, "%s},\n", indent)
	}
}

func writeFacadeOperation(b *strings.Builder, op routeOperation, _ []string, indent string, cacheTargets map[string]routeOperation) {
	info, _ := newOperationInfo(op)
	if strings.EqualFold(op.Method, "GET") {
		fmt.Fprintf(b, "%srequest: (%s) => %s(%s),\n", indent, facadeQueryRequestSignature(info.PathParams, info.ParamTypeName), info.Name, facadeQueryRequestCallArgs(info.PathParams))
		fmt.Fprintf(b, "%squeryKeyRoot: %s,\n", indent, queryKeyRootFunctionName(info.Name))
		fmt.Fprintf(b, "%squeryKey: %s,\n", indent, queryKeyFunctionName(info.Name))
		fmt.Fprintf(b, "%squery: (%s) => %s(%s),\n", indent, facadeQueryOptionsSignature(info, true), queryOptionsFunctionName(info.Name), facadeQueryOptionsCallArgs(info))
		fmt.Fprintf(b, "%squeryOptions: (%s) => %s(%s),\n", indent, facadeQueryOptionsSignature(info, true), queryOptionsFunctionName(info.Name), facadeQueryOptionsCallArgs(info))
		fmt.Fprintf(b, "%sinvalidate: (%s) => queryClient.invalidateQueries({ queryKey: %s(%s) }),\n", indent, facadeInvalidateSignature(info), queryKeyFunctionName(info.Name), strings.Join(queryKeyCallArgs(info.PathParams, info.RequestType), ", "))
		fmt.Fprintf(b, "%sinvalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: %s() }),\n", indent, queryKeyRootFunctionName(info.Name))
		fmt.Fprintf(b, "%sprefetch: (%s) => queryClient.prefetchQuery(%s(%s)),\n", indent, facadePrefetchSignature(info), queryOptionsFunctionName(info.Name), facadePrefetchCallArgs(info))
		return
	}
	info.MutationVariables = mutationVariableTypeName(info.Name, info.PathParams, info.RequestType)
	fmt.Fprintf(b, "%srequest: (%s) => %s(%s),\n", indent, facadeMutationRequestSignature(info.PathParams, info.ParamTypeName, info.RequestType), info.Name, facadeMutationRequestCallArgs(info.PathParams, info.RequestType))
	fmt.Fprintf(b, "%smutationKey: %s,\n", indent, mutationKeyFunctionName(info.Name))
	fmt.Fprintf(b, "%smutation: (options: RiidoMutationOptions<%s, %s> = {}) => %s(config, options),\n", indent, info.ResponseType, info.MutationVariables, mutationOptionsFunctionName(info.Name))
	fmt.Fprintf(b, "%smutationOptions: (options: RiidoMutationOptions<%s, %s> = {}) => %s(config, options),\n", indent, info.ResponseType, info.MutationVariables, mutationOptionsFunctionName(info.Name))
	b.WriteString(indent + "invalidates: {\n")
	allCalls := make([]string, 0, len(op.Op.Client.Invalidates))
	for _, tag := range op.Op.Client.Invalidates {
		targetOp := cacheTargets[tag]
		if targetOp.Op.OperationID == "" {
			continue
		}
		prop := invalidationPropertyName(tag, op.Op.Client.Module)
		fn := queryKeyRootFunctionName(targetOp.Op.OperationID)
		fmt.Fprintf(b, "%s  %s: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: %s() }),\n", indent, prop, fn)
		allCalls = append(allCalls, fmt.Sprintf("queryClient.invalidateQueries({ queryKey: %s() })", fn))
	}
	if len(allCalls) > 0 {
		fmt.Fprintf(b, "%s  all: (queryClient: QueryClient) => Promise.all([%s]),\n", indent, strings.Join(allCalls, ", "))
	}
	b.WriteString(indent + "},\n")
}

func writeReactFacade(b *strings.Builder, ops []routeOperation) error {
	root, err := buildFacadeTree(ops)
	if err != nil {
		return err
	}
	writeJSDoc(b,
		"control-plane API facade에 React Query hook을 얹은 client 전용 wrapper입니다.",
		"hook은 반드시 `@/lib/react-query`를 통과하므로 riido-client의 workspace/demo 정책을 우회하지 않습니다.",
	)
	b.WriteString("export function useRiidoControlPlaneClient(config: core.RiidoClientConfig): RiidoControlPlaneReactClient {\n")
	b.WriteString("  const coreClient = useMemo(() => core.createRiidoControlPlaneClient(config), [config.baseUrl, config.fetcher, config.aiAgentToken]);\n\n")
	b.WriteString("  return useMemo(\n")
	b.WriteString("    () => ({\n")
	writeReactFacadeChildren(b, root, nil, "      ")
	b.WriteString("    }),\n")
	b.WriteString("    [coreClient],\n")
	b.WriteString("  );\n")
	b.WriteString("}\n\n")
	return nil
}

func writeReactFacadeChildren(b *strings.Builder, node *facadeNode, path []string, indent string) {
	for _, name := range sortedNodeNames(node) {
		child := node.Children[name]
		nextPath := append(path, name)
		fmt.Fprintf(b, "%s%s: {\n", indent, quoteProperty(name))
		if child.Op != nil {
			writeReactFacadeOperation(b, *child.Op, nextPath, indent+"  ")
		}
		writeReactFacadeChildren(b, child, nextPath, indent+"  ")
		fmt.Fprintf(b, "%s},\n", indent)
	}
}

func writeReactFacadeOperation(b *strings.Builder, op routeOperation, path []string, indent string) {
	info, _ := newOperationInfo(op)
	accessor := "coreClient." + strings.Join(path, ".")
	fmt.Fprintf(b, "%s...%s,\n", indent, accessor)
	if strings.EqualFold(op.Method, "GET") {
		fmt.Fprintf(b, "%suseQuery: (%s) => useQuery<%s, Error>(%s.query(%s)),\n",
			indent,
			reactQueryHookSignature(info),
			reactType(info.ResponseType),
			accessor,
			strings.Join(reactQueryHookCallArgs(info), ", "),
		)
		return
	}
	info.MutationVariables = mutationVariableTypeName(info.Name, info.PathParams, info.RequestType)
	fmt.Fprintf(b, "%suseMutation: (options: core.RiidoMutationOptions<%s, %s> = {}) => useMutation<%s, Error, %s>(%s.mutation(options)),\n",
		indent,
		reactType(info.ResponseType),
		reactType(info.MutationVariables),
		reactType(info.ResponseType),
		reactType(info.MutationVariables),
		accessor,
	)
}

func buildFacadeTree(ops []routeOperation) (*facadeNode, error) {
	root := &facadeNode{Children: map[string]*facadeNode{}}
	for _, op := range ops {
		path := facadePathParts(op)
		node := root
		for _, part := range path {
			if node.Children == nil {
				node.Children = map[string]*facadeNode{}
			}
			child := node.Children[part]
			if child == nil {
				child = &facadeNode{Children: map[string]*facadeNode{}}
				node.Children[part] = child
			}
			node = child
		}
		if node.Op != nil {
			return nil, fmt.Errorf("duplicate facade path %s", strings.Join(path, "."))
		}
		opCopy := op
		node.Op = &opCopy
	}
	return root, nil
}

func newOperationInfo(op routeOperation) (operationInfo, error) {
	name := strings.TrimSpace(op.Op.OperationID)
	if name == "" {
		return operationInfo{}, fmt.Errorf("%s %s missing operationId", op.Method, op.Path)
	}
	params := pathParams(op.Path)
	responseType := responseType(op.Op)
	eventStream := isEventStream(op.Op)
	if eventStream {
		responseType = "Response"
	}
	requestType := requestType(op.Op)
	return operationInfo{
		Route:             op,
		Name:              name,
		PathParams:        params,
		ParamTypeName:     exportedName(name) + "PathParams",
		RequestType:       requestType,
		ResponseType:      responseType,
		MutationVariables: mutationVariableTypeName(name, params, requestType),
		EventStream:       eventStream,
	}, nil
}

func queryOperationByCacheTag(ops []routeOperation) map[string]routeOperation {
	out := map[string]routeOperation{}
	for _, op := range ops {
		if strings.EqualFold(op.Method, "GET") && op.Op.Client.CacheTag != "" {
			out[op.Op.Client.CacheTag] = op
		}
	}
	return out
}

func flattenOperations(paths map[string]map[string]operation) []routeOperation {
	var ops []routeOperation
	for path, byMethod := range paths {
		for method, op := range byMethod {
			ops = append(ops, routeOperation{Method: method, Path: path, Op: op})
		}
	}
	sort.Slice(ops, func(i, j int) bool {
		if ops[i].Path != ops[j].Path {
			return ops[i].Path < ops[j].Path
		}
		return ops[i].Method < ops[j].Method
	})
	return ops
}

func sortedNodeNames(node *facadeNode) []string {
	names := make([]string, 0, len(node.Children))
	for name := range node.Children {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func moduleDescriptions(spec openAPISpec) map[string]string {
	out := map[string]string{}
	for _, module := range spec.ClientModules {
		out[module.Module] = module.Description
	}
	return out
}

func namespaceDescriptions(spec openAPISpec) map[string]string {
	out := map[string]string{}
	for _, module := range spec.ClientModules {
		for _, namespace := range module.Namespaces {
			out[namespaceKey(module.Module, namespace.Path)] = namespace.Description
		}
	}
	return out
}

func namespaceKey(module string, path []string) string {
	return module + "." + strings.Join(path, ".")
}

func facadePathParts(op routeOperation) []string {
	return append([]string{op.Op.Client.Module}, op.Op.Client.FacadePath...)
}

func generatedPathFromClient(client clientMetadata) string {
	if strings.TrimSpace(client.Module) == "" || len(client.FacadePath) == 0 {
		return ""
	}
	return client.Module + "." + strings.Join(client.FacadePath, ".")
}

func contractGeneratedPath(op routeOperation) string {
	if generatedPath := strings.TrimSpace(op.Op.Client.GeneratedPath); generatedPath != "" {
		return generatedPath
	}
	return generatedPathFromClient(op.Op.Client)
}

func moduleLocalGeneratedPath(op routeOperation) string {
	return strings.Join(op.Op.Client.FacadePath, ".")
}

func generatedAccessPath(op routeOperation) string {
	return "riido." + contractGeneratedPath(op)
}

func operationGeneratedPathCommentLines(op routeOperation) []string {
	return []string{
		operationSummary(op),
		fmt.Sprintf("계약 generated path: `%s`", contractGeneratedPath(op)),
		fmt.Sprintf("검색용 generated 경로: `%s`", moduleLocalGeneratedPath(op)),
		fmt.Sprintf("접근 예시: `%s`", generatedAccessPath(op)),
	}
}

func operationPropertyDescriptionLines(info operationInfo) []string {
	lines := operationGeneratedPathCommentLines(info.Route)
	if strings.EqualFold(info.Route.Method, "GET") {
		lines = append(lines, fmt.Sprintf("cache tag: `%s`", info.Route.Op.Client.CacheTag))
	} else if len(info.Route.Op.Client.Invalidates) > 0 {
		lines = append(lines, "invalidates: `"+strings.Join(info.Route.Op.Client.Invalidates, "`, `")+"`")
	}
	return lines
}

func operationEndpointType(operationID string, react bool) string {
	if react {
		return reactEndpointInterfaceName(operationID)
	}
	return endpointInterfaceName(operationID)
}

func endpointInterfaceName(operationID string) string {
	return exportedName(operationID) + "Endpoint"
}

func reactEndpointInterfaceName(operationID string) string {
	return exportedName(operationID) + "ReactEndpoint"
}

func namespaceInterfaceName(module string, path []string, react bool) string {
	parts := []string{"Riido", facadeTypeSegment(module)}
	for _, part := range path {
		parts = append(parts, facadeTypeSegment(part))
	}
	if react {
		parts = append(parts, "React")
	}
	if len(path) == 0 {
		parts = append(parts, "Module")
	} else {
		parts = append(parts, "Namespace")
	}
	return strings.Join(parts, "")
}

func facadeTypeSegment(segment string) string {
	switch segment {
	case "aiAgent":
		return "AIAgent"
	default:
		return exportedName(segment)
	}
}

func queryKeyRootFunctionName(operationID string) string {
	return operationID + "QueryKeyRoot"
}

func queryKeyFunctionName(operationID string) string {
	return operationID + "QueryKey"
}

func queryOptionsFunctionName(operationID string) string {
	return operationID + "QueryOptions"
}

func mutationKeyFunctionName(operationID string) string {
	return operationID + "MutationKey"
}

func mutationOptionsFunctionName(operationID string) string {
	return operationID + "MutationOptions"
}

func mutationVariableTypeName(operationID string, params []string, requestType string) string {
	if len(params) == 0 && requestType == "" {
		return "void"
	}
	return exportedName(operationID) + "MutationVariables"
}

func mutationFunctionVariable(mutationVariable string) string {
	if mutationVariable == "void" {
		return ""
	}
	return "variables: " + mutationVariable
}

func queryOptionsSignature(info operationInfo, withDefault bool) string {
	args := []string{"config: RiidoClientConfig"}
	if len(info.PathParams) > 0 {
		args = append(args, "params: "+info.ParamTypeName)
	}
	args = append(args, defaultable(fmt.Sprintf("options: RiidoQueryOptions<%s>", info.ResponseType), withDefault))
	return strings.Join(args, ", ")
}

func facadeQueryRequestSignature(params []string, paramTypeName string) string {
	if len(params) > 0 {
		return fmt.Sprintf("params: %s, options?: RiidoRequestOptions", paramTypeName)
	}
	return "options?: RiidoRequestOptions"
}

func facadeQueryRequestCallArgs(params []string) string {
	if len(params) > 0 {
		return "config, params, options"
	}
	return "config, options"
}

func facadeQueryOptionsSignature(info operationInfo, withDefault bool) string {
	options := defaultable(fmt.Sprintf("options: RiidoQueryOptions<%s>", info.ResponseType), withDefault)
	if len(info.PathParams) > 0 {
		return fmt.Sprintf("params: %s, %s", info.ParamTypeName, options)
	}
	return options
}

func facadeQueryOptionsCallArgs(info operationInfo) string {
	args := []string{"config"}
	if len(info.PathParams) > 0 {
		args = append(args, "params")
	}
	args = append(args, "options")
	return strings.Join(args, ", ")
}

func facadeInvalidateSignature(info operationInfo) string {
	args := []string{"queryClient: QueryClient"}
	if len(info.PathParams) > 0 {
		args = append(args, "params: "+info.ParamTypeName)
	}
	return strings.Join(args, ", ")
}

func facadePrefetchSignature(info operationInfo) string {
	args := []string{"queryClient: QueryClient"}
	if len(info.PathParams) > 0 {
		args = append(args, "params: "+info.ParamTypeName)
	}
	args = append(args, fmt.Sprintf("options?: RiidoQueryOptions<%s>", info.ResponseType))
	return strings.Join(args, ", ")
}

func facadePrefetchCallArgs(info operationInfo) string {
	args := []string{"config"}
	if len(info.PathParams) > 0 {
		args = append(args, "params")
	}
	args = append(args, "options")
	return strings.Join(args, ", ")
}

func facadeMutationRequestSignature(params []string, paramTypeName, requestType string) string {
	var args []string
	if len(params) > 0 {
		args = append(args, "params: "+paramTypeName)
	}
	if requestType != "" {
		args = append(args, "body: "+requestType)
	}
	args = append(args, "options?: RiidoRequestOptions")
	return strings.Join(args, ", ")
}

func facadeMutationRequestCallArgs(params []string, requestType string) string {
	args := []string{"config"}
	if len(params) > 0 {
		args = append(args, "params")
	}
	if requestType != "" {
		args = append(args, "body")
	}
	args = append(args, "options")
	return strings.Join(args, ", ")
}

func reactQueryHookSignature(info operationInfo) string {
	options := fmt.Sprintf("options?: core.RiidoQueryOptions<%s>", reactType(info.ResponseType))
	if len(info.PathParams) > 0 {
		return fmt.Sprintf("params: core.%s, %s", info.ParamTypeName, options)
	}
	return options
}

func reactQueryHookCallArgs(info operationInfo) []string {
	var args []string
	if len(info.PathParams) > 0 {
		args = append(args, "params")
	}
	args = append(args, "options")
	return args
}

func reactType(typeName string) string {
	switch typeName {
	case "Response", "void", "unknown":
		return typeName
	default:
		return "core." + typeName
	}
}

func defaultable(signature string, withDefault bool) string {
	if withDefault {
		return signature + " = {}"
	}
	return strings.Replace(signature, "options: ", "options?: ", 1)
}

func queryRequestCallArgs(info operationInfo, signalName string) []string {
	args := []string{"config"}
	if len(info.PathParams) > 0 {
		args = append(args, "params")
	}
	if info.RequestType != "" {
		args = append(args, "body")
	}
	args = append(args, "{ signal }")
	return args
}

func queryKeyArgs(params []string, paramTypeName, requestType string) string {
	var args []string
	if len(params) > 0 {
		args = append(args, "params: "+paramTypeName)
	}
	if requestType != "" {
		args = append(args, "body: "+requestType)
	}
	return strings.Join(args, ", ")
}

func queryKeyCallArgs(params []string, requestType string) []string {
	var args []string
	if len(params) > 0 {
		args = append(args, "params")
	}
	if requestType != "" {
		args = append(args, "body")
	}
	return args
}

func queryKeyTail(params []string, requestType string) string {
	parts := queryKeyCallArgs(params, requestType)
	if len(parts) == 0 {
		return ""
	}
	return ", " + strings.Join(parts, ", ")
}

func invalidationPropertyName(cacheTag string, module string) string {
	prefix := module + "."
	trimmed := strings.TrimPrefix(cacheTag, prefix)
	parts := strings.Split(trimmed, ".")
	for i, part := range parts {
		if i == 0 {
			parts[i] = part
			continue
		}
		parts[i] = exportedName(part)
	}
	return safeIdentifier(strings.Join(parts, ""))
}

func requestType(op operation) string {
	if op.RequestBody == nil {
		return ""
	}
	for _, content := range op.RequestBody.Content {
		return schemaType(content.Schema, true)
	}
	return ""
}

func responseType(op operation) string {
	for _, ok := range successfulResponses(op.Responses) {
		for _, content := range ok.Content {
			return schemaType(content.Schema, true)
		}
	}
	return "unknown"
}

func isEventStream(op operation) bool {
	for _, ok := range successfulResponses(op.Responses) {
		for contentType := range ok.Content {
			if strings.EqualFold(contentType, "text/event-stream") {
				return true
			}
		}
	}
	return false
}

func successfulResponses(responses map[string]response) []response {
	statuses := make([]string, 0, len(responses))
	for status := range responses {
		if len(status) == 3 && status[0] == '2' {
			statuses = append(statuses, status)
		}
	}
	sort.Strings(statuses)
	out := make([]response, 0, len(statuses))
	for _, status := range statuses {
		out = append(out, responses[status])
	}
	return out
}

func schemaType(s schema, topLevel bool) string {
	if s.Ref != "" {
		return refName(s.Ref)
	}
	if len(s.Enum) > 0 {
		return stringUnion(s.Enum)
	}
	if len(s.OneOf) > 0 {
		parts := make([]string, 0, len(s.OneOf))
		for _, item := range s.OneOf {
			parts = append(parts, schemaType(item, true))
		}
		return strings.Join(parts, " | ")
	}
	switch s.Type {
	case "array":
		if s.Items == nil {
			return "unknown[]"
		}
		return schemaType(*s.Items, false) + "[]"
	case "boolean":
		return "boolean"
	case "integer", "number":
		return "number"
	case "object":
		if topLevel {
			return "Record<string, unknown>"
		}
		return "unknown"
	case "string":
		return "string"
	default:
		return "unknown"
	}
}

func refName(ref string) string {
	const prefix = "#/components/schemas/"
	return strings.TrimPrefix(ref, prefix)
}

func stringUnion(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, fmt.Sprintf("%q", value))
	}
	return strings.Join(quoted, " | ")
}

func pathParams(path string) []string {
	matches := pathParamPattern.FindAllStringSubmatch(path, -1)
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		out = append(out, match[1])
	}
	return out
}

func pathTemplate(path string, params []string) string {
	out := fmt.Sprintf("%q", path)
	for _, param := range params {
		out = strings.ReplaceAll(out, "{"+param+"}", "${params."+safeIdentifier(param)+"}")
	}
	if len(params) > 0 {
		return "`" + strings.Trim(out, "\"") + "`"
	}
	return out
}

func typeDescription(name string, s schema) string {
	if description := strings.TrimSpace(s.Description); description != "" {
		return description
	}
	return name + " 타입입니다."
}

func operationSummary(op routeOperation) string {
	if summary := strings.TrimSpace(op.Op.Summary); summary != "" {
		return summary
	}
	return fmt.Sprintf("%s %s", strings.ToUpper(op.Method), op.Path)
}

func writeJSDoc(b *strings.Builder, lines ...string) {
	trimmed := nonEmptyCommentLines(lines)
	if len(trimmed) == 0 {
		return
	}
	b.WriteString("/**\n")
	for _, line := range trimmed {
		fmt.Fprintf(b, " * %s\n", escapeCommentText(line))
	}
	b.WriteString(" */\n")
}

func writeIndentedJSDoc(b *strings.Builder, indent string, lines ...string) {
	trimmed := nonEmptyCommentLines(lines)
	if len(trimmed) == 0 {
		return
	}
	fmt.Fprintf(b, "%s/**\n", indent)
	for _, line := range trimmed {
		fmt.Fprintf(b, "%s * %s\n", indent, escapeCommentText(line))
	}
	fmt.Fprintf(b, "%s */\n", indent)
}

func nonEmptyCommentLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func escapeCommentText(line string) string {
	return strings.ReplaceAll(line, "*/", "* /")
}

func exportedName(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

func optionalMark(required bool) string {
	if required {
		return ""
	}
	return "?"
}

func quoteProperty(name string) string {
	if safeIdentifier(name) == name {
		return name
	}
	return fmt.Sprintf("%q", name)
}

func safeIdentifier(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}
