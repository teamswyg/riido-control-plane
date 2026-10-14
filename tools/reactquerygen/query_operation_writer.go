package main

import (
	"fmt"
	"strings"
)

func writeQueryOperation(b *strings.Builder, info operationInfo) {
	writeJSDoc(
		b,
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
	writeQueryOptionsOperation(b, info)
}

func writeQueryOptionsOperation(b *strings.Builder, info operationInfo) {
	writeJSDoc(b, operationSummary(info.Route), "useQuery 또는 queryClient.prefetchQuery에 전달할 수 있는 옵션입니다.")
	fmt.Fprintf(b, "export function %s(%s) {\n", queryOptionsFunctionName(info.Name), queryOptionsSignature(info, true))
	b.WriteString("  const { signal, ...queryOptions } = options;\n")
	fmt.Fprintf(
		b, "  return {\n    ...queryOptions,\n    queryKey: %s(%s),\n    queryFn: () => %s(%s),\n  };\n",
		queryKeyFunctionName(info.Name),
		strings.Join(queryKeyCallArgs(info.PathParams, info.RequestType), ", "),
		info.Name,
		strings.Join(queryRequestCallArgs(info, "signal"), ", "),
	)
	b.WriteString("}\n\n")
}
