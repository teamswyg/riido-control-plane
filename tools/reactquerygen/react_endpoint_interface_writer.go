package main

import (
	"fmt"
	"strings"
)

func writeReactEndpointInterfaces(b *strings.Builder, ops []routeOperation) error {
	for _, op := range ops {
		info, err := newOperationInfo(op)
		if err != nil {
			return err
		}
		if strings.EqualFold(op.Method, "GET") {
			writeReactQueryEndpointInterface(b, op, info)
			continue
		}
		info.MutationVariables = mutationVariableTypeName(info.Name, info.PathParams, info.RequestType)
		writeReactMutationEndpointInterface(b, op, info)
	}
	return nil
}

func writeReactQueryEndpointInterface(b *strings.Builder, op routeOperation, info operationInfo) {
	writeJSDoc(b, append(operationGeneratedPathCommentLines(op), "client의 `@/lib/react-query` 정책을 통과하는 query hook endpoint입니다.")...)
	fmt.Fprintf(b, "export interface %s extends core.%s {\n", reactEndpointInterfaceName(info.Name), endpointInterfaceName(info.Name))
	writeIndentedJSDoc(b, "  ", "React Query useQuery hook입니다.")
	fmt.Fprintf(b, "  readonly useQuery: (%s) => UseQueryResult<%s, Error>;\n", reactQueryHookSignature(info), reactType(info.ResponseType))
	b.WriteString("}\n\n")
}

func writeReactMutationEndpointInterface(b *strings.Builder, op routeOperation, info operationInfo) {
	writeJSDoc(b, append(operationGeneratedPathCommentLines(op), "client의 `@/lib/react-query` 정책을 통과하는 mutation hook endpoint입니다.")...)
	fmt.Fprintf(b, "export interface %s extends core.%s {\n", reactEndpointInterfaceName(info.Name), endpointInterfaceName(info.Name))
	writeIndentedJSDoc(b, "  ", "React Query useMutation hook입니다.")
	fmt.Fprintf(b, "  readonly useMutation: (options?: core.RiidoMutationOptions<%s, %s>) => UseMutationResult<%s, Error, %s>;\n", reactType(info.ResponseType), reactType(info.MutationVariables), reactType(info.ResponseType), reactType(info.MutationVariables))
	b.WriteString("}\n\n")
}
