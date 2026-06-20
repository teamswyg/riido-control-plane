package main

import "strings"

func moduleDescriptions(spec openAPISpec) map[string]string {
	out := map[string]string{}
	for _, module := range spec.ClientModules {
		out[module.Module] = module.Description
	}
	return out
}

func namespaceDescriptions(spec openAPISpec) map[string]string {
	out := map[string]string{}
	for _, module := range spec.ClientModules {
		for _, namespace := range module.Namespaces {
			out[namespaceKey(module.Module, namespace.Path)] = namespace.Description
		}
	}
	return out
}

func namespaceKey(module string, path []string) string {
	return module + "." + strings.Join(path, ".")
}

func facadePathParts(op routeOperation) []string {
	return append([]string{op.Op.Client.Module}, op.Op.Client.FacadePath...)
}

func generatedPathFromClient(client clientMetadata) string {
	if strings.TrimSpace(client.Module) == "" || len(client.FacadePath) == 0 {
		return ""
	}
	return client.Module + "." + strings.Join(client.FacadePath, ".")
}

func contractGeneratedPath(op routeOperation) string {
	if generatedPath := strings.TrimSpace(op.Op.Client.GeneratedPath); generatedPath != "" {
		return generatedPath
	}
	return generatedPathFromClient(op.Op.Client)
}

func moduleLocalGeneratedPath(op routeOperation) string {
	return strings.Join(op.Op.Client.FacadePath, ".")
}

func generatedAccessPath(op routeOperation) string {
	return "riido." + contractGeneratedPath(op)
}
