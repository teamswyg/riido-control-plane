package main

import (
	"fmt"
	"strings"
)

func writeCoreQueryEndpointInterface(b *strings.Builder, info operationInfo) {
	writeJSDoc(b, append(
		operationGeneratedPathCommentLines(info.Route),
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
	writeCoreMutationEndpointInterfaceHeader(b, info)
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
