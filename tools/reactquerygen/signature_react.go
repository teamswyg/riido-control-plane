package main

import (
	"strings"
)

func reactQueryHookSignature(info operationInfo) string {
	options := "options?: core.RiidoQueryOptions<" + reactType(info.ResponseType) + ">"
	if len(info.PathParams) > 0 {
		return "params: core." + info.ParamTypeName + ", " + options
	}
	return options
}

func reactQueryHookCallArgs(info operationInfo) []string {
	var args []string
	if len(info.PathParams) > 0 {
		args = append(args, "params")
	}
	args = append(args, "options")
	return args
}

func reactType(typeName string) string {
	switch typeName {
	case "Response", "void", "unknown":
		return typeName
	default:
		return "core." + typeName
	}
}

func defaultable(signature string, withDefault bool) string {
	if withDefault {
		return signature + " = {}"
	}
	return strings.Replace(signature, "options: ", "options?: ", 1)
}
