package main

import "strings"

func facadeMutationRequestSignature(params []string, paramTypeName, requestType string) string {
	var args []string
	if len(params) > 0 {
		args = append(args, "params: "+paramTypeName)
	}
	if requestType != "" {
		args = append(args, "body: "+requestType)
	}
	args = append(args, "options?: RiidoRequestOptions")
	return strings.Join(args, ", ")
}

func facadeMutationRequestCallArgs(params []string, requestType string) string {
	args := []string{"config"}
	if len(params) > 0 {
		args = append(args, "params")
	}
	if requestType != "" {
		args = append(args, "body")
	}
	args = append(args, "options")
	return strings.Join(args, ", ")
}

func invalidationPropertyName(cacheTag, module string) string {
	prefix := module + "."
	trimmed := strings.TrimPrefix(cacheTag, prefix)
	parts := strings.Split(trimmed, ".")
	for i, part := range parts {
		if i == 0 {
			parts[i] = part
			continue
		}
		parts[i] = exportedName(part)
	}
	return safeIdentifier(strings.Join(parts, ""))
}
