package main

import (
	"fmt"
	"strings"
)

func queryOptionsSignature(info operationInfo, withDefault bool) string {
	args := []string{"config: RiidoClientConfig"}
	if len(info.PathParams) > 0 {
		args = append(args, "params: "+info.ParamTypeName)
	}
	args = append(args, defaultable(fmt.Sprintf("options: RiidoQueryOptions<%s>", info.ResponseType), withDefault))
	return strings.Join(args, ", ")
}

func facadeQueryRequestSignature(params []string, paramTypeName string) string {
	if len(params) > 0 {
		return fmt.Sprintf("params: %s, options?: RiidoRequestOptions", paramTypeName)
	}
	return "options?: RiidoRequestOptions"
}

func facadeQueryRequestCallArgs(params []string) string {
	if len(params) > 0 {
		return "config, params, options"
	}
	return "config, options"
}

func facadeQueryOptionsSignature(info operationInfo, withDefault bool) string {
	options := defaultable(fmt.Sprintf("options: RiidoQueryOptions<%s>", info.ResponseType), withDefault)
	if len(info.PathParams) > 0 {
		return fmt.Sprintf("params: %s, %s", info.ParamTypeName, options)
	}
	return options
}

func facadeQueryOptionsCallArgs(info operationInfo) string {
	args := []string{"config"}
	if len(info.PathParams) > 0 {
		args = append(args, "params")
	}
	args = append(args, "options")
	return strings.Join(args, ", ")
}
