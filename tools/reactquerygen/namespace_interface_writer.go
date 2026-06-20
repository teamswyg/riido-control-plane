package main

import (
	"fmt"
	"strings"
)

func writeNamespaceInterface(b *strings.Builder, module string, path []string, node *facadeNode, descriptions map[string]string, rootDescription string, react bool) {
	for _, name := range sortedNodeNames(node) {
		child := node.Children[name]
		if len(child.Children) > 0 {
			writeNamespaceInterface(b, module, append(path, name), child, descriptions, rootDescription, react)
		}
	}
	description := namespaceDescription(module, path, descriptions, rootDescription)
	writeJSDoc(b, description)
	interfaceName := namespaceInterfaceName(module, path, react)
	if node.Op != nil {
		info, _ := newOperationInfo(*node.Op)
		fmt.Fprintf(b, "export interface %s extends %s {\n", interfaceName, operationEndpointType(info.Name, react))
	} else {
		fmt.Fprintf(b, "export interface %s {\n", interfaceName)
	}
	writeNamespaceProperties(b, module, path, node, descriptions, react)
	b.WriteString("}\n\n")
}

func namespaceDescription(module string, path []string, descriptions map[string]string, rootDescription string) string {
	description := rootDescription
	if len(path) > 0 {
		description = descriptions[namespaceKey(module, path)]
	}
	if description == "" {
		description = strings.Join(append([]string{module}, path...), ".") + " namespace입니다."
	}
	return description
}

func writeNamespaceProperties(b *strings.Builder, module string, path []string, node *facadeNode, descriptions map[string]string, react bool) {
	for _, name := range sortedNodeNames(node) {
		child := node.Children[name]
		if child.Op != nil && len(child.Children) == 0 {
			info, _ := newOperationInfo(*child.Op)
			writeIndentedJSDoc(b, "  ", operationPropertyDescriptionLines(info)...)
			fmt.Fprintf(b, "  readonly %s: %s;\n", quoteProperty(name), operationEndpointType(info.Name, react))
			continue
		}
		childPath := append(append([]string(nil), path...), name)
		writeNamespaceChildProperty(b, module, childPath, name, child, descriptions, react)
	}
}
