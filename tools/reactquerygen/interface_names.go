package main

import "strings"

func operationEndpointType(operationID string, react bool) string {
	if react {
		return reactEndpointInterfaceName(operationID)
	}
	return endpointInterfaceName(operationID)
}

func endpointInterfaceName(operationID string) string {
	return exportedName(operationID) + "Endpoint"
}

func reactEndpointInterfaceName(operationID string) string {
	return exportedName(operationID) + "ReactEndpoint"
}

func namespaceInterfaceName(module string, path []string, react bool) string {
	parts := []string{"Riido", facadeTypeSegment(module)}
	for _, part := range path {
		parts = append(parts, facadeTypeSegment(part))
	}
	if react {
		parts = append(parts, "React")
	}
	if len(path) == 0 {
		parts = append(parts, "Module")
	} else {
		parts = append(parts, "Namespace")
	}
	return strings.Join(parts, "")
}

func facadeTypeSegment(segment string) string {
	switch segment {
	case "aiAgent":
		return "AIAgent"
	default:
		return exportedName(segment)
	}
}
