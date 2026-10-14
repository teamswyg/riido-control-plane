package main

import "strings"

func writeReactFacade(b *strings.Builder, ops []routeOperation) error {
	root, err := buildFacadeTree(ops)
	if err != nil {
		return err
	}
	writeJSDoc(
		b,
		"control-plane API facade에 React Query hook을 얹은 client 전용 wrapper입니다.",
		"hook은 반드시 `@/lib/react-query`를 통과하므로 riido-client의 workspace/demo 정책을 우회하지 않습니다.",
	)
	b.WriteString("export function useRiidoControlPlaneClient(config: core.RiidoClientConfig): RiidoControlPlaneReactClient {\n")
	b.WriteString("  const coreClient = useMemo(() => core.createRiidoControlPlaneClient(config), [config.baseUrl, config.fetcher, config.aiAgentToken]);\n\n")
	b.WriteString("  return useMemo(\n")
	b.WriteString("    () => ({\n")
	writeReactFacadeChildren(b, root, nil, "      ")
	b.WriteString("    }),\n")
	b.WriteString("    [coreClient],\n")
	b.WriteString("  );\n")
	b.WriteString("}\n\n")
	return nil
}

func writeReactFacadeChildren(b *strings.Builder, node *facadeNode, path []string, indent string) {
	for _, name := range sortedNodeNames(node) {
		child := node.Children[name]
		nextPath := append(append([]string(nil), path...), name)
		writeReactFacadeChild(b, child, name, nextPath, indent)
	}
}
