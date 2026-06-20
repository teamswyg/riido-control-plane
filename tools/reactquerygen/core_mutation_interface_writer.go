package main

import (
	"fmt"
	"strings"
)

func writeCoreMutationEndpointInterfaceHeader(b *strings.Builder, info operationInfo) {
	writeJSDoc(b, append(
		operationGeneratedPathCommentLines(info.Route),
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
}
