package main

import "fmt"

func requireNamedRules(kind string, rules []namedRule, names []string) error {
	seen := make(map[string]struct{}, len(rules))
	for _, rule := range rules {
		if rule.Detail == "" {
			return fmt.Errorf("%s %q detail is required", kind, rule.Name)
		}
		seen[rule.Name] = struct{}{}
	}
	for _, name := range names {
		if _, ok := seen[name]; !ok {
			return fmt.Errorf("%s %q is required", kind, name)
		}
	}
	return nil
}

func requireShape(shapes []shapeRule, name string, fields []string) error {
	for _, shape := range shapes {
		if shape.Name != name {
			continue
		}
		for _, field := range fields {
			if !hasString(shape.Fields, field) {
				return fmt.Errorf("thread history v3 shape %q missing field %q", name, field)
			}
		}
		return nil
	}
	return fmt.Errorf("thread history v3 shape %q is required", name)
}

func hasString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
