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

func writePathParams(b *strings.Builder, op routeOperation, info operationInfo) {
	if len(info.PathParams) == 0 {
		return
	}
	writeJSDoc(b, operationSummary(op), "경로 파라미터입니다.")
	fmt.Fprintf(b, "export interface %s {\n", info.ParamTypeName)
	for _, param := range info.PathParams {
		fmt.Fprintf(b, "  %s: string;\n", safeIdentifier(param))
	}
	b.WriteString("}\n\n")
}
