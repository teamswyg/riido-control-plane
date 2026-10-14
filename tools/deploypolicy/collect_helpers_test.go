package deploypolicy

import "regexp"

func collectGitHubConfigRefs(body, namespace string) []string {
	re := regexp.MustCompile(`\$\{\{\s*` + regexp.QuoteMeta(namespace) + `\.([A-Z0-9_]+)\s*\}\}`)
	matches := re.FindAllStringSubmatch(body, -1)
	seen := map[string]bool{}
	var refs []string
	for _, match := range matches {
		if len(match) < 2 || seen[match[1]] {
			continue
		}
		seen[match[1]] = true
		refs = append(refs, match[1])
	}
	return refs
}

func collectRiidoAIServerKeyLiterals(body string) []string {
	re := regexp.MustCompile(`RIIDO_AI_SERVER_[A-Z0-9_]+`)
	matches := re.FindAllString(body, -1)
	seen := map[string]bool{}
	var keys []string
	for _, match := range matches {
		if seen[match] {
			continue
		}
		seen[match] = true
		keys = append(keys, match)
	}
	return keys
}

func collectNonCDRuntimeKeyNames(keys []nonCDRuntimeKey) []string {
	names := make([]string, 0, len(keys))
	for _, key := range keys {
		names = append(names, key.Name)
	}
	return names
}
