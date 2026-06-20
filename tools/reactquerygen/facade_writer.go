package main

import (
	"fmt"
	"strings"
)

func writeFacade(b *strings.Builder, ops []routeOperation) error {
	root, err := buildFacadeTree(ops)
	if err != nil {
		return err
	}
	cacheTargets := queryOperationByCacheTag(ops)
	writeJSDoc(
		b,
		"control-plane API를 DSL client metadata의 module/namespace별로 묶은 config-bound facade입니다.",
		"React QueryClient를 대체하지 않고 request, query/queryOptions, mutation/mutationOptions와 명시적 cache helper만 제공합니다.",
	)
	b.WriteString("export function createRiidoControlPlaneClient(config: RiidoClientConfig): RiidoControlPlaneClient {\n")
	b.WriteString("  return {\n")
	writeFacadeChildren(b, root, nil, "    ", cacheTargets)
	b.WriteString("  };\n")
	b.WriteString("}\n\n")
	return nil
}

func writeFacadeChildren(b *strings.Builder, node *facadeNode, path []string, indent string, cacheTargets map[string]routeOperation) {
	for _, name := range sortedNodeNames(node) {
		child := node.Children[name]
		fmt.Fprintf(b, "%s%s: {\n", indent, quoteProperty(name))
		if child.Op != nil {
			writeFacadeOperation(b, *child.Op, append(path, name), indent+"  ", cacheTargets)
		}
		writeFacadeChildren(b, child, append(path, name), indent+"  ", cacheTargets)
		fmt.Fprintf(b, "%s},\n", indent)
	}
}
