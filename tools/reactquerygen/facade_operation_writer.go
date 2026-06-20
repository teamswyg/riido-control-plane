package main

import (
	"fmt"
	"strings"
)

func writeFacadeOperation(b *strings.Builder, op routeOperation, _ []string, indent string, cacheTargets map[string]routeOperation) {
	info, _ := newOperationInfo(op)
	if strings.EqualFold(op.Method, "GET") {
		writeFacadeQueryOperation(b, info, indent)
		return
	}
	info.MutationVariables = mutationVariableTypeName(info.Name, info.PathParams, info.RequestType)
	writeFacadeMutationOperation(b, op, info, indent, cacheTargets)
}

func writeFacadeQueryOperation(b *strings.Builder, info operationInfo, indent string) {
	fmt.Fprintf(b, "%srequest: (%s) => %s(%s),\n", indent, facadeQueryRequestSignature(info.PathParams, info.ParamTypeName), info.Name, facadeQueryRequestCallArgs(info.PathParams))
	fmt.Fprintf(b, "%squeryKeyRoot: %s,\n", indent, queryKeyRootFunctionName(info.Name))
	fmt.Fprintf(b, "%squeryKey: %s,\n", indent, queryKeyFunctionName(info.Name))
	fmt.Fprintf(b, "%squery: (%s) => %s(%s),\n", indent, facadeQueryOptionsSignature(info, true), queryOptionsFunctionName(info.Name), facadeQueryOptionsCallArgs(info))
	fmt.Fprintf(b, "%squeryOptions: (%s) => %s(%s),\n", indent, facadeQueryOptionsSignature(info, true), queryOptionsFunctionName(info.Name), facadeQueryOptionsCallArgs(info))
	fmt.Fprintf(b, "%sinvalidate: (%s) => queryClient.invalidateQueries({ queryKey: %s(%s) }),\n", indent, facadeInvalidateSignature(info), queryKeyFunctionName(info.Name), strings.Join(queryKeyCallArgs(info.PathParams, info.RequestType), ", "))
	fmt.Fprintf(b, "%sinvalidateAll: (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: %s() }),\n", indent, queryKeyRootFunctionName(info.Name))
	fmt.Fprintf(b, "%sprefetch: (%s) => queryClient.prefetchQuery(%s(%s)),\n", indent, facadePrefetchSignature(info), queryOptionsFunctionName(info.Name), facadePrefetchCallArgs(info))
}

func writeFacadeMutationOperation(b *strings.Builder, op routeOperation, info operationInfo, indent string, cacheTargets map[string]routeOperation) {
	fmt.Fprintf(b, "%srequest: (%s) => %s(%s),\n", indent, facadeMutationRequestSignature(info.PathParams, info.ParamTypeName, info.RequestType), info.Name, facadeMutationRequestCallArgs(info.PathParams, info.RequestType))
	fmt.Fprintf(b, "%smutationKey: %s,\n", indent, mutationKeyFunctionName(info.Name))
	fmt.Fprintf(b, "%smutation: (options: RiidoMutationOptions<%s, %s> = {}) => %s(config, options),\n", indent, info.ResponseType, info.MutationVariables, mutationOptionsFunctionName(info.Name))
	fmt.Fprintf(b, "%smutationOptions: (options: RiidoMutationOptions<%s, %s> = {}) => %s(config, options),\n", indent, info.ResponseType, info.MutationVariables, mutationOptionsFunctionName(info.Name))
	b.WriteString(indent + "invalidates: {\n")
	writeFacadeInvalidates(b, op, indent, cacheTargets)
	b.WriteString(indent + "},\n")
}
