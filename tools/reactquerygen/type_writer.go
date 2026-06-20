package main

import (
	"fmt"
	"sort"
	"strings"
)

func writeTypes(b *strings.Builder, schemas map[string]schema) {
	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		writeSchemaType(b, name, schemas[name])
	}
}

func writeSchemaType(b *strings.Builder, name string, s schema) {
	if len(s.Enum) > 0 {
		writeJSDoc(b, typeDescription(name, s))
		fmt.Fprintf(b, "export type %s = %s;\n\n", name, stringUnion(s.Enum))
		return
	}
	if len(s.OneOf) > 0 {
		parts := make([]string, 0, len(s.OneOf))
		for _, item := range s.OneOf {
			parts = append(parts, schemaType(item, true))
		}
		writeJSDoc(b, typeDescription(name, s))
		fmt.Fprintf(b, "export type %s = %s;\n\n", name, strings.Join(parts, " | "))
		return
	}
	if s.Type == "object" || len(s.Properties) > 0 {
		writeObjectSchemaType(b, name, s)
	}
}

func writeObjectSchemaType(b *strings.Builder, name string, s schema) {
	required := map[string]struct{}{}
	for _, field := range s.Required {
		required[field] = struct{}{}
	}
	writeJSDoc(b, typeDescription(name, s))
	fmt.Fprintf(b, "export interface %s {\n", name)
	fields := make([]string, 0, len(s.Properties))
	for field := range s.Properties {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	for _, field := range fields {
		_, ok := required[field]
		if description := strings.TrimSpace(s.Properties[field].Description); description != "" {
			writeIndentedJSDoc(b, "  ", description)
		}
		fmt.Fprintf(b, "  %s%s: %s;\n", quoteProperty(field), optionalMark(ok), schemaType(s.Properties[field], true))
	}
	b.WriteString("}\n\n")
}
