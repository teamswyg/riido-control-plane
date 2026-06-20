package main

import "strings"

func schemaType(s schema, topLevel bool) string {
	if s.Ref != "" {
		return refName(s.Ref)
	}
	if len(s.Enum) > 0 {
		return stringUnion(s.Enum)
	}
	if len(s.OneOf) > 0 {
		parts := make([]string, 0, len(s.OneOf))
		for _, item := range s.OneOf {
			parts = append(parts, schemaType(item, true))
		}
		return strings.Join(parts, " | ")
	}
	switch s.Type {
	case "array":
		if s.Items == nil {
			return "unknown[]"
		}
		return schemaType(*s.Items, false) + "[]"
	case "boolean":
		return "boolean"
	case "integer", "number":
		return "number"
	case "object":
		if valueType, ok := additionalPropertiesType(s.AdditionalProperties); ok {
			return "Record<string, " + valueType + ">"
		}
		if topLevel {
			return "Record<string, unknown>"
		}
		return "unknown"
	case "string":
		return "string"
	default:
		return "unknown"
	}
}

func refName(ref string) string {
	const prefix = "#/components/schemas/"
	return strings.TrimPrefix(ref, prefix)
}
