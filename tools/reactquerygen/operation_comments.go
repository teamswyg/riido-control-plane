package main

import (
	"fmt"
	"strings"
)

func operationGeneratedPathCommentLines(op routeOperation) []string {
	return []string{
		operationSummary(op),
		fmt.Sprintf("계약 generated path: `%s`", contractGeneratedPath(op)),
		fmt.Sprintf("검색용 generated 경로: `%s`", moduleLocalGeneratedPath(op)),
		fmt.Sprintf("접근 예시: `%s`", generatedAccessPath(op)),
	}
}

func operationPropertyDescriptionLines(info operationInfo) []string {
	lines := operationGeneratedPathCommentLines(info.Route)
	if strings.EqualFold(info.Route.Method, "GET") {
		lines = append(lines, fmt.Sprintf("cache tag: `%s`", info.Route.Op.Client.CacheTag))
	} else if len(info.Route.Op.Client.Invalidates) > 0 {
		lines = append(lines, "invalidates: `"+strings.Join(info.Route.Op.Client.Invalidates, "`, `")+"`")
	}
	return lines
}

func operationSummary(op routeOperation) string {
	if summary := strings.TrimSpace(op.Op.Summary); summary != "" {
		return summary
	}
	return fmt.Sprintf("%s %s", strings.ToUpper(op.Method), op.Path)
}
