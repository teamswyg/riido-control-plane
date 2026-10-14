package main

import (
	"fmt"
	"strings"
)

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
