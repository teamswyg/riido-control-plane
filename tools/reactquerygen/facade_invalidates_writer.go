package main

import (
	"fmt"
	"strings"
)

func writeFacadeInvalidates(b *strings.Builder, op routeOperation, indent string, cacheTargets map[string]routeOperation) {
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
}
