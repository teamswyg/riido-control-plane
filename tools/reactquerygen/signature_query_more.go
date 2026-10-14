package main

import (
	"fmt"
	"strings"
)

func facadeInvalidateSignature(info operationInfo) string {
	args := []string{"queryClient: QueryClient"}
	if len(info.PathParams) > 0 {
		args = append(args, "params: "+info.ParamTypeName)
	}
	return strings.Join(args, ", ")
}

func facadePrefetchSignature(info operationInfo) string {
	args := []string{"queryClient: QueryClient"}
	if len(info.PathParams) > 0 {
		args = append(args, "params: "+info.ParamTypeName)
	}
	args = append(args, fmt.Sprintf("options?: RiidoQueryOptions<%s>", info.ResponseType))
	return strings.Join(args, ", ")
}

func facadePrefetchCallArgs(info operationInfo) string {
	args := []string{"config"}
	if len(info.PathParams) > 0 {
		args = append(args, "params")
	}
	args = append(args, "options")
	return strings.Join(args, ", ")
}
