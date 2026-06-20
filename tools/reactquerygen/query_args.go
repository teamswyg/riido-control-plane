package main

import "strings"

func queryRequestCallArgs(info operationInfo, signalName string) []string {
	args := []string{"config"}
	if len(info.PathParams) > 0 {
		args = append(args, "params")
	}
	if info.RequestType != "" {
		args = append(args, "body")
	}
	args = append(args, "{ "+signalName+" }")
	return args
}

func queryKeyArgs(params []string, paramTypeName, requestType string) string {
	var args []string
	if len(params) > 0 {
		args = append(args, "params: "+paramTypeName)
	}
	if requestType != "" {
		args = append(args, "body: "+requestType)
	}
	return strings.Join(args, ", ")
}

func queryKeyCallArgs(params []string, requestType string) []string {
	var args []string
	if len(params) > 0 {
		args = append(args, "params")
	}
	if requestType != "" {
		args = append(args, "body")
	}
	return args
}

func queryKeyTail(params []string, requestType string) string {
	parts := queryKeyCallArgs(params, requestType)
	if len(parts) == 0 {
		return ""
	}
	return ", " + strings.Join(parts, ", ")
}
