package main

import (
	"fmt"
	"strings"
)

func writeNamespaceChildProperty(b *strings.Builder, module string, childPath []string, name string, child *facadeNode, descriptions map[string]string, react bool) {
	desc := descriptions[namespaceKey(module, childPath)]
	if desc == "" {
		if child.Op != nil {
			info, _ := newOperationInfo(*child.Op)
			desc = strings.Join(operationPropertyDescriptionLines(info), " ")
		} else {
			desc = strings.Join(append([]string{module}, childPath...), ".") + " namespace입니다."
		}
	}
	writeIndentedJSDoc(b, "  ", desc)
	fmt.Fprintf(b, "  readonly %s: %s;\n", quoteProperty(name), namespaceInterfaceName(module, childPath, react))
}
