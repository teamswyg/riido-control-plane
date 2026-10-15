package main

import (
	"fmt"
	"strings"
)

func writeOperation(b *strings.Builder, op routeOperation) error {
	info, err := newOperationInfo(op)
	if err != nil {
		return err
	}
	writePathParams(b, op, info)
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
	fmt.Fprintf(b, "  const path = %s;\n", pathTemplate(op.Path, pathParams(op.Path)))
	requestPath := "path"
	for _, p := range op.Op.Parameters {
		if p.In == "query" {
			requestPath = "requestPath"
			break
		}
	}
	if requestPath != "path" {
		b.WriteString("  const query = new URLSearchParams();\n")
		for _, p := range op.Op.Parameters {
			if p.In == "query" {
				fmt.Fprintf(b, "  if (params.%s !== undefined) query.set(%q, String(params.%s));\n", safeIdentifier(p.Name), p.Name, safeIdentifier(p.Name))
			}
		}
		b.WriteString("  const requestPath = query.size ? `${path}?${query.toString()}` : path;\n")
	}
	initParts := []string{fmt.Sprintf("method: '%s'", strings.ToUpper(op.Method)), "signal: options.signal"}
	if info.RequestType != "" {
		initParts = append(initParts, "body: JSON.stringify(body)")
	}
	if info.EventStream {
		fmt.Fprintf(b, "  return riidoRawRequest(config, %s, { %s });\n", requestPath, strings.Join(initParts, ", "))
	} else {
		fmt.Fprintf(b, "  return riidoRequest<%s>(config, %s, { %s });\n", info.ResponseType, requestPath, strings.Join(initParts, ", "))
	}
	b.WriteString("}\n\n")
	if strings.EqualFold(op.Method, "GET") {
		writeQueryOperation(b, info)
		return nil
	}
	writeMutationOperation(b, info)
	return nil
}

func writePathParams(b *strings.Builder, op routeOperation, info operationInfo) {
	if len(info.PathParams) == 0 {
		return
	}
	writeJSDoc(b, operationSummary(op), "경로 파라미터입니다.")
	fmt.Fprintf(b, "export interface %s {\n", info.ParamTypeName)
	for _, param := range info.PathParams {
		typ, optional := "string", ""
		for _, p := range op.Op.Parameters {
			if p.In == "query" && p.Name == param {
				if !p.Required {
					optional = "?"
				}
				if p.Schema.Type == "integer" || p.Schema.Type == "number" {
					typ = "number"
				} else if p.Schema.Type == "boolean" {
					typ = "boolean"
				}
			}
		}
		fmt.Fprintf(b, "  %s%s: %s;\n", safeIdentifier(param), optional, typ)
	}
	b.WriteString("}\n\n")
}
