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
	Paths      map[string]map[string]operation `json:"paths"`
	Components struct {
		Schemas map[string]schema `json:"schemas"`
	} `json:"components"`
}

type operation struct {
	OperationID string              `json:"operationId"`
	Summary     string              `json:"summary"`
	Parameters  []parameter         `json:"parameters"`
	RequestBody *requestBody        `json:"requestBody"`
	Responses   map[string]response `json:"responses"`
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
	body, err := generate(spec)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, body, 0o644)
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
	var b strings.Builder
	b.WriteString("// 이 파일은 tools/reactquerygen으로 생성됩니다. 직접 수정하지 마세요.\n")
	b.WriteString("// 원본: contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json\n\n")
	b.WriteString("import { useMutation, useQuery, type UseMutationOptions, type UseQueryOptions } from '@tanstack/react-query';\n\n")
	writeTypes(&b, spec.Components.Schemas)
	ops := flattenOperations(spec.Paths)
	if len(ops) == 0 {
		return nil, errors.New("OpenAPI operations are required")
	}
	writeJSDoc(&b, "앱에서 사용하는 fetch 구현을 주입하기 위한 타입입니다.")
	b.WriteString("export type RiidoFetcher = typeof fetch;\n\n")
	writeJSDoc(&b,
		"control-plane 호출에 필요한 기본 설정입니다.",
		"`baseUrl`은 예: `http://ai-api.riido.io`처럼 마지막 슬래시 없이 전달해도 됩니다.",
		"`fetcher`는 테스트나 앱 공통 transport 래핑이 필요할 때만 주입합니다.",
	)
	b.WriteString("export interface RiidoClientConfig {\n  baseUrl: string;\n  token: string;\n  fetcher?: RiidoFetcher;\n}\n\n")
	writeJSDoc(&b, "요청 단위로 전달할 수 있는 옵션입니다. 현재는 AbortSignal만 전달합니다.")
	b.WriteString("export interface RiidoRequestOptions {\n  signal?: AbortSignal;\n}\n\n")
	b.WriteString("async function riidoRequest<T>(config: RiidoClientConfig, path: string, init: RequestInit = {}): Promise<T> {\n")
	b.WriteString("  const fetcher = config.fetcher ?? fetch;\n")
	b.WriteString("  const response = await fetcher(`${config.baseUrl.replace(/\\/$/, '')}${path}`, {\n")
	b.WriteString("    ...init,\n    headers: {\n      Accept: 'application/json',\n      Authorization: `Bearer ${config.token}`,\n      ...(init.body ? { 'Content-Type': 'application/json' } : {}),\n      ...init.headers,\n    },\n  });\n")
	b.WriteString("  if (!response.ok) {\n    throw new Error(`Riido API ${response.status}: ${await response.text()}`);\n  }\n")
	b.WriteString("  return response.json() as Promise<T>;\n}\n\n")
	b.WriteString("async function riidoRawRequest(config: RiidoClientConfig, path: string, init: RequestInit = {}): Promise<Response> {\n")
	b.WriteString("  const fetcher = config.fetcher ?? fetch;\n")
	b.WriteString("  const response = await fetcher(`${config.baseUrl.replace(/\\/$/, '')}${path}`, {\n")
	b.WriteString("    ...init,\n    headers: {\n      Authorization: `Bearer ${config.token}`,\n      ...init.headers,\n    },\n  });\n")
	b.WriteString("  if (!response.ok) {\n    throw new Error(`Riido API ${response.status}: ${await response.text()}`);\n  }\n")
	b.WriteString("  return response;\n}\n\n")
	for _, op := range ops {
		if err := writeOperation(&b, op); err != nil {
			return nil, err
		}
	}
	out := bytes.TrimRight([]byte(b.String()), "\n")
	return append(out, '\n'), nil
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

func writeOperation(b *strings.Builder, op routeOperation) error {
	name := op.Op.OperationID
	if name == "" {
		return fmt.Errorf("%s %s missing operationId", op.Method, op.Path)
	}
	pathParams := pathParams(op.Path)
	paramTypeName := exportedName(name) + "PathParams"
	if len(pathParams) > 0 {
		writeJSDoc(b, operationSummary(op), "경로 파라미터입니다.")
		fmt.Fprintf(b, "export interface %s {\n", paramTypeName)
		for _, param := range pathParams {
			fmt.Fprintf(b, "  %s: string;\n", safeIdentifier(param))
		}
		b.WriteString("}\n\n")
	}
	requestType := requestType(op.Op)
	eventStream := isEventStream(op.Op)
	responseType := responseType(op.Op)
	if eventStream {
		responseType = "Response"
	}
	args := []string{"config: RiidoClientConfig"}
	if len(pathParams) > 0 {
		args = append(args, "params: "+paramTypeName)
	}
	if requestType != "" {
		args = append(args, "body: "+requestType)
	}
	args = append(args, "options: RiidoRequestOptions = {}")
	writeJSDoc(b, operationSummary(op))
	fmt.Fprintf(b, "export async function %s(%s): Promise<%s> {\n", name, strings.Join(args, ", "), responseType)
	fmt.Fprintf(b, "  const path = %s;\n", pathTemplate(op.Path, pathParams))
	initParts := []string{fmt.Sprintf("method: '%s'", strings.ToUpper(op.Method)), "signal: options.signal"}
	if requestType != "" {
		initParts = append(initParts, "body: JSON.stringify(body)")
	}
	if eventStream {
		fmt.Fprintf(b, "  return riidoRawRequest(config, path, { %s });\n", strings.Join(initParts, ", "))
	} else {
		fmt.Fprintf(b, "  return riidoRequest<%s>(config, path, { %s });\n", responseType, strings.Join(initParts, ", "))
	}
	b.WriteString("}\n\n")
	queryKeyName := name + "QueryKey"
	writeJSDoc(b, operationSummary(op), "이 호출에 사용하는 React Query 키입니다.")
	fmt.Fprintf(b, "export function %s(%s): readonly unknown[] {\n", queryKeyName, queryKeyArgs(pathParams, paramTypeName, requestType))
	fmt.Fprintf(b, "  return [%q%s] as const;\n", name, queryKeyTail(pathParams, requestType))
	b.WriteString("}\n\n")
	if strings.EqualFold(op.Method, "GET") {
		hookName := "use" + exportedName(name)
		args := []string{"config: RiidoClientConfig"}
		callArgs := []string{"config"}
		keyArgs := []string{}
		if len(pathParams) > 0 {
			args = append(args, "params: "+paramTypeName)
			callArgs = append(callArgs, "params")
			keyArgs = append(keyArgs, "params")
		}
		callArgs = append(callArgs, "options")
		args = append(args, fmt.Sprintf("options: Omit<UseQueryOptions<%s>, 'queryKey' | 'queryFn'> & RiidoRequestOptions = {}", responseType))
		writeJSDoc(b, operationSummary(op), "React Query query hook입니다.")
		fmt.Fprintf(b, "export function %s(%s) {\n", hookName, strings.Join(args, ", "))
		fmt.Fprintf(b, "  return useQuery<%s>({\n", responseType)
		fmt.Fprintf(b, "    ...options,\n    queryKey: %s(%s),\n    queryFn: () => %s(%s),\n  });\n", queryKeyName, strings.Join(keyArgs, ", "), name, strings.Join(callArgs, ", "))
		b.WriteString("}\n\n")
		return nil
	}
	hookName := "use" + exportedName(name)
	mutationVariable := mutationVariableType(pathParams, paramTypeName, requestType)
	writeJSDoc(b, operationSummary(op), "React Query mutation hook입니다.")
	fmt.Fprintf(b, "export function %s(config: RiidoClientConfig, options: UseMutationOptions<%s, Error, %s> = {}) {\n", hookName, responseType, mutationVariable)
	fmt.Fprintf(b, "  return useMutation<%s, Error, %s>({\n", responseType, mutationVariable)
	b.WriteString("    ...options,\n")
	b.WriteString("    mutationFn: (variables) => ")
	callArgs := []string{"config"}
	if len(pathParams) > 0 {
		callArgs = append(callArgs, "variables.params")
	}
	if requestType != "" {
		callArgs = append(callArgs, "variables.body")
	}
	callArgs = append(callArgs, "{}")
	fmt.Fprintf(b, "%s(%s),\n", name, strings.Join(callArgs, ", "))
	b.WriteString("  });\n")
	b.WriteString("}\n\n")
	return nil
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

func queryKeyTail(params []string, requestType string) string {
	var parts []string
	if len(params) > 0 {
		parts = append(parts, "params")
	}
	if requestType != "" {
		parts = append(parts, "body")
	}
	if len(parts) == 0 {
		return ""
	}
	return ", " + strings.Join(parts, ", ")
}

func mutationVariableType(params []string, paramTypeName, requestType string) string {
	fields := []string{}
	if len(params) > 0 {
		fields = append(fields, "params: "+paramTypeName)
	}
	if requestType != "" {
		fields = append(fields, "body: "+requestType)
	}
	if len(fields) == 0 {
		return "void"
	}
	return "{ " + strings.Join(fields, "; ") + " }"
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
