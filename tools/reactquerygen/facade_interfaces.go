package main

import (
	"fmt"
	"strings"
)

func writeFacadeInterfaces(b *strings.Builder, spec openAPISpec, ops []routeOperation, react bool) error {
	root, err := buildFacadeTree(ops)
	if err != nil {
		return err
	}
	if react {
		if err := writeReactEndpointInterfaces(b, ops); err != nil {
			return err
		}
	} else if err := writeCoreEndpointInterfaces(b, ops); err != nil {
		return err
	}
	descriptions := namespaceDescriptions(spec)
	moduleDescriptions := moduleDescriptions(spec)
	for _, moduleName := range sortedNodeNames(root) {
		writeNamespaceInterface(b, moduleName, nil, root.Children[moduleName], descriptions, moduleDescriptions[moduleName], react)
	}
	writeControlPlaneClientInterface(b, root, moduleDescriptions, react)
	return nil
}

func writeCoreEndpointInterfaces(b *strings.Builder, ops []routeOperation) error {
	cacheTargets := queryOperationByCacheTag(ops)
	for _, op := range ops {
		info, err := newOperationInfo(op)
		if err != nil {
			return err
		}
		if strings.EqualFold(op.Method, "GET") {
			writeCoreQueryEndpointInterface(b, info)
			continue
		}
		info.MutationVariables = mutationVariableTypeName(info.Name, info.PathParams, info.RequestType)
		writeCoreMutationEndpointInterface(b, info, cacheTargets)
	}
	return nil
}

func writeControlPlaneClientInterface(b *strings.Builder, root *facadeNode, moduleDescriptions map[string]string, react bool) {
	name := "RiidoControlPlaneClient"
	if react {
		name = "RiidoControlPlaneReactClient"
	}
	writeJSDoc(b, "control-plane API를 DSL client metadata의 module/namespace별로 묶은 config-bound facade입니다.")
	fmt.Fprintf(b, "export interface %s {\n", name)
	for _, module := range sortedNodeNames(root) {
		description := moduleDescriptions[module]
		if description == "" {
			description = module + " module입니다."
		}
		writeIndentedJSDoc(b, "  ", description)
		fmt.Fprintf(b, "  readonly %s: %s;\n", quoteProperty(module), namespaceInterfaceName(module, nil, react))
	}
	b.WriteString("}\n\n")
}
