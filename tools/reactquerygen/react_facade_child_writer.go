package main

import (
	"fmt"
	"strings"
)

func writeReactFacadeChild(b *strings.Builder, child *facadeNode, name string, path []string, indent string) {
	fmt.Fprintf(b, "%s%s: {\n", indent, quoteProperty(name))
	if child.Op != nil {
		writeReactFacadeOperation(b, *child.Op, path, indent+"  ")
	}
	writeReactFacadeChildren(b, child, path, indent+"  ")
	fmt.Fprintf(b, "%s},\n", indent)
}

func writeReactFacadeOperation(b *strings.Builder, op routeOperation, path []string, indent string) {
	info, _ := newOperationInfo(op)
	accessor := "coreClient." + strings.Join(path, ".")
	fmt.Fprintf(b, "%s...%s,\n", indent, accessor)
	if strings.EqualFold(op.Method, "GET") {
		fmt.Fprintf(b, "%suseQuery: (%s) => useQuery<%s, Error>(%s.query(%s)),\n",
			indent, reactQueryHookSignature(info), reactType(info.ResponseType), accessor, strings.Join(reactQueryHookCallArgs(info), ", "))
		return
	}
	info.MutationVariables = mutationVariableTypeName(info.Name, info.PathParams, info.RequestType)
	fmt.Fprintf(
		b, "%suseMutation: (options: core.RiidoMutationOptions<%s, %s> = {}) => useMutation<%s, Error, %s>(%s.mutation(options)),\n",
		indent, reactType(info.ResponseType), reactType(info.MutationVariables), reactType(info.ResponseType), reactType(info.MutationVariables), accessor,
	)
}
